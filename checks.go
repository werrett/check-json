package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/fractalcat/nagiosplugin"
)

type Options struct {
	FlagStatus func(int) `long:"status" short:"s" description:"Checks the numerical HTTP return status (eg. 200)"`

	FlagPageSize func(string) `long:"page-size" short:"m" description:"Checks response content length is in the given range (format: min:max)"`

	FlagHeaders func(string) `long:"header-equals" short:"d" description:"Key=value checks for HTTP response headers (key:value)"`

	FlagRegexp func(string) `long:"regexp" short:"r" description:"Checks the response body for a string using a regular expression."`

	FlagKeyExists func(string) `long:"key-exists" short:"e" description:"Checks existence of these keys from JSON response"`

	FlagKeyEquals func(string) `long:"key-equals" short:"q" description:"A regex to check the value of specific key values from JSON response"`

	FlagKeyLte func(string) `long:"key-lte" short:"l" description:"Check the returned value is less than this for a JSON key"`

	FlagKeyGte func(string) `long:"key-gte" short:"g" description:"Check the returned value is greater than this for a JSON key"`

	Verbose bool `long:"verbose" short:"v" description:"Display extra details (eg. response bodies) for debugging" default:"false"`
}

var opts Options

// A map of tests to actually perform. Driven by flags.
var Tests = make(map[string]bool)

// A slice of reasons behind failed checks.
var FailReasons = make([]error, 0)

// Test the responses numeric status (eg. 200)
var StatusTest int

// Min/max size for response Content-Length
var PageSizeTest = map[string]int64{"min": 0, "max": 0}

// Tests on the response Headers. Consists of header name and expected val.
var HeaderTests = make(map[string]string)

// Test the response body using a regex match
var RegexpTest *regexp.Regexp

// Tests on the response JSON body.
// Each test has an operator (eg equals) and a value (eg. "success")
var JsonTests = make([]JsonTest, 0)

func init() {

	opts.FlagStatus = func(code int) {
		Tests["status"] = true

		StatusTest = code
	}

	opts.FlagPageSize = func(str string) {
		Tests["page-size"] = true

		s, err := parseFlagPair("page-size", str)
		check(err)

		min, err := strconv.ParseInt(s[0], 10, 0)
		if err != nil {
			nagiosplugin.Exit(
				nagiosplugin.CRITICAL,
				fmt.Sprintf("Page size min '%s' parameter is not an integer", s[0]),
			)
		}
		PageSizeTest["min"] = min

		max, err := strconv.ParseInt(s[1], 10, 0)
		if err != nil {
			nagiosplugin.Exit(
				nagiosplugin.CRITICAL,
				fmt.Sprintf("Page size max '%s' parameter is not an integer", s[1]),
			)
		}
		PageSizeTest["max"] = max

		if max < min {
			nagiosplugin.Exit(
				nagiosplugin.CRITICAL,
				fmt.Sprintf("Page size range must in format max:min"),
			)
		}

	}

	opts.FlagHeaders = func(str string) {
		Tests["headers"] = true

		s, err := parseFlagPair("header-equals", str)
		check(err)

		HeaderTests[s[0]] = s[1]
	}

	opts.FlagRegexp = func(str string) {
		Tests["regexp"] = true

		var err error
		RegexpTest, err = regexp.Compile(str)

		if err != nil {
			nagiosplugin.Exit(
				nagiosplugin.CRITICAL,
				fmt.Sprintf("String '%s' not a valid regexp: %s", str, err),
			)
		}
	}

	// JSON keys to test if they exist
	opts.FlagKeyExists = func(str string) {
		Tests["keys"] = true
		JsonTests = append(JsonTests, JsonTest{str, "", "exists"})
	}

	opts.FlagKeyEquals = func(str string) {
		Tests["keys"] = true

		s, err := parseFlagPair("key-equals", str)
		check(err)
		JsonTests = append(JsonTests, JsonTest{s[0], s[1], "equals"})
	}

	opts.FlagKeyLte = func(str string) {
		Tests["keys"] = true

		s, err := parseFlagPair("key-lte", str)
		check(err)

		v, err := strconv.ParseFloat(s[1], 64)
		if err != nil {
			nagiosplugin.Exit(
				nagiosplugin.CRITICAL,
				fmt.Sprintf("Key '%s' parameter is not an integer", s[1]),
			)
		}

		JsonTests = append(JsonTests, JsonTest{s[0], v, "lte"})
	}

	opts.FlagKeyGte = func(str string) {
		Tests["keys"] = true

		s, err := parseFlagPair("key-gte", str)
		check(err)

		v, err := strconv.ParseFloat(s[1], 64)
		if err != nil {
			nagiosplugin.Exit(
				nagiosplugin.CRITICAL,
				fmt.Sprintf("Key '%s' parameter is not an integer", s[1]),
			)
		}

		JsonTests = append(JsonTests, JsonTest{s[0], v, "gte"})
	}

}

/*
 * Check functions
 */

// Check the numerical HTTP return code
func checkStatus(status int) (bool, error) {

	// Return false if we don't find out specific HTTP status
	if status != StatusTest {
		return false, errors.New(
			fmt.Sprintf("HTTP Status Code was '%d', expected '%d'", status, StatusTest))
	}

	return true, nil // All tests passed, no errors
}

// Check the HTTP response size
func checkPageSize(size int64) (bool, error) {

	// Return false if we don't find out specific HTTP status
	if size < PageSizeTest["min"] || PageSizeTest["max"] < size {
		return false, errors.New(
			fmt.Sprintf("HTTP Response Size was '%d', out side of range '%d-%d'",
				size, PageSizeTest["min"], PageSizeTest["max"]))
	}

	return true, nil // All tests passed, no errors
}

// Check headers in the HTTP response
func checkHeaders(hdrs map[string][]string) (bool, error) {

	for k, v := range HeaderTests { // Iterate through tests

		if hdrs[k] == nil {
			return false,
				errors.New(fmt.Sprintf("Header '%s' not in HTTP response", k))
		}

		ret := false
		for _, h := range hdrs[k] { // One header can have multiple values
			match, _ := regexp.MatchString(v, h)
			if match {
				ret = true // Return true if there is at least one hit
			}
		}

		if !ret {
			return false,
				errors.New(
					fmt.Sprintf("Header '%s' does not equal '%s'", k, v),
				)
		}

	}

	return true, nil // All tests passed, no errors
}

// Check whether body includes a regex string
func checkRegexp(body []byte) (bool, error) {

	match := RegexpTest.Match(body)
	if !match {
		return false,
			errors.New(fmt.Sprintf("Regexp '%s' not in HTTP response",
				RegexpTest.String()))
	}

	return true, nil // All tests passed, no errors
}

// Check JSON variabes in response body
func checkJson(j interface{}, tst JsonTest) (bool, error) {

	var match bool
	var failReasons error

	// Unmarshall generic decoding:
	//  - interface{} = strings, integers, and booleans,
	//  - []interface{} = arrays,
	//  - map[string]interface{} for objects
	//
	// See http://stackoverflow.com/a/22470287 &&
	// http://blog.golang.org/json-and-go

	switch t := j.(type) {

	// We have a JSON array
	// eg. ["one", "two", "three"]
	case []interface{}:
		for _, v := range t {
			var result bool
			// FIXME: Need to pass the key name with the value.
			result, failReasons = checkJson(v, tst)
			match = match || result
		}

	// We have further JSON objects to decode
	// eg. {"obj1": {"obj2": {"key": "value"}}}
	case map[string]interface{}:
		jsn := j.(map[string]interface{})

		// If jsn has the key we're looking for. Test it.
		if jsn[tst.key] != nil {

			match, failReasons = checkJsonValue(jsn, tst)

		} else { // Do further unzipping. There might be more objects.
			match = false

			for _, v := range t {
				result, err := checkJson(v, tst)
				match = match || result
				if err != nil {
					failReasons = err
				}
			}

		}

	}

	return match, failReasons
}

// Check the value of a specific JSON key:value object
func checkJsonValue(jsn map[string]interface{}, tst JsonTest) (bool, error) {

	// Switch based on test type
	switch tst.operator {

	case "equals":
		// Convert JSON value to string and do a regex match
		jv := fmt.Sprintf("%s", jsn[tst.key])
		match, _ := regexp.MatchString(tst.value.(string), jv)
		if !match {
			return true,
				errors.New(
					fmt.Sprintf("Key '%s' does not equal '%s'", tst.key, tst.value),
				)
		}

	case "lte":
		_, ok := jsn[tst.key].(float64)
		if !ok {
			return true,
				errors.New(
					fmt.Sprintf("Key '%s' value is not an integer", tst.key),
				)
		}

		if tst.value.(float64) < jsn[tst.key].(float64) {
			return true,
				errors.New(
					fmt.Sprintf("Key '%s' is greater than '%g'", tst.key, tst.value),
				)
		}

	case "gte":
		_, ok := jsn[tst.key].(float64)
		if !ok {
			return true,
				errors.New(
					fmt.Sprintf("Key '%s' value is not an integer", tst.key),
				)
		}

		if tst.value.(float64) > jsn[tst.key].(float64) {
			return true,
				errors.New(
					fmt.Sprintf("Key '%s' is less than '%g'", tst.key, tst.value),
				)
		}

	case "exists":
		if jsn[tst.key] == nil {
			return false, nil
		}

	}
	return true, nil // Json key exists and all tests passed
}
