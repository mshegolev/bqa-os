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
// result in double quotes. The escapes match what UnquoteYAML reverses.
func QuoteYAML(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\t", "\\t")
	return "\"" + value + "\""
}

// UnquoteYAML reverses QuoteYAML: it trims surrounding whitespace and, when the
// value is a double-quoted scalar, unescapes the exact set of escapes QuoteYAML
// produces (\\, \", \r, \n, \t). Bare (unquoted) scalars are returned trimmed.
func UnquoteYAML(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		inner := value[1 : len(value)-1]
		r := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\"`, "\"", `\\`, "\\")
		return r.Replace(inner)
	}
	return value
}
