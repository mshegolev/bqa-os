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
// backslashes, double quotes, and control whitespace (newline/carriage-return/
// tab) so an embedded newline never breaks out of the scalar, then wraps the
// result in double quotes. The escapes match what unquoteYAMLScalar reverses.
func QuoteYAML(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\t", "\\t")
	return "\"" + value + "\""
}
