// Package doctor inspects the health of a .bqa workspace.
package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// Check is the result of a single workspace health check.
type Check struct {
	Name       string
	OK         bool
	Detail     string
	Suggestion string
}

// Report aggregates all workspace checks.
type Report struct {
	Checks []Check
	OK     bool
}

// requiredDirs mirrors the workspace layout created by `bqa init`.
var requiredDirs = []string{
	"input/sessions/raw",
	"input/sessions/normalized",
	"output",
	"registry",
	"memory",
	"agents",
	"skills",
	"workflows",
	"rules",
	"guardrails",
	"prompts",
}

// Run checks that every required workspace directory exists under baseDir and
// that the Knowledge Extractor build prerequisites are in place. None of the
// checks read private session content; they only inspect structure.
func Run(baseDir string) Report {
	if baseDir == "" {
		baseDir = ".bqa"
	}
	report := Report{OK: true}
	for _, dir := range requiredDirs {
		path := filepath.Join(baseDir, filepath.FromSlash(dir))
		info, err := os.Stat(path)
		ok := err == nil && info.IsDir()
		detail := path
		if !ok {
			detail = path + " (missing)"
			report.OK = false
		}
		report.Checks = append(report.Checks, Check{Name: dir, OK: ok, Detail: detail, Suggestion: suggestIf(!ok, "run `bqa init` to create missing directories")})
	}

	for _, c := range buildPrereqChecks(baseDir) {
		if !c.OK {
			report.OK = false
		}
		report.Checks = append(report.Checks, c)
	}
	return report
}

// buildPrereqChecks inspects the Knowledge Extractor build prerequisites.
func buildPrereqChecks(baseDir string) []Check {
	var checks []Check

	// .bqa/ workspace root exists.
	baseOK := isDir(baseDir)
	checks = append(checks, Check{
		Name:       "workspace",
		OK:         baseOK,
		Detail:     detailFor(baseDir, baseOK),
		Suggestion: suggestIf(!baseOK, "run `bqa init` to create the workspace"),
	})

	// .bqa/input/sessions/index.json exists.
	indexPath := filepath.Join(baseDir, "input", "sessions", "index.json")
	indexOK := isFile(indexPath)
	checks = append(checks, Check{
		Name:       "session-index",
		OK:         indexOK,
		Detail:     detailFor(indexPath, indexOK),
		Suggestion: suggestIf(!indexOK, "run `bqa discover` then `bqa ingest2` to collect sessions"),
	})

	// normalized session directory exists.
	normalizedDir := filepath.Join(baseDir, "input", "sessions", "normalized")
	normalizedOK := isDir(normalizedDir)
	checks = append(checks, Check{
		Name:       "normalized-sessions",
		OK:         normalizedOK,
		Detail:     detailFor(normalizedDir, normalizedOK),
		Suggestion: suggestIf(!normalizedOK, "run `bqa ingest2` to normalize discovered sessions"),
	})

	// index has at least one entry.
	entryCount, entryErr := countIndexEntries(indexPath)
	entriesOK := entryErr == nil && entryCount > 0
	entryDetail := indexPath + " has " + strconv.Itoa(entryCount) + " entries"
	if entryErr != nil {
		entryDetail = indexPath + " (unreadable index)"
	} else if entryCount == 0 {
		entryDetail = indexPath + " (no entries)"
	}
	checks = append(checks, Check{
		Name:       "session-entries",
		OK:         entriesOK,
		Detail:     entryDetail,
		Suggestion: suggestIf(!entriesOK, "run `bqa discover` and `bqa ingest2` to populate the session index"),
	})

	// knowledge directory exists (after build).
	knowledgeDir := filepath.Join(baseDir, "knowledge")
	knowledgeOK := isDir(knowledgeDir)
	checks = append(checks, Check{
		Name:       "knowledge-dir",
		OK:         knowledgeOK,
		Detail:     detailFor(knowledgeDir, knowledgeOK),
		Suggestion: suggestIf(!knowledgeOK, "run `bqa build` to generate knowledge artifacts"),
	})

	// .bqa/knowledge/ location is writable.
	writableOK, writableDetail := checkWritable(knowledgeDir)
	checks = append(checks, Check{
		Name:       "knowledge-writable",
		OK:         writableOK,
		Detail:     writableDetail,
		Suggestion: suggestIf(!writableOK, "ensure the workspace is writable before running `bqa build`"),
	})

	return checks
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func detailFor(path string, ok bool) string {
	if ok {
		return path
	}
	return path + " (missing)"
}

func suggestIf(failed bool, suggestion string) string {
	if failed {
		return suggestion
	}
	return ""
}

// countIndexEntries reads only the structural entries[] array of the index,
// never the private session content the entries point at.
func countIndexEntries(indexPath string) (int, error) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return 0, err
	}
	var index ports.SessionIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return 0, err
	}
	return len(index.Entries), nil
}

// checkWritable verifies the knowledge directory location can be written to by
// creating and removing a temporary probe file. The parent is created if the
// knowledge directory itself does not exist yet.
func checkWritable(knowledgeDir string) (bool, string) {
	// doctor is a read-only health check; it must not create the directory as a
	// side effect. Probe writability only when the directory already exists.
	if !isDir(knowledgeDir) {
		return false, knowledgeDir + " (not writable: directory does not exist — run `bqa build`)"
	}
	probe := filepath.Join(knowledgeDir, ".bqa-doctor-write-probe")
	if err := os.WriteFile(probe, []byte("probe"), 0o600); err != nil {
		return false, knowledgeDir + " (not writable: " + err.Error() + ")"
	}
	_ = os.Remove(probe)
	return true, knowledgeDir + " (writable)"
}
