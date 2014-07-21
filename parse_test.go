package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestMailParsing(t *testing.T) {

	cv.Convey("Given a mail message", t, func() {

		msg := `
To: "Me" <me@p-i-v-o-t-a.l-labs.com>
From: <me@p-i-v-o-t-a.l-labs.com>
Subject: "test-subject"
Mime-Version: 1.0
Content-Type: multipart/mixed;
 boundary=b7877c6ee71cd8d2a363751f0782a83f9449fd9adab87cad2b2ac0af0822

--b7877c6ee71cd8d2a363751f0782a83f9449fd9adab87cad2b2ac0af0822
Content-Type: multipart/alternative;
 boundary=10ed87264bf00684a2782eb8e5af5a479e9644010088bc37164d511ad53f

--10ed87264bf00684a2782eb8e5af5a479e9644010088bc37164d511ad53fw
Content-Type: text/plain; charset=utf-8
Content-Transfer-Encoding: quoted-printable

test-body
--10ed87264bf00684a2782eb8e5af5a479e9644010088bc37164d511ad53f
Content-Transfer-Encoding: base64
Content-Type: text/html; charset=utf-8

PGZvbnQgc2l6ZT0iNCI+PHByZT4KdGVzdC1ib2R5CjwvcHJlPjwvZm9udD4K
--10ed87264bf00684a2782eb8e5af5a479e9644010088bc37164d511ad53f--

--b7877c6ee71cd8d2a363751f0782a83f9449fd9adab87cad2b2ac0af0822--
.
`
		cv.Convey("we should be able to extract the body from the mime-types", func() {
			cv.So(BodyOfMail(msg), cv.ShouldEqual, "test-body")
		})
	})
}
