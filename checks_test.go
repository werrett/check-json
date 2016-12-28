package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

/*
 * Data models to hold test cases for various checks
 */

type TestStatusCase struct {
	option func(int)
	param  int
	send   int
	match  bool
	errStr string
}

type TestPageSizeCase struct {
	option func(string)
	param  string
	send   int64
	match  bool
	errStr string
}

type TestHeadersCase struct {
	option func(string)
	param  string
	send   http.Header
	match  bool
	errStr string
}

type TestRegexpCase struct {
	option func(string)
	param  string
	send   []byte
	match  bool
	errStr string
}

type TestJsonCase struct {
	option func(string)
	param  string
	send   []byte
	match  bool
	errStr string
}

type TestJsonValueCase struct {
	jsonBlob []byte
	test     JsonTest
	match    bool
	errStr   string
}

/*
 * Tests for primary functions
 */

func Test_checkStatus(t *testing.T) {

	cases := []TestStatusCase{
		{opts.FlagStatus, 200, 200, true, ""},
		{opts.FlagStatus, 404, 404, true, ""},
		{opts.FlagStatus, 404, 200, false,
			"HTTP Status Code was '200', expected '404'"},
	}

	for _, c := range cases {
		// Call the option with the param specified in the case.
		// eg. opts.FlagStatus(200) simulates --status=200
		c.option(c.param)

		// Test the flag that drives the test
		expect(t, Tests["status"], true)

		// Run status check
		match, err := checkStatus(c.send)
		expect(t, c.match, match)

		// If we expect didn't expect a match test the error code
		if !c.match {
			expect(t, c.errStr, err.Error())
		}

	}
}

func Test_checkPageSize(t *testing.T) {

	cases := []TestPageSizeCase{
		{opts.FlagPageSize, "0:124", 100, true, ""},
		{opts.FlagPageSize, "1024:65535", 1024, true, ""},
		{opts.FlagPageSize, "5000:10000", 10000, true, ""},
		{opts.FlagPageSize, "5000:10000", 100, false,
			"HTTP Response Size was '100', out side of range '5000-10000'"},
		{opts.FlagPageSize, "10:11", 12, false,
			"HTTP Response Size was '12', out side of range '10-11'"},
	}

	for _, c := range cases {
		// Call the option with the param specified in the case.
		// eg. opts.FlagPageSize("0:124") simulates --page-size=0:124
		c.option(c.param)

		// Test the flag that drives the test
		expect(t, Tests["page-size"], true)

		// Run page size check
		match, err := checkPageSize(c.send)
		expect(t, c.match, match)

		// If we expect didn't expect a match test the error code
		if !c.match {
			expect(t, c.errStr, err.Error())
		}

	}
}

func Test_testHeaders(t *testing.T) {

	cases := []TestHeadersCase{

		{opts.FlagHeaders, "X-Pot-Type:Teapot",
			http.Header{"X-Pot-Type": {"Teapot"}}, true, ""},

		// Multiple headers share the same key.
		{opts.FlagHeaders, "X-Pot-Type:Teapot",
			http.Header{"X-Pot-Type": {"Teapot", "Flowerpot"}}, true, ""},

		// Multiple headers share the same key.
		{opts.FlagHeaders, "X-Pot-Type:Teapot",
			http.Header{"X-Pot-Type": {"Potpan", "Teapot"}}, true, ""},

		// Required headers not correct value
		{opts.FlagHeaders, "X-Pot-Type:Teapot",
			http.Header{"X-Pot-Type": {"Polpot"}}, false,
			"Header 'X-Pot-Type' does not equal 'Teapot'"},

		// Required headers not present.
		{opts.FlagHeaders, "X-Pot-Type:Teapot",
			http.Header{"X-No-Pots": {"Here"}}, false,
			"Header 'X-Pot-Type' not in HTTP response"},

		// No headers returned at all.
		{opts.FlagHeaders, "X-Pot-Type:Teapot",
			nil, false,
			"Header 'X-Pot-Type' not in HTTP response"},
	}

	for _, c := range cases {
		// Call the option with the param specified in the case.
		// eg. opts.FlagHeaders("key:val") simulates --headers=key:val
		c.option(c.param)

		// Test the flag that drives the test
		expect(t, Tests["headers"], true)

		// Run header check
		match, err := checkHeaders(c.send)
		expect(t, c.match, match)

		// If we expect didn't expect a match the error code
		if !c.match {
			expect(t, c.errStr, err.Error())
		}

	}
}

func Test_checkRegexp(t *testing.T) {

	cases := []TestRegexpCase{
		{opts.FlagRegexp, "^The.*fox.+dog$",
			[]byte("The quick brown fox ... lazy dog"), true, ""},

		{opts.FlagRegexp, "dwarves",
			[]byte("Here be dragons"), false,
			"Regexp 'dwarves' not in HTTP response"},
	}

	for _, c := range cases {
		// Call the option with the param specified in the case.
		// eg. opts.FlagRegexp(".+") simulates --regexp=.+
		c.option(c.param)

		// Test the flag that drives the test
		expect(t, Tests["regexp"], true)

		// Run regex test
		match, err := checkRegexp(c.send)
		expect(t, c.match, match)

		// If we expect didn't expect a match the error code
		if !c.match {
			expect(t, c.errStr, err.Error())
		}

	}
}

func Test_checkJson(t *testing.T) {

	cases := []TestJsonCase{
		{opts.FlagKeyExists, "foo",
			[]byte(`{"baz":"qux", "foo":"bar"}`), true, ""},

		{opts.FlagKeyEquals, "foo:bar",
			[]byte(`{"foo":"bar", "baz":"qux"}`), true, ""},

		{opts.FlagKeyLte, "foo:100",
			[]byte(`{"foo":50, "baz":"qux"}`), true, ""},

		{opts.FlagKeyLte, "foo:10.00001",
			[]byte(`{"baz":"qux", "foo":10}`), true, ""},

		{opts.FlagKeyGte, "foo:100",
			[]byte(`{"foo":150, "baz":"qux"}`), true, ""},

		{opts.FlagKeyGte, "foo:9.99999",
			[]byte(`{"baz":"qux", "foo":10}`), true, ""},

		{opts.FlagKeyExists, "foo",
			[]byte(`{"baz":"qux", "wibble":"wobble"}`), false,
			""},

		{opts.FlagKeyExists, "foo",
			[]byte(`{"baz":"qux", "wibble":"wobble"}`), false,
			""},

		{opts.FlagKeyEquals, "foo:bar",
			[]byte(`{"baz":"qux", "foo":"fub"}`), true,
			"Key 'foo' does not equal 'bar'"},

		{opts.FlagKeyLte, "foo:10",
			[]byte(`{"foo":"bar", "baz":"qux"}`), true,
			"Key 'foo' value is not an integer"},

		{opts.FlagKeyLte, "foo:1",
			[]byte(`{"baz":"qux", "foo":1000000000000000}`), true,
			"Key 'foo' is greater than '1'"},

		{opts.FlagKeyGte, "foo:10",
			[]byte(`{"baz":"qux", "foo":"10"}`), true,
			"Key 'foo' value is not an integer"},

		{opts.FlagKeyGte, "foo:1000",
			[]byte(`{"baz":"qux", "foo":1}`), true,
			"Key 'foo' is less than '1000'"},
	}

	for _, c := range cases {

		// Clear old settings
		Tests["keys"] = false
		JsonTests = JsonTests[:0]

		// Call the option with the param specified in the case.
		// eg. opts.FlagKeyLte("foo:100") simulatutes --lte=foo:100
		c.option(c.param)

		// Test the flag that drives the test
		expect(t, Tests["keys"], true)

		// Unmarshall JSON in test case
		var jsonMap map[string]interface{}
		err := json.Unmarshal(c.send, &jsonMap)
		check(err)

		// Run Json test
		match, err := checkJson(jsonMap, JsonTests[0])
		expect(t, c.match, match)

		// If we didn't expect a match, test the error code
		if c.errStr != "" {
			expect(t, c.errStr, err.Error())
		}

	}
}

func Test_checkJsonValue(t *testing.T) {

	var cases = []TestJsonValueCase{
		{[]byte(`{"Foo":1,"Baz":"Qux"}`),
			JsonTest{"Baz", "Qux", "equals"},
			true, ""},

		{[]byte(`{"Animal":{"Name":"Platypus", "Order":"Monotremata"}}`),
			JsonTest{"Name", "Platypus", "equals"},
			true, ""},

		{[]byte(`{"Animal":{"Mammal":{"Name":"Platypus"}}}`),
			JsonTest{"Name", "Platypus", "equals"},
			true, ""},

		{[]byte(`[{"Foo":1,"Baz":"Qux"}, {"success":true}]`),
			JsonTest{"success", "true", "equals"},
			true, ""},

		{[]byte(`[{"Foo":1,"Baz":"Qux"}, {"success":true}, {"success":false}]`),
			JsonTest{"success", "true", "equals"},
			true, ""},

		{[]byte(`[{"Foo":100,"Baz":"Qux"}]`),
			JsonTest{"Foo", 150.00, "lte"},
			true, ""},

		{[]byte(`[{"Foo":100,"Baz":"Qux"}]`),
			JsonTest{"Foo", 50.00, "lte"},
			true, "Key 'Foo' is greater than '50'"},

		{[]byte(`[{"Foo":100,"Baz":"Qux"}]`),
			JsonTest{"Foo", 50.00, "gte"},
			true, ""},

		{[]byte(`[{"Foo":100,"Baz":"Qux"}]`),
			JsonTest{"Foo", 150.00, "gte"},
			true, "Key 'Foo' is less than '150'"},

		{[]byte(`{"Wibble":"Wobble","Baz":"Qux"}`),
			JsonTest{"Foo", "Baz", "equals"},
			false, ""},

		{[]byte(`{"Foo":"Ber","Baz":"Qux"}`),
			JsonTest{"Foo", "Bar", "equals"},
			true, "Key 'Foo' does not equal 'Bar'"},

		// Null array test. Should return "Key not found"
		{[]byte(`[]`),
			JsonTest{"success", "true", "equals"},
			false, ""},

		// Null object test. Should return "Key not found"
		{[]byte(`{}`),
			JsonTest{"success", "true", "equals"},
			false, ""},
	}

	for _, c := range cases {

		var result interface{}
		err := json.Unmarshal(c.jsonBlob, &result)
		if err != nil {
			fmt.Println("error unzipping json:", err)
		}

		match, err := checkJson(result, c.test)
		expect(t, c.match, match)

		// If we expect didn't expect a match test the error code
		if c.errStr != "" {
			expect(t, c.errStr, err.Error())
		}

	}
}
