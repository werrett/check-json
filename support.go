package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
)

// Struct to hold tests of JSON response (key exists, key equals, etc)
type JsonTest struct {
	key      string // Key is a new field.
	value    interface{}
	operator string
}

var flagSeperator = ":"
var flagPairRegexp = regexp.MustCompile(".+" + flagSeperator + ".+")

/*
 * Generic helper functions
 */

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parseFlagPair(flagName, flagValue string) ([]string, error) {

	match := flagPairRegexp.MatchString(flagValue)
	if !match {
		return nil, errors.New(
			fmt.Sprintf("Flag %s needs to be in 'key:value' format", flagName),
		)
	}

	return strings.Split(flagValue, flagSeperator), nil
}

/*
 * Generic helper functions for Tests
 */

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf(
			"Expected %v - Got %v",
			a, b,
		)
	}
}

func expectErr(t *testing.T, a interface{}) {
	_, ok := a.(error)
	if !ok {
		t.Errorf("Expected error.")
	}
}

func httpServer(code int, hdrs http.Header, body string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Set HTTP headers
		for hdrKey, hdrArray := range hdrs {
			for _, hdrVal := range hdrArray {
				w.Header().Set(hdrKey, hdrVal)
			}
		}

		// Set the repsonse HTTP status (eg. 200)
		w.WriteHeader(code)

		// Set HTTP response body
		fmt.Fprintln(w, body)
	}))
	return ts
}

/*
 * File helpers
 */

func readFile(f string) io.Reader {
	file, err := os.Open(f)
	check(err)
	return bufio.NewReader(file)
}

func readStdIn() io.Reader {
	return bufio.NewReader(os.Stdin)
}

/*
 * Byte / stream helpers
 */

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

func streamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.String()
}
