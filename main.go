package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/fractalcat/nagiosplugin"
	"github.com/jessevdk/go-flags"
)

var parser = flags.NewParser(&opts, flags.Default)
var preflightChecks []func()

/*
 * Main thread
 */
func main() {

	// Setup the Nagios check
	nagiosCheck := nagiosplugin.NewCheck()

	// Make sure the check always (as much as possible) exits with
	// the correct output and return code if we terminate unexpectedly.
	defer nagiosCheck.Finish()

	// Parse flags and quit if none supplied.
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	// Preflight checks. Add headers, do auth, etc.
	for i := 0; i < len(preflightChecks); i++ {
		preflightChecks[i]()
	}

	url, err := buildUrl(httpOpts.Ssl, httpOpts.Hostname, httpOpts.Uri)
	check(err)
	status, hdrs, body, size := httpRequest(httpOpts.Method, url, httpOpts.Post)

	if Tests["status"] {
		match, reason := checkStatus(status)
		if !match {
			FailReasons = append(FailReasons, reason)
		}
	}

	if Tests["page-size"] {
		match, reason := checkPageSize(size)
		if !match {
			FailReasons = append(FailReasons, reason)
		}
	}

	if Tests["headers"] {
		// Test headers(eg. conten-type=json)
		match, reason := checkHeaders(map[string][]string(hdrs))
		if !match {
			FailReasons = append(FailReasons, reason)
		}
	}

	if Tests["regexp"] {
		match, reason := checkRegexp(body)
		if !match {
			FailReasons = append(FailReasons, reason)
		}
	}

	if Tests["keys"] {
		// Unmarshal JSON into a map of strings
		// Variables will have to cast to be used.
		var respJson map[string]interface{}
		err = json.Unmarshal(body, &respJson)
		check(err)

		// Test keys in JSON response
		for _, tst := range JsonTests {
			match, reason := checkJson(respJson, tst)

			if !match && reason == nil {
				reason = errors.New(fmt.Sprintf("Key '%s' not in JSON response", tst.key))
			}

			if reason != nil {
				FailReasons = append(FailReasons, reason)
			}
		}

	}

	if len(FailReasons) != 0 {
		nagiosplugin.Exit(
			nagiosplugin.CRITICAL,
			fmt.Sprintf("Test(s) Failed: %s\n", FailReasons[0]),
		)
	} else {
		nagiosCheck.AddResult(nagiosplugin.OK, "All tests passed")
	}

	return
}
