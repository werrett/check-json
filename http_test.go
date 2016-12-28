package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

/*
 * Data models to hold HTTP test cases
 */

type BuildUrlCase struct {
	ssl                 bool
	hostname, uri, want string
}

type HttpRequestCase struct {
	code int
	hdrs http.Header
	body string
	size int64
}

/*
 * Tests for primary functions
 */

func Test_buildUrl_Success(t *testing.T) {

	// Normal URL build cases
	cases := []BuildUrlCase{
		{false, "localhost", "/wibble", "http://localhost/wibble"},
		{true, "localhost", "/wibble", "https://localhost/wibble"},
		{false, "localhost", "", "http://localhost/"},
	}

	for _, c := range cases {
		got, _ := buildUrl(c.ssl, c.hostname, c.uri)
		expect(t, c.want, got)
	}

}

func Test_buildUrl_Errors(t *testing.T) {
	// URL build error cases
	cases := []BuildUrlCase{{false, "", "/wibble", ""}}

	for _, c := range cases {
		_, err := buildUrl(c.ssl, c.hostname, c.uri)
		expectErr(t, err)
	}
}

func Test_httpRequest(t *testing.T) {

	cases := []HttpRequestCase{
		{200, http.Header{"X-Pot-Type": {"Teapot"}}, "I am a teapot", 14},
		{401, nil, "You're not the boss of me.", 27},
		{404, nil, "This not the webpage you're looking for.", 41},
	}

	for _, c := range cases {
		ts := httpServer(c.code, c.hdrs, c.body)
		defer ts.Close()

		code, hdrs, body, size := httpRequest("GET", ts.URL, "")

		expect(t, c.code, code)
		expect(t, c.body, strings.TrimSpace(fmt.Sprintf("%s", body)))
		expect(t, c.size, size)

		// Iterate through headers set during test case and make sure it
		// is in the response
		for hdrKey, hdrArray := range c.hdrs {
			for i, hdrVal := range hdrArray {
				expect(t, hdrVal, hdrs[hdrKey][i])
			}
		}

	}
}
