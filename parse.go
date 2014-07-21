package main

import (
	"fmt"
	"regexp"
	"strconv"
)

/* notes on GoCD notification meaning:

http://www.gocd.cf-app.com/go/help/dev_notifications.html

You can set up notifications for different events

All - all the runs for the stage
Passes - the passed runs for the stage
Fails - the stage runs that failed
Breaks - the stage run that broke the build
Fixed - the stage run that fixed the previous failure
Cancelled - the stage run that was cancelled

Illustration

At the moment:
Previous build Pass, current build Fail: Event: Break } we treat these two as failures
Previous build Fail, current build Fail: Event: Fail  }
Previous build Fail, current build Pass: Event: Fixed } we treat these two as pass
Previous build Pass, current build Pass: Event: Pass  }

*/

var bodyReString = `Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit\n\n((.|\n)+).\n`

var bodyRegex = regexp.MustCompile(bodyReString)

var subjectReString = `\nSubject: (.+)\n`
var subjectRegex = regexp.MustCompile(subjectReString)

func BodyOfMail(msg string) string {
	matches := bodyRegex.FindStringSubmatch(msg)
	if matches == nil || len(matches) < 1 {
		panic(fmt.Sprintf("no body delimiter '%s' found in msg: %s", bodyReString, msg))
	}
	return matches[1]
}

func SubjectOfMail(msg string) string {
	matches := subjectRegex.FindStringSubmatch(msg)
	if matches == nil || len(matches) < 1 {
		panic(fmt.Sprintf("no body delimiter '%s' found in msg: %s", subjectReString, msg))
	}
	return matches[1]
}

type GoCDNotice struct {
	Pass          bool
	PipelineName  string
	PipelineBuild int
	StageName     string
	StageBuild    int
	GitRev        string
	Checkins      string
	DetailsURL    string
}

func NewGoCDNotice() *GoCDNotice {
	return &GoCDNotice{}
}

func ParseSubject(subject string) (pass bool, pipe string, pipeno int, stage string, stageno int) {
	var subjectString = `Stage \[([^/]+)\/(\d+)\/([^/]+)\/(\d+)\] (.+)`
	var subjectRe = regexp.MustCompile(subjectString)
	matches := subjectRe.FindStringSubmatch(subject)
	if matches == nil || len(matches) < 6 {
		panic(fmt.Sprintf("bad subject '%s' matched by %s", subject, subjectString))
	}

	var err error
	pass = (matches[5] == "passed" || matches[5] == "is fixed")
	pipe = matches[1]
	pipeno, err = strconv.Atoi(matches[2])
	if err != nil {
		panic(fmt.Sprintf("failed to parse pipeno: %s", matches[2]))
	}
	stage = matches[3]
	stageno, err = strconv.Atoi(matches[4])
	if err != nil {
		panic(fmt.Sprintf("failed to parse stageno: %s", matches[4]))
	}
	return
}

func ParseEmail(email string) *GoCDNotice {

	//body := BodyOfMail(email)
	subject := SubjectOfMail(email)

	pass, pipe, pipeno, stage, stageno := ParseSubject(subject)

	n := NewGoCDNotice()
	n.Pass = pass
	n.PipelineName = pipe
	n.PipelineBuild = pipeno
	n.StageName = stage
	n.StageBuild = stageno
	return n
}
