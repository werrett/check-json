package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

/*
 * Data models to hold test cases for various checks
 */

type TestTemplateTestCase struct {
	text  string
	match bool
}

type TestTemplateCmdCase struct {
	text   string
	expect string
}

/*
 * Tests for primary functions
 */

func Test_templateTest(t *testing.T) {

	cases := []TestTemplateTestCase{
		{"{{ isotime 2006/01/02 }}", true},
		{"wibble {{ isotime 2006/01/02 }} and wobble", true},
		{"{{ env HOME }}", true},
		{"{{env HOME}}", true},
		{"{{ isotime 2006/01/02/15 ", false},
		{"{{ isotime 2006/01/02/15 }", false},
		{"}}", false},
	}

	for _, c := range cases {

		// Look for template command
		match := templateTest(c.text)
		expect(t, c.match, match)

	}
}

func Test_templateCmd(t *testing.T) {

	time := time.Now()
	format := "2006/01/02"

	envKey := "CHECKJSONENV"
	envVal := "not_hazardous"

	err := os.Setenv(envKey, envVal)
	check(err)

	cases := []TestTemplateCmdCase{
		{fmt.Sprintf("isotime %s", format),
			fmt.Sprintf("%s", time.Format(format))},
		{fmt.Sprintf("env %s }}", envKey),
			fmt.Sprintf("%s", envVal)},
		{"noppy nop",
			"noppy nop"},
	}

	for _, c := range cases {

		// Do substitutions of template commands
		result := templateCmd(c.text)
		expect(t, c.expect, result)

	}

	err = os.Setenv(envKey, "")
	check(err)
}
