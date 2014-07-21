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

func GoCDSuccessEmail() string {
	return `
Date: Mon, 21 Jul 2014 17:51:14 +0000 (UTC)
From: jaten@pivotallabs.com
Sender: jaten@pivotallabs.com
Reply-To: jaten@pivotallabs.com
To: jaten@pivotallabs.com
Message-ID: <670094551.5.1405965074534.JavaMail.jaten@pivotallabs.com>
Subject: Stage [testemailsending/2/defaultStage/1] passed
MIME-Version: 1.0
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit

See details: http://www.gocd.cf-app.com/go/pipelines/testemailsending/2/defaultStage/1

-- CHECK-INS --

Git: https://github.com/pivotal-cf-experimental/go-guerrilla
revision: 0111df6930fa11a28febde2197b591a5a67fb3e4, modified by Jason E. Aten <j.e.aten@gmail.com> on 2014-07-21 15:36:00.0
new failing test for body of email extraction in place
added body.go
modified goguerrilla.conf
modified goguerrilla.go
added parse_test.go

Sent by Go on behalf of releng
.
`
}

func GoCDFailEmail() string {
	return `
Return-Path: <jaten@pivotallabs.com>
Received: from precise64 (gateway-sf.pivotallabs.com. [204.15.0.254])
        by mx.google.com with ESMTPSA id z2sm2921171pdj.16.2014.07.17.16.10.07
        for <jaten@pivotallabs.com>
        (version=TLSv1 cipher=ECDHE-RSA-RC4-SHA bits=128/128);
        Thu, 17 Jul 2014 16:10:08 -0700 (PDT)
Date: Thu, 17 Jul 2014 23:10:07 +0000 (UTC)
From: jaten@pivotallabs.com
Reply-To: jaten@pivotallabs.com
To: jaten@pivotallabs.com
Message-ID: <988236891.4.1405638608652.JavaMail.jaten@pivotallabs.com>
Subject: Stage [testpipe/4/alwaysfail/1] failed
MIME-Version: 1.0
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit

See details: http://10.0.2.15:8153/go/pipelines/testpipe/4/alwaysfail/1

-- CHECK-INS --

Git: https://github.com/JasonAtenPivotal/golang-mysql-starter
revision: b6bc4425c88912f97630cc316f9d06f1691807ed, modified by glycerine <j.e.aten@gmail.com> on 2014-07-16 07:47:53.0
gofmt
modified main.go

Sent by Go on behalf of admin
`
}
