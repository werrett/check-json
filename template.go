package main

import (
	"os"
	"regexp"
	"strings"
	"time"
)

// Setup the regex used to identify {{ template text }}
var templateRegex = regexp.MustCompile(`{{ *([^{}]+) *}}`)

// Test if the passed string has {{ template }} in it
func templateTest(str string) bool {
	return templateRegex.MatchString(str)
}

// Handle {{ templates }} found in the post request body
func templateHndlr(str string) string {
	for _, match := range templateRegex.FindAllStringSubmatch(str, -1) {
		result := templateCmd(match[1])
		str = strings.Replace(str, match[0], result, 1)
	}
	return str
}

// Handle individual {{ template }} commands
func templateCmd(templateTxt string) string {
	params := strings.Split(templateTxt, " ")

	// Switch on template commands.
	switch params[0] {

	// Display date/time with given format
	case "isotime":
		t := time.Now()
		// Remove "` or space chars surroudning date format
		format := strings.Trim(strings.Join(params[1:], " "), "\"` ")
		return t.Format(format)

		// Get an environmental variable
	case "env": //
		envVar := strings.Trim(params[1], "\" ")
		return os.Getenv(envVar)

		// If no command found just return the string minus the {{}}
	default:
		return templateTxt
	}
}
