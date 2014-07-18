package main

import (
	"testing"
	"os/exec"
	"github.com/glycerine/gophermail"
	"fmt"
	"time"
	cv "github.com/smartystreets/goconvey/convey"
)


func TestMailSendReceive(t *testing.T) {
	cv.Convey("Given a running instance of go-guerrilla", t, func() {

	        addr := "localhost:2525"
		mailchan := make(chan gophermail.Message)
	        guer := NewGoGuerrillaSmtpServer(addr, mailchan)

		subject := "test-subject"
		body := "test-body"

		cv.Convey(fmt.Sprintf("it should receive mail when sent mail using our gomailclient utility to %v", addr), func() {

			c := exec.Command("gomailclient/gomailclient")
			c.Args = []string{addr, subject, body}
			_, err := c.Output()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Waiting for go-guerrilla to receive the email")

			select {
			       case receievedMail := <- mailchan:
				   cv.So(receivedMail.Body, cv.ShouldEqual, body)
		 	           cv.So(receivedMail.Subject, cv.ShouldEqual, subject)
                               case <- time.After(10 * 1e9):
                                   fmt.Printf("go-guerilla did not recieve email after 10 seconds, failing test.\n")
                                   cv.So(true, cv.ShouldEqual, false)
			}

			
		})
	})
}
