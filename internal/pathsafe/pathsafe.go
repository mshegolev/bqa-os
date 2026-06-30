// Package pathsafe validates untrusted relative paths before they are joined to
// a trusted base directory.
package pathsafe

import (
	"os"
	"path/filepath"
	"strings"
)

// RelClean cleans path and reports whether it is a safe relative path: not
// absolute, not "..", and not escaping its base via a ".." prefix. The returned
// string is the cleaned path; callers should join it to their base directory.
// ok is false for unsafe input, leaving the caller to format a context-specific
// error.
func RelClean(path string) (cleaned string, ok bool) {
	cleaned = filepath.Clean(path)
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", false
	}
	return cleaned, true
}
