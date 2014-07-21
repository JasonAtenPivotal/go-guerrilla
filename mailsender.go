package main

import (
	"io"

	"github.com/glycerine/gophermail"
)

func SendTestMail(subject string, body string) {
	//fmt.Printf("SendTestMail() starting.\n")
	sendGopherMail(subject, body)
	//log.Printf("SentTestMail() done.")
}

func sendGopherMail(subject string, body string) {

	addr := "localhost:2525"
	textbody := body
	htmlbody := "<font size=\"4\"><pre>\n" + body + "\n</pre></font>\n"
	//t := time.Now()
	msg := &gophermail.Message{From: "me@p-i-v-o-t-a.l-labs.com",
		//ReplyTo string // optional
		To:       []string{"Me <me@p-i-v-o-t-a.l-labs.com>"},
		Subject:  subject,
		Body:     textbody,
		HTMLBody: htmlbody,
		//    Attachments []Attachment // optional
		// Extra mail headers.
		// Headers mail.Header
	}

	//fmt.Printf("SendTestMail()/sendGopherMail() about to call gophermail.SendMail().\n")
	err := gophermail.SendMail(addr, nil, msg)
	if err != io.EOF && err != nil {
		panic(err)
	}
}
