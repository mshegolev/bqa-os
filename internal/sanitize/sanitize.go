package sanitize

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Result struct {
	FilesScanned int
	FilesChanged int
	Redactions   int
}

var patterns = []struct {
	Name string
	Re   *regexp.Regexp
}{
	{"aws_access_key", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"generic_token", regexp.MustCompile(`(?i)(token|api_key|apikey|secret|password)\s*[:=]\s*['\"]?[^'\"\s]+`)},
	{"private_key", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`)},
	{"email", regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)},
}

func Path(root string, write bool) (Result, error) {
	var result Result
	root = filepath.Clean(root)

	info, err := os.Stat(root)
	if err != nil {
		return result, err
	}
	if !info.IsDir() {
		return sanitizeFile(root, write)
	}

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == "node_modules" || base == ".venv" || base == "venv" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isTextCandidate(path) {
			return nil
		}
		fileResult, err := sanitizeFile(path, write)
		if err != nil {
			return err
		}
		result.FilesScanned += fileResult.FilesScanned
		result.FilesChanged += fileResult.FilesChanged
		result.Redactions += fileResult.Redactions
		return nil
	})
	return result, err
}

// Text redacts known secret patterns (AWS keys, generic tokens, private keys,
// emails) from an in-memory string. It returns the redacted text and the number
// of redactions applied, leaving the input untouched. Use this when content is
// already in memory and must not be written back to disk.
func Text(content string) (redacted string, redactions int) {
	redacted = content
	for _, pattern := range patterns {
		matches := pattern.Re.FindAllString(redacted, -1)
		if len(matches) == 0 {
			continue
		}
		redactions += len(matches)
		redacted = pattern.Re.ReplaceAllString(redacted, fmt.Sprintf("[REDACTED_%s]", strings.ToUpper(pattern.Name)))
	}
	return redacted, redactions
}

func sanitizeFile(path string, write bool) (Result, error) {
	var result Result
	data, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	if looksBinary(data) {
		return result, nil
	}
	result.FilesScanned = 1
	original := string(data)
	updated, redactions := Text(original)
	if updated != original {
		result.FilesChanged = 1
		result.Redactions = redactions
		if write {
			if err := os.WriteFile(path, []byte(updated), 0o600); err != nil {
				return result, err
			}
		}
	}
	return result, nil
}

func isTextCandidate(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md", ".txt", ".yaml", ".yml", ".json", ".toml", ".env", ".log", ".sql", ".py", ".go", ".js", ".ts", ".graphql", ".gql":
		return true
	default:
		return ext == ""
	}
}

func looksBinary(data []byte) bool {
	limit := len(data)
	if limit > 4096 {
		limit = 4096
	}
	for i := 0; i < limit; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}
