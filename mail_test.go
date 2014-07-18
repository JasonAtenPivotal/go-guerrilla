package main

import (
	"testing"
	"os/exec"
	"github.com/glycerine/gophermail"
	"fmt"
	cv "github.com/smartystreets/goconvey/convey"
)


func TestMailSendReceive(t *testing.T) {
	cv.Convey("A running instance of go-guerrilla", t, func() {

	        addr := "localhost:25"
		mailchan := make(chan gophermail.Message)
	        guer := NewGoGuerrillaSmtpServer(addr, mailchan)

		subject := "test-subject"
		body := "test-body"

		cv.Convey(fmt.Sprintf("should receive mail when sent mail using our gomailclient utility to %v", addr), func() {

			c := exec.Command("gomailclient/gomailclient")
			c.Args = []string{addr, subject, body}
			_, err := c.Output()
			if err != nil {
				panic(err)
			}

			receievedMail := <- mailchan

			cv.So(receivedMail.Body, cv.ShouldEqual, body)
			cv.So(receivedMail.Subject, cv.ShouldEqual, subject)
			
		})
	})
}
