// Package textutil holds small, dependency-free string helpers shared across
// core use cases.
package textutil

import "strings"

// HasAny reports whether text contains any of the given needles.
func HasAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

// QuoteYAML escapes a value for use as a double-quoted YAML scalar: it escapes
// backslashes and double quotes, then wraps the result in double quotes.
func QuoteYAML(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return "\"" + value + "\""
}
