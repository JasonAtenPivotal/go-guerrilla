package main

import (
	"fmt"
	"regexp"
)

var bodyReString = `Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit\n\n((.|\n)+).\n`

var bodyRegex = regexp.MustCompile(bodyReString)

func BodyOfMail(msg string) string {
	matches := bodyRegex.FindStringSubmatch(msg)
	if matches == nil || len(matches) < 1 {
		panic(fmt.Sprintf("no body delimiter '%s' found in msg: %s", bodyReString, msg))
	}
	return matches[1]
}

type GoCDNotice struct {
	Pass bool
	// [testemailsending/2/defaultStage/1]
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

func ParseEmail(email string) *GoCDNotice {
	return NewGoCDNotice()
}
