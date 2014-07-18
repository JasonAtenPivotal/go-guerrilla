package main

import (
	"testing"
	"os/exec"
	//"fmt"
	"io/ioutil"
	"os"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestCertGen(t *testing.T) {
	origdir, tmpdir := GenTempDir()
	//fmt.Printf("attempting to generate certs in tempdir: '%s'\n", tmpdir)

	defer CleanupTempDir(origdir, tmpdir)

	cv.Convey("Given a built generate_cert/generate_cert utility program", t, func() {

		cv.Convey("it should generate certs in our temp directory of the expected file length.", func() {

			c := exec.Command(origdir + "/generate_cert/generate_cert")
			_, err := c.Output()
			if err != nil {
				panic(err)
			}
			certExists, certFi := FileExists(tmpdir + "/cert.pem")
			keyExists, keyFi := FileExists(tmpdir + "/key.pem")

			//fmt.Printf("certFi = %#v\n", certFi)
			//fmt.Printf("keyFi  = %#v\n", keyFi)

			cv.So(certExists, cv.ShouldEqual, true)
			cv.So(keyExists, cv.ShouldEqual, true)

			cv.So(certFi.Size(), cv.ShouldEqual, 1074)
			cv.So(keyFi.Size(), cv.ShouldBeGreaterThan, 1600)
			
		})
	})
}

func GenTempDir() (origdir string, tmpdir string) {

	// make new temp dir that will have no ".goqclusterid files in it
	var err error
	origdir, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	tmpdir, err = ioutil.TempDir(origdir, "tempdir")
	if err != nil {
		panic(err)
	}
	err = os.Chdir(tmpdir)
	if err != nil {
		panic(err)
	}

	return origdir, tmpdir
}

func CleanupTempDir(origdir string, tmpdir string) {
	os.Chdir(origdir)
	err := os.RemoveAll(tmpdir)
	if err != nil {
		panic(err)
	}
}

func FileExists(name string) (bool, os.FileInfo) {
	fi, err := os.Stat(name)
	if err != nil {
		return false, nil
	}
	if fi.IsDir() {
		return false, nil
	}
	return true, fi
}
