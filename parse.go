package main

import (
	"fmt"
	"regexp"
	"strconv"
)

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
	pass = matches[5] == "passed"
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
