package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/smtp"
	"time"

	"github.com/glycerine/gophermail"
)

func main() {

	body := "test-mail-body-here"

	sendGopherMail(fmt.Sprintf("gomailclient send time: %v\n", time.Now().Local().String())+body, "started")
	log.Printf("gomailclient done.")
}

func simpleSendEmailExampleUsingWorkingSendMail() {
	/*
	   	goodmsg := `Subject: gomailclient test-from-local
	   To: me <me@p-i-v-o-t-a.l-labs.com>
	   X-Mailer: gomailclient.go
	   From: me@p-i-v-o-t-a.l-labs.com (me)

	   I'm the text body
	   `

	   	WorkingSendMail(nil, goodmsg)
	*/
}

// from http://grokbase.com/t/gg/golang-nuts/137q8xwneh/go-nuts-smtp-self-signed-certificate
/*
 * server at addr, switches to TLS if possible,
 * authenticates with mechanism a if possible, and then sends an email from
 * address from, to addresses to, with message msg.
 */
//func SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {

// a can be nil
func WorkingSendMail(a smtp.Auth, msg string) error {

	from := "me@p-i-v-o-t-a.l-labs.com"
	to := []string{"me@p-i-v-o-t-a.l-labs.com"}
	addr := "mail.p-i-v-o-t-a.l-labs.com:25" // jea example

	c, err := smtp.Dial(addr)

	if err != nil {
		panic(err)
		return err
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	config := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	if err = c.StartTLS(config); err != nil {
		panic(err)
		return err
	}

	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, toaddr := range to {
		if err = c.Rcpt(toaddr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func sendGopherMail(body string, titlePrefix string) {

	addr := "localhost:2525"
	textbody := body
	htmlbody := "<font size=\"4\"><pre>\n" + body + "\n</pre></font>\n"
	t := time.Now()
	msg := &gophermail.Message{From: "me@p-i-v-o-t-a.l-labs.com",
		//ReplyTo string // optional
		To:       []string{"Me <me@p-i-v-o-t-a.l-labs.com>"},
		Subject:  titlePrefix + " test-mail-subject " + t.Local().String(),
		Body:     textbody,
		HTMLBody: htmlbody,
		//    Attachments []Attachment // optional
		// Extra mail headers.
		// Headers mail.Header
	}

	err := gophermail.SendMail(addr, nil, msg)
	if err != io.EOF && err != nil {
		panic(err)
	}
}
