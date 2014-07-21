package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

var rawPassedEmail = `Date: Mon, 21 Jul 2014 17:51:14 +0000 (UTC)
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

var rawFixedEmail = `From: jaten@pivotallabs.com
Sender: jaten@pivotallabs.com
Reply-To: jaten@pivotallabs.com
To: jaten@pivotallabs.com
Message-ID: <1654536890.9.1405980345430.JavaMail.jaten@pivotallabs.com>
Subject: Stage [testemailsending/9/defaultStage/1] is fixed
MIME-Version: 1.0
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit

See details: http://www.gocd.cf-app.com/go/pipelines/testemailsending/9/defaultStage/1

-- CHECK-INS --

Git: https://github.com/pivotal-cf-experimental/go-guerrilla
revision: cd41deac1668937fe00f49b6246beeed198ba307, modified by Jason E. Aten <jaten@pivotallabs.com> on 2014-07-21 21:47:36.0
subject line info from gocd email extracted
modified parse.go
modified parse_test.go

Sent by Go on behalf of releng
.
`

var rawBrokenEmail = `Date: Mon, 21 Jul 2014 21:54:37 +0000 (UTC)
From: jaten@pivotallabs.com
Sender: jaten@pivotallabs.com
Reply-To: jaten@pivotallabs.com
To: jaten@pivotallabs.com
Message-ID: <241171120.6.1405979677517.JavaMail.jaten@pivotallabs.com>
Subject: Stage [testemailsending/6/defaultStage/1] is broken
MIME-Version: 1.0
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit

See details: http://www.gocd.cf-app.com/go/pipelines/testemailsending/6/defaultStage/1

-- CHECK-INS --

Git: https://github.com/pivotal-cf-experimental/go-guerrilla
revision: cd41deac1668937fe00f49b6246beeed198ba307, modified by Jason E. Aten <jaten@pivotallabs.com> on 2014-07-21 21:47:36.0
subject line info from gocd email extracted
modified parse.go
modified parse_test.go

Sent by Go on behalf of releng
.
`

var rawFailedEmail = `Date: Mon, 21 Jul 2014 21:58:26 +0000 (UTC)
From: jaten@pivotallabs.com
Sender: jaten@pivotallabs.com
Reply-To: jaten@pivotallabs.com
To: jaten@pivotallabs.com
Message-ID: <1249059361.7.1405979906992.JavaMail.jaten@pivotallabs.com>
Subject: Stage [testemailsending/6/defaultStage/1] failed
MIME-Version: 1.0
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit

See details: http://www.gocd.cf-app.com/go/pipelines/testemailsending/6/defaultStage/2

-- CHECK-INS --

Git: https://github.com/pivotal-cf-experimental/go-guerrilla
revision: cd41deac1668937fe00f49b6246beeed198ba307, modified by Jason E. Aten <jaten@pivotallabs.com> on 2014-07-21 21:47:36.0
subject line info from gocd email extracted
modified parse.go
modified parse_test.go

Sent by Go on behalf of releng
.
`

var expectedPassedBody = `See details: http://www.gocd.cf-app.com/go/pipelines/testemailsending/2/defaultStage/1

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
			cv.So(BodyOfMail(rawPassedEmail), cv.ShouldEqual, expectedPassedBody)
		})
	})
}

func TestMailParsing(t *testing.T) {
	cv.Convey("Given a GoCD passing email message", t, func() {
		cv.Convey("pass/fail status, pipeline, stage, and their numbers should be extracted from the subject line", func() {
			parsedEmail := ParseEmail(rawPassedEmail)
			cv.So(parsedEmail.Pass, cv.ShouldEqual, true)
			cv.So(parsedEmail.PipelineName, cv.ShouldEqual, "testemailsending")
			cv.So(parsedEmail.PipelineBuild, cv.ShouldEqual, 2)
			cv.So(parsedEmail.StageName, cv.ShouldEqual, "defaultStage")
			cv.So(parsedEmail.StageBuild, cv.ShouldEqual, 1)
		})
		cv.Convey("and the same from a fixed subject line", func() {
			parsedEmail := ParseEmail(rawFixedEmail)
			cv.So(parsedEmail.Pass, cv.ShouldEqual, true)
			cv.So(parsedEmail.PipelineName, cv.ShouldEqual, "testemailsending")
			cv.So(parsedEmail.PipelineBuild, cv.ShouldEqual, 9)
			cv.So(parsedEmail.StageName, cv.ShouldEqual, "defaultStage")
			cv.So(parsedEmail.StageBuild, cv.ShouldEqual, 1)
		})
	})

	cv.Convey("Given a GoCD failing email message", t, func() {
		cv.Convey("from a failed subject line should yield pass/fail, pipeline name & no, stage name & no", func() {
			parsedEmail := ParseEmail(rawFailedEmail)
			cv.So(parsedEmail.Pass, cv.ShouldEqual, false)
			cv.So(parsedEmail.PipelineName, cv.ShouldEqual, "testemailsending")
			cv.So(parsedEmail.PipelineBuild, cv.ShouldEqual, 6)
			cv.So(parsedEmail.StageName, cv.ShouldEqual, "defaultStage")
			cv.So(parsedEmail.StageBuild, cv.ShouldEqual, 1)
		})
		cv.Convey("from a broken subject line should work the same way", func() {
			parsedEmail := ParseEmail(rawBrokenEmail)
			cv.So(parsedEmail.Pass, cv.ShouldEqual, false)
			cv.So(parsedEmail.PipelineName, cv.ShouldEqual, "testemailsending")
			cv.So(parsedEmail.PipelineBuild, cv.ShouldEqual, 6)
			cv.So(parsedEmail.StageName, cv.ShouldEqual, "defaultStage")
			cv.So(parsedEmail.StageBuild, cv.ShouldEqual, 1)
		})

	})
}

func TestGitRevisionRetrieval(t *testing.T) {
	cv.Convey("Given a GoCD passing email message", t, func() {
		cv.Convey("the git revision: line should be extracted from the body", func() {
			parsedEmail := ParseEmail(rawPassedEmail)
			cv.So(parsedEmail.GitRev, cv.ShouldEqual, "0111df6930fa11a28febde2197b591a5a67fb3e4, modified by Jason E. Aten <j.e.aten@gmail.com> on 2014-07-21 15:36:00.0")
		})
	})
}
