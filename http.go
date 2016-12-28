package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
)

type HttpOptions struct {
	Hostname string `long:"hostname" short:"H" description:"Web server to query"`

	Uri string `long:"uri" short:"u" description:"URI to GET or POST" default:"/"`

	Method string `long:"method" short:"j" description:"HTTP method (eg. HEAD, OPTIONS, TRACE, PUT, DELETE)" default:"GET"`

	Post string `long:"post" short:"P" description:"Body of POST Request"`

	Authorization func(string) `long:"authorization" short:"a" description:"Basic HTTP auth (username:password)"`

	Ssl bool `long:"ssl" short:"S" description:"Enforce SSL" default:"false"`

	Headers map[string]string `long:"header" short:"k" description:"Key,value pairs to add as headers in HTTP request (name:value format)"`
}

var httpOpts HttpOptions

func init() {
	// If Authorization flag provided, add authentication headers
	httpOpts.Authorization = func(str string) {
		data := base64.StdEncoding.EncodeToString([]byte(str))
		httpOpts.Headers["Authorization"] = "Basic " + data
	}

	parser.AddGroup("HTTP Options", "HTTP", &httpOpts)
}

func buildUrl(ssl bool, hostname string, uri string) (string, error) {

	if hostname == "" {
		return "", errors.New("Hostname is blank")
	}

	// Check the URI for template text, if so send it to the handler
	if templateTest(uri) {
		uri = templateHndlr(uri)
	}

	if uri == "" {
		uri = "/"
	}

	// If SSL enforce, add https:// in front of the URL
	var protocol string
	if ssl {
		protocol = "https://"
	} else {
		protocol = "http://"
	}

	return fmt.Sprintf("%s%s%s", protocol, hostname, uri), nil
}

func setReqHeaders(req *http.Request, hdrs map[string]string) {

	for k, v := range hdrs { // Get all, cept last values
		req.Header[k] = []string{v}
	}
}

func httpRequest(
	method string,
	urlStr string,
	bodyFile string,
) (int, map[string][]string, []byte, int64) {

	// Build the API Request
	var req *http.Request
	var err error

	switch bodyFile {
	case "": // Nil request body
		req, err = http.NewRequest(method, urlStr, nil)

	case "stdin": // Read request from standard in
		bodyStm := readStdIn()
		bodyStr := streamToString(bodyStm)

		// Check the POST body for template text, if so send it to the handler
		if templateTest(bodyStr) {
			bodyStr = templateHndlr(bodyStr)
		}

		newBody := strings.NewReader(bodyStr)
		req, err = http.NewRequest(method, urlStr, newBody)

	default: // Load request body from file
		bodyStm := readFile(bodyFile)
		bodyStr := streamToString(bodyStm)

		// Check the POST body for template text, if so send it to the handler
		if templateTest(bodyStr) {
			bodyStr = templateHndlr(bodyStr)
		}

		newBody := strings.NewReader(bodyStr)
		req, err = http.NewRequest(method, urlStr, newBody)

	}
	check(err)

	// Add the appropriate headers to the request
	setReqHeaders(req, httpOpts.Headers)

	// Print the request body if verbose flag set.
	if opts.Verbose {
		dat, err := httputil.DumpRequest(req, true)
		check(err)
		fmt.Printf("%s\n", dat)
	}

	// Make the HTTP Request
	client := http.DefaultClient
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	// Print the response body if verbose flag set.
	if opts.Verbose {
		dat, err := httputil.DumpResponse(resp, true)
		check(err)
		fmt.Printf("%s\n", dat)
	}

	// Read the API request response
	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	return resp.StatusCode, resp.Header, body, resp.ContentLength
}
