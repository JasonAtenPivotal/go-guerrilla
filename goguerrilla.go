/**
Go-Guerrilla SMTPd
A minimalist SMTP server written in Go, made for receiving large volumes of mail.
Works either as a stand-alone or in conjunction with Nginx SMTP proxy.
TO DO: add http server for nginx

Copyright (c) 2012 Flashmob, GuerrillaMail.com

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

What is Go Guerrilla SMTPd?
It's a small SMTP server written in Go, optimized for receiving email.
Written for GuerrillaMail.com which processes tens of thousands of emails
every hour.

Benchmarking:
http://www.jrh.org/smtp/index.html
Test 500 clients:
$ time smtp-source -c -l 5000 -t test@spam4.me -s 500 -m 5000 5.9.7.183

Version: 1.1
Author: Flashmob, GuerrillaMail.com
Contact: flashmob@gmail.com
License: MIT
Repository: https://github.com/flashmob/Go-Guerrilla-SMTPd
Site: http://www.guerrillamail.com/

See README for more details

TODO: after failing tls,

*/

package main

import (
	"bufio"

	"crypto/rand"
	"crypto/tls"

	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Client struct {
	state        int
	helo         string
	mail_from    string
	rcpt_to      string
	read_buffer  string
	response     string
	address      string
	data         string
	subject      string
	hash         string
	time         int64
	tls_on       bool
	conn         net.Conn
	bufin        *bufio.Reader
	bufout       *bufio.Writer
	kill_time    int64
	errors       int
	clientId     int64
	savedNotify  chan int
	notifyOnDone chan *Client
}

type Smtpd struct {
	Addr             string
	ExternalMailchan chan *Client // notify external clients (like tests) that mail arrived.
	RequestStop      chan bool    // ask the Stmtp to shutdown by sending on this channel.
	Done             chan bool    // Stmtpd will close this as its last action, so you can wait for it to complete.

	TLSconfig *tls.Config
	max_size  int // max email DATA size
	timeout   time.Duration
	sem       chan int // currently active clients

	Savers       []*Saver
	SaveMailChan chan *Client // Savers all read from this

	// avoid startup races with our test client, by
	// closing PortBound when the server has its endpoint bound.
	PortBound chan bool
	Cfg       Config
	Clients   *ClientTracker
}

func NewGoGuerrillaSmtpd(addr string, mailchan chan *Client) *Smtpd {
	cfg := NewConfig()
	return &Smtpd{
		Addr:             addr,
		ExternalMailchan: mailchan,
		RequestStop:      make(chan bool),
		Done:             make(chan bool),
		PortBound:        make(chan bool),
		SaveMailChan:     make(chan *Client, 5),
		Savers:           make([]*Saver, 0),
		Cfg:              *cfg,
		Clients:          NewClientTracker(cfg),
	}
}

type Config struct {
	Verbose      bool
	UseLogFile   bool
	Map          map[string]string
	AllowedHosts map[string]bool
}

func NewConfig() *Config {
	return &Config{
		Map:          make(map[string]string),
		AllowedHosts: make(map[string]bool, 15),
	}
}

// defaults. configure() loads settings them from a json file,
//  fills in these if not set in the file.
var gConfig = map[string]string{
	"GSMTP_MAX_SIZE":         "131072",
	"GSMTP_HOST_NAME":        "server.example.com", // This should also be set to reflect your RDNS
	"GSMTP_LOG_FILE":         "",                   // Eg. /var/log/goguerrilla.log or leave blank if no logging
	"GSMTP_TIMEOUT":          "100",                // how many seconds before timeout.
	"GSTMP_LISTEN_INTERFACE": "0.0.0.0:2525",
	"GSMTP_PUB_KEY":          "/etc/ssl/certs/ssl-cert-snakeoil.pem",
	"GSMTP_PRV_KEY":          "/etc/ssl/private/ssl-cert-snakeoil.key",
	"GSMTP_VERBOSE":          "Y",

	"GM_MAIL_TABLE":        "mail_queue",
	"GM_ALLOWED_HOSTS":     "guerrillamail.de,guerrillamailblock.com",
	"GM_PRIMARY_MAIL_HOST": "guerrillamail.com",
	"GM_MAX_CLIENTS":       "500",

	"MYSQL_HOST": "127.0.0.1:3306",
	"MYSQL_USER": "go_guerrilla",
	"MYSQL_PASS": "ok",
	"MYSQL_DB":   "go_guerrilla",

	"NGINX_AUTH_ENABLED": "N",              // Y or N
	"NGINX_AUTH":         "127.0.0.1:8025", // If using Nginx proxy, ip and port to serve Auth requsts
	"SGID":               "1008",           // group id
	"SUID":               "1008",           // user id, from /etc/passwd
}

func (cfg *Config) logln(level int, s string) {
	if cfg.Verbose {
		fmt.Println(s)
	}
	if level == 2 {
		log.Fatalf(s)
	}
	if cfg.UseLogFile {
		log.Println(s)
	}
}

func (s *Smtpd) configure() {
	var configFile, iface string

	log.SetOutput(os.Stdout)
	// parse command line arguments
	var pver = flag.Bool("v", false, "be verbose")
	flag.StringVar(&configFile, "config", "goguerrilla.conf", "Path to the configuration file")
	flag.StringVar(&iface, "if", "", "Interface and port to listen on, eg. 127.0.0.1:2525 ")
	flag.Parse()

	// load in the config.
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln("Could not read config file '%s'", configFile)
	}
	var jsonMap map[string]string
	err = json.Unmarshal(b, &jsonMap)
	if err != nil {
		log.Fatalln("Could not parse config file: %s", err)
	}

	// fill in the defaults from gConfig
	for k, v := range gConfig {
		if _, already := jsonMap[k]; !already {
			jsonMap[k] = v
		}
	}
	//fmt.Printf("jsonMap = %#v\n", jsonMap)

	// let command line override
	if pver != nil && *pver {
		s.Cfg.Verbose = true
	}
	if jsonMap["GSMTP_VERBOSE"] == "Y" {
		s.Cfg.Verbose = true
	}
	if len(iface) > 0 {
		jsonMap["GSTMP_LISTEN_INTERFACE"] = iface
	}

	// map the allow hosts for easy lookup
	if arr := strings.Split(jsonMap["GM_ALLOWED_HOSTS"], ","); len(arr) > 0 {
		s.Cfg.AllowedHosts = make(map[string]bool)
		for i := 0; i < len(arr); i++ {
			s.Cfg.AllowedHosts[arr[i]] = true
		}
	}

	//fmt.Printf("configFile = %#v\n", configFile)
	//fmt.Printf("gConfig = %#v\n", gConfig)

	var n int
	var n_err error
	// sem is an active clients channel used for counting clients
	if n, n_err = strconv.Atoi(jsonMap["GM_MAX_CLIENTS"]); n_err != nil {
		n = 50
	}
	// currently active client list
	s.sem = make(chan int, n)
	// database writing workers
	// timeout for reads
	if n, n_err = strconv.Atoi(jsonMap["GSMTP_TIMEOUT"]); n_err != nil {
		s.timeout = time.Duration(10)
	} else {
		s.timeout = time.Duration(n)
	}

	s.Cfg.logln(1, fmt.Sprintf("[configure] Timeout set to %d", s.timeout))
	// max email size
	if s.max_size, n_err = strconv.Atoi(jsonMap["GSMTP_MAX_SIZE"]); n_err != nil {
		s.max_size = 131072
	}
	// custom log file
	if len(jsonMap["GSMTP_LOG_FILE"]) > 0 {
		logfile, err := os.OpenFile(jsonMap["GSMTP_LOG_FILE"], os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0600)
		if err != nil {
			log.Fatal("Unable to open log file ["+jsonMap["GSMTP_LOG_FILE"]+"]: ", err)
		}
		log.SetOutput(logfile)
		s.Cfg.UseLogFile = true
	}

	s.Cfg.Map = jsonMap
}

func main() {
	mailchan := make(chan *Client)
	addr := "0.0.0.0:2525"
	server := NewGoGuerrillaSmtpd(addr, mailchan)
	server.Start()

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	// Stop the service gracefully.
	server.Shutdown()
	log.Println("shutdown complete.")
}

func (s *Smtpd) Shutdown() {
	s.RequestStop <- true

	// don't return until server is all down.
	<-s.Done
}

// track open client connections,
// so we can shutdown cleanly
type ClientTracker struct {
	Clients map[*Client]bool

	ClientDoneChannel  chan *Client
	ClientAddedChannel chan *Client

	// lifecycle
	RequestStop chan bool
	Done        chan bool

	ShutdownRequested bool
	Cfg               Config
}

func NewClientTracker(cfg *Config) *ClientTracker {
	return &ClientTracker{
		Clients:            make(map[*Client]bool),
		ClientDoneChannel:  make(chan *Client),
		ClientAddedChannel: make(chan *Client),
		RequestStop:        make(chan bool),
		Done:               make(chan bool),
		Cfg:                *cfg,
	}
}

func (s *ClientTracker) Start() {

	go func() {
		for {
			select {
			case cl := <-s.ClientAddedChannel:
				s.Clients[cl] = true

			case cl := <-s.ClientDoneChannel:
				delete(s.Clients, cl)
				if len(s.Clients) == 0 && s.ShutdownRequested {
					s.Cfg.logln(1, "ClientTracker has no more clients: exiting.\n")
					close(s.Done)
					return
				}
			case <-s.RequestStop:
				s.ShutdownRequested = true
				s.Cfg.logln(1, "ClientTracker sees RequestStop, flagging in all clients to exit.\n")
				if len(s.Clients) == 0 {
					s.Cfg.logln(1, "ClientTracker has no more clients: exiting.\n")
					close(s.Done)
					return
				}
			}
		}
	}()
}

func isTimeout(err error) bool {
	e, ok := err.(net.Error)
	return ok && e.Timeout()
}

func (s *Smtpd) Start() {
	s.Clients.Start()
	go func() {
		s.configure()
		cert, err := tls.LoadX509KeyPair(s.Cfg.Map["GSMTP_PUB_KEY"], s.Cfg.Map["GSMTP_PRV_KEY"])
		if err != nil {
			s.Cfg.logln(2, fmt.Sprintf("There was a problem loading the certificate: %s", err))
		}

		s.TLSconfig = &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.VerifyClientCertIfGiven, ServerName: s.Cfg.Map["GSMTP_HOST_NAME"]}
		s.TLSconfig.Rand = rand.Reader
		// start some savemail workers
		for i := 0; i < 3; i++ {
			s.Cfg.logln(1, fmt.Sprintf("Starting mail processing worker %d", i))
			ns := NewSaver(s.SaveMailChan, s.Cfg, s.ExternalMailchan)
			s.Savers = append(s.Savers, ns)
			ns.start()
		}
		if s.Cfg.Map["NGINX_AUTH_ENABLED"] == "Y" {
			go s.nginxHTTPAuth()
		}

		// Start listening for SMTP connections

		laddr, err := net.ResolveTCPAddr("tcp", s.Cfg.Map["GSTMP_LISTEN_INTERFACE"])
		if nil != err {
			log.Fatalln(err)
		}

		listener, err := net.ListenTCP("tcp", laddr)
		if err != nil {
			s.Cfg.logln(2, fmt.Sprintf("Cannot listen on port, %v", err))
		} else {
			s.Cfg.logln(1, fmt.Sprintf("Listening on tcp %s", s.Cfg.Map["GSTMP_LISTEN_INTERFACE"]))
		}
		close(s.PortBound)

		var clientId int64
		clientId = 1
		for {
			select {
			case <-s.RequestStop:
				s.Cfg.logln(1, "Smtpd: stop requested, initiating shutdown...\n")
				// stop clients
				s.Clients.RequestStop <- true
				<-s.Clients.Done

				s.Cfg.logln(1, "Smtpd: shutdown progress: clients done...\n")

				// stop savers
				for _, saver := range s.Savers {
					saver.RequestStop <- true
					<-saver.Done
				}
				s.Cfg.logln(1, "Smtpd: shutdown progress: savers done...\n")

				// we are done
				close(s.Done)
				return
			default:
				//s.Cfg.logln(1, "Smtpd::Start(), about to block on Accept for 1000 msec.\n")
				listener.SetDeadline(time.Now().Add(1000 * time.Millisecond))
				conn, err := listener.Accept()
				if isTimeout(err) {
					continue
				}
				if err != nil {
					s.Cfg.logln(1, fmt.Sprintf("Accept error: %s", err))
					continue
				}
				s.Cfg.logln(1, fmt.Sprintf(" There are now "+strconv.Itoa(runtime.NumGoroutine())+" serving goroutines"))
				s.Cfg.logln(1, fmt.Sprintf("Connecting client %d. RemoteAddr: %s", clientId, conn.RemoteAddr().String()))
				s.sem <- 1 // Wait for active queue to drain.

				cl := &Client{
					conn:         conn,
					address:      conn.RemoteAddr().String(),
					time:         time.Now().Unix(),
					bufin:        bufio.NewReader(conn),
					bufout:       bufio.NewWriter(conn),
					clientId:     clientId,
					savedNotify:  make(chan int),
					notifyOnDone: s.Clients.ClientDoneChannel,
				}
				s.Clients.ClientAddedChannel <- cl
				go s.handleClient(cl)

				clientId++
			}
		}
	}()

	// Wait until port is bound, to avoid races with clients.
	// Hence clients can start hitting the server immediately after Start() returns.
	<-s.PortBound
}

func (s *Smtpd) handleClient(client *Client) {
	defer func() {
		s.closeClient(client)
		client.notifyOnDone <- client
	}()

	greeting := "220 " + s.Cfg.Map["GSMTP_HOST_NAME"] +
		" SMTP Guerrilla-SMTPd #" + strconv.FormatInt(client.clientId, 10) + " (" + strconv.Itoa(len(s.sem)) + ") " + time.Now().Format(time.RFC1123Z)
	// advertiseTls := "250-STARTTLS\r\n"
	advertiseAuth := "250-AUTH PLAIN\r\n"
	advertiseTls := ""
	for i := 0; i < 100; i++ {
		switch client.state {
		case 0:
			s.Cfg.logln(1, fmt.Sprintf("clientId %d is in state 0", client.clientId))
			s.responseAdd(client, greeting)
			client.state = 1
		case 1:
			s.Cfg.logln(1, fmt.Sprintf("clientId %d is in state 1", client.clientId))
			input, err := s.readSmtp(client)
			if err != nil {
				s.Cfg.logln(1, fmt.Sprintf("[handleClient] Read error: %v", err))
				if err == io.EOF {
					s.Cfg.logln(1, fmt.Sprintf("[handleClient] client closed the connection? Client: %v", client))
					// client closed the connection already
					return
				}
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					s.Cfg.logln(1, fmt.Sprintf("[handleClient] client is too slow? %v", neterr))
					// too slow, timeout
					return
				}
				break
			}
			input = strings.Trim(input, " \n\r")
			cmd := strings.ToUpper(input)
			switch {
			case strings.Index(cmd, "HELO") == 0:
				if len(input) > 5 {
					client.helo = input[5:]
				}
				s.responseAdd(client, "250 "+s.Cfg.Map["GSMTP_HOST_NAME"]+" Hello ")
			case strings.Index(cmd, "EHLO") == 0:
				if len(input) > 5 {
					client.helo = input[5:]
				}
				s.responseAdd(client, "250-"+s.Cfg.Map["GSMTP_HOST_NAME"]+" Hello "+client.helo+"["+client.address+"]"+"\r\n"+"250-SIZE "+s.Cfg.Map["GSMTP_MAX_SIZE"]+"\r\n"+advertiseTls+advertiseAuth+"250 HELP")
				// s.responseAdd(client, "250-"+s.Cfg.Map["GSMTP_HOST_NAME"]+" Hello "+client.helo+"["+client.address+"]"+"\r\n"+"250-SIZE "+s.Cfg.Map["GSMTP_MAX_SIZE"]+"\r\n"+advertiseTls+"")
			case strings.Index(cmd, "MAIL FROM:") == 0:
				if len(input) > 10 {
					client.mail_from = input[10:]
				}
				s.responseAdd(client, "250 Ok")
			case strings.Index(cmd, "XCLIENT") == 0:
				// Nginx sends this
				// XCLIENT ADDR=212.96.64.216 NAME=[UNAVAILABLE]
				client.address = input[13:]
				client.address = client.address[0:strings.Index(client.address, " ")]
				fmt.Println("client address:[" + client.address + "]")
				s.responseAdd(client, "250 OK")
			case strings.Index(cmd, "RCPT TO:") == 0:
				if len(input) > 8 {
					client.rcpt_to = input[8:]
				}
				s.responseAdd(client, "250 Accepted")
			case strings.Index(cmd, "NOOP") == 0:
				s.responseAdd(client, "250 OK")
			case strings.Index(cmd, "AUTH") == 0:
				s.responseAdd(client, "235 2.7.0 Authentication successful")
			case strings.Index(cmd, "RSET") == 0:
				client.mail_from = ""
				client.rcpt_to = ""
				s.responseAdd(client, "250 OK")
			case strings.Index(cmd, "DATA") == 0:
				s.responseAdd(client, "354 Enter message, ending with \".\" on a line by itself")
				client.state = 2
			case (strings.Index(cmd, "STARTTLS") == 0) && !client.tls_on:
				s.responseAdd(client, "220 Ready to start TLS")
				// go to start TLS state
				client.state = 3
			case strings.Index(cmd, "QUIT") == 0:
				s.responseAdd(client, "221 Bye")
				killClient(client)
			default:
				s.responseAdd(client, fmt.Sprintf("500 unrecognized command"))
				client.errors++
				if client.errors > 3 {
					s.responseAdd(client, fmt.Sprintf("500 Too many unrecognized commands"))
					killClient(client)
				}
			}
		case 2:
			s.Cfg.logln(1, "client.state 2")
			var err error
			client.data, err = s.readSmtp(client)
			if err == nil {
				// to do: timeout when adding to SaveMailChan
				// place on the channel so that one of the save mail workers can pick it up
				s.SaveMailChan <- client
				// wait for the save to complete
				status := <-client.savedNotify

				if status == 1 {
					s.responseAdd(client, "250 OK : queued as "+client.hash)
				} else {
					s.responseAdd(client, "554 Error: transaction failed, you may not be on the allowed hosts list.")
				}
			} else {
				s.Cfg.logln(1, fmt.Sprintf("DATA read error: %v", err))
			}
			client.state = 1
		case 3:
			s.Cfg.logln(1, "client.state 3")
			// upgrade to TLS
			var tlsConn *tls.Conn
			tlsConn = tls.Server(client.conn, s.TLSconfig)
			err := tlsConn.Handshake() // not necessary to call here, but might as well
			if err == nil {
				client.conn = net.Conn(tlsConn)
				client.bufin = bufio.NewReader(client.conn)
				client.bufout = bufio.NewWriter(client.conn)
				client.tls_on = true
			} else {
				s.Cfg.logln(1, fmt.Sprintf("Could not TLS handshake:%v", err))
			}
			advertiseTls = ""
			client.state = 1
		}
		// Send a response back to the client
		err := s.responseWrite(client)
		if err != nil {
			if err == io.EOF {
				// client closed the connection already
				return
			}
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				// too slow, timeout
				return
			}
		}
		if client.kill_time > 1 {
			return
		}
	}

}

func (s *Smtpd) responseAdd(client *Client, line string) {
	s.Cfg.logln(1, fmt.Sprintf("[responseAdd] > \"%s\"", line))
	client.response = line + "\r\n"
}
func (s *Smtpd) closeClient(client *Client) {
	client.conn.Close()
	<-s.sem // Done; enable next client to run.
}
func killClient(client *Client) {
	client.kill_time = time.Now().Unix()
}

func (s *Smtpd) readSmtp(client *Client) (input string, err error) {
	var reply string
	// Command state terminator by default
	suffix := "\r\n"
	if client.state == 2 {
		// DATA state
		suffix = "\r\n.\r\n"
	}
	for err == nil {
		client.conn.SetDeadline(time.Now().Add(s.timeout * time.Second))
		reply, err = client.bufin.ReadString('\n')

		if reply != "" {
			input = input + reply
			if len(input) > s.max_size {
				err = errors.New("Maximum DATA size exceeded (" + strconv.Itoa(s.max_size) + ")")
				return input, err
			}
			if client.state == 2 {
				// Extract the subject while we are at it.
				scanSubject(client, reply)
			}
		}
		if err != nil {
			s.Cfg.logln(1, fmt.Sprintf("[readSmtp] Error: \"%v\"; read from client: \"%v\"", err, reply))
			break
		}
		if strings.HasSuffix(input, suffix) {
			break
		}
	}
	s.Cfg.logln(1, fmt.Sprintf("[readSmtp]   < \"%v\", err: \"%v\"", input, err))
	return input, err
}

func (s *Smtpd) responseWrite(client *Client) (err error) {
	var size int
	client.conn.SetDeadline(time.Now().Add(s.timeout * time.Second))
	size, err = client.bufout.WriteString(client.response)
	client.bufout.Flush()
	client.response = client.response[size:]
	return err
}

type Saver struct {
	SaveMailChan    chan *Client
	RequestStop     chan bool // ask for shutdown by sending on this
	Done            chan bool // close as last action
	Cfg             Config
	NotifyAfterSave chan *Client
}

func NewSaver(saveMailChan chan *Client, cfg Config, notify chan *Client) *Saver {
	return &Saver{
		SaveMailChan:    saveMailChan,
		RequestStop:     make(chan bool),
		Done:            make(chan bool),
		Cfg:             cfg,
		NotifyAfterSave: notify,
	}
}

func (s *Saver) start() {
	go func() {
		var to string
		//var err error
		//var body string
		//var length int

		//  receives values from the channel repeatedly until it is closed.
		s.Cfg.logln(1, "saveMail entering loop")
		for {
			select {
			case <-s.RequestStop:
				s.Cfg.logln(1, "saver sees RequestStop, exiting.\n")
				close(s.Done)
				return
			case client := <-s.SaveMailChan:
				s.Cfg.logln(1, fmt.Sprintf("saveMail processing client %#v", client))
				if user, _, addr_err := s.Cfg.validateEmailData(client); addr_err != nil { // user, host, addr_err
					s.Cfg.logln(1, fmt.Sprintln("mail_from did not validate: %v", addr_err)+" client.mail_from:"+client.mail_from)
					// notify client that a save completed, -1 = error
					client.savedNotify <- -1
					continue
				} else {
					to = user + "@" + s.Cfg.Map["GM_PRIMARY_MAIL_HOST"]
				}
				//length = len(client.data)
				fmt.Printf("debug: client.subject is: '%s'\n", client.subject)		
				client.subject = mimeHeaderDecode(client.subject)

				if s.Cfg.Verbose {
					fmt.Printf("debug: client.data is: '%s'\n", client.data)
					fmt.Printf("debug: client.subject is: '%s'\n", client.subject)
				}

				client.hash = md5hex(to + client.mail_from + client.subject + strconv.FormatInt(time.Now().UnixNano(), 10))
				// Add extra headers
				add_head := ""
				add_head += "Delivered-To: " + to + "\r\n"
				add_head += "Received: from " + client.helo + " (" + client.helo + "  [" + client.address + "])\r\n"
				add_head += "	by " + s.Cfg.Map["GSMTP_HOST_NAME"] + " with SMTP id " + client.hash + "@" + s.Cfg.Map["GSMTP_HOST_NAME"] + ";\r\n"
				add_head += "	" + time.Now().Format(time.RFC1123Z) + "\r\n"
				// compress to save space
				client.data = compress(add_head + client.data)
				//body = "gzencode"

				client.savedNotify <- 1

				if s.NotifyAfterSave != nil {
					s.NotifyAfterSave <- client
				}
			}
		}
	}()
}
