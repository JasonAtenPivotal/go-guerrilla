package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

var rawSuccessEmail = `Date: Mon, 21 Jul 2014 17:51:14 +0000 (UTC)
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

var expectedSuccessBody = `See details: http://www.gocd.cf-app.com/go/pipelines/testemailsending/2/defaultStage/1

-- CHECK-INS --

Git: https://github.com/pivotal-cf-experimental/go-guerrilla
revision: 0111df6930fa11a28febde2197b591a5a67fb3e4, modified by Jason E. Aten <j.e.aten@gmail.com> on 2014-07-21 15:36:00.0
new failing test for body of email extraction in place
added body.go
modified goguerrilla.conf
modified goguerrilla.go
added parse_test.go

Sent by Go on behalf of releng
`

func TestMailBodyExtraction(t *testing.T) {
	cv.Convey("Given a GoCD success mail message", t, func() {
		cv.Convey("we should be able to extract the body", func() {
			cv.So(BodyOfMail(rawSuccessEmail), cv.ShouldEqual, expectedSuccessBody)
		})
	})
}

func TestMailParsing(t *testing.T) {
	cv.Convey("Given a GoCD mail message", t, func() {
		cv.Convey("pass/fail status should be extracted from the subject line", func() {
			parsedEmail := ParseEmail(rawSuccessEmail)
			cv.So(parsedEmail.Pass, cv.ShouldEqual, true)
			cv.So(parsedEmail.PipelineName, cv.ShouldEqual, "testemailsending")
			cv.So(parsedEmail.PipelineBuild, cv.ShouldEqual, 1)
			cv.So(parsedEmail.StageName, cv.ShouldEqual, "defaultStage")
			cv.So(parsedEmail.StageBuild, cv.ShouldEqual, 1)
		})
	})
}
