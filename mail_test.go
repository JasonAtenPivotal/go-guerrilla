package main

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestMailSendReceive(t *testing.T) {
	addr := "localhost:2525"
	mailchan := make(chan *Client)
	guer := NewGoGuerrillaSmtpd(addr, mailchan)
	guer.Start()
	defer guer.Shutdown()

	cv.Convey("Given a running instance of go-guerrilla", t, func() {

		subject := "test-subject"
		body := "test-body"
		var receivedMail *Client

		cv.Convey(fmt.Sprintf("it should receive mail when sent mail using our mailsender.go:SendTestMail() routine to: %v", addr), func() {

			if false {
				c := exec.Command("gomailclient/gomailclient")
				c.Args = []string{addr, subject, body}
				_, err := c.Output()
				if err != nil {
					panic(err)
				}
			}
			go SendTestMail(subject, body)

			fmt.Printf("\n       ... mail_test waiting for go-guerrilla to receive the email.\n")

			select {
			case receivedMail = <-mailchan:
				cv.So(receivedMail.subject, cv.ShouldEqual, `"`+subject+`"`)
				fmt.Printf("\n     good: received mail with expected subject: \"%s\"\n", subject)
			case <-time.After(10 * 1e9):
				fmt.Printf("go-guerilla did not recieve email after 10 seconds, failing test.\n")
				cv.So(true, cv.ShouldEqual, false)
			}

		})
	})
}
