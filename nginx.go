package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// If running Nginx as a proxy, give Nginx the IP address and port for the SMTP server
// Primary use of Nginx is to terminate TLS so that Go doesn't need to deal with it.
// This could perform auth and load balancing too
// See http://wiki.nginx.org/MailCoreModule
func (s *Smtpd) nginxHTTPAuth() {
	parts := strings.Split(gConfig["GSTMP_LISTEN_INTERFACE"], ":")
	gConfig["HTTP_AUTH_HOST"] = parts[0]
	gConfig["HTTP_AUTH_PORT"] = parts[1]
	fmt.Println(parts)
	http.HandleFunc("/", s.nginxHTTPAuthHandler)
	err := http.ListenAndServe(gConfig["NGINX_AUTH"], nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func (s *Smtpd) nginxHTTPAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Auth-Status", "OK")
	w.Header().Add("Auth-Server", gConfig["HTTP_AUTH_HOST"])
	w.Header().Add("Auth-Port", gConfig["HTTP_AUTH_PORT"])
	fmt.Fprint(w, "")
}
