package ingest

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestImportLocalNormalizesNotesAndWritesIndex(t *testing.T) {
	notesDir := t.TempDir()
	baseDir := t.TempDir()

	// Synthetic ETL notes/logs only — no private data.
	writeNote(t, notesDir, "etl_1.md", "# ETL 1\nPipeline netsuite -> warehouse.\nObserved failure: null account_id.\n")
	writeNote(t, notesDir, "etl_2.log", "2026-06-30 INFO loaded 1000 rows\n2026-06-30 ERROR duplicate key\n")
	writeNote(t, notesDir, "etl_3.txt", "manual check: row counts match source\n")
	// A secret that must be redacted out of the normalized artifact.
	writeNote(t, notesDir, "etl_4.md", "config: password=supersecret123\nrun ok\n")
	// A file with an ignored extension must not be imported.
	writeNote(t, notesDir, "ignore.bin", "binary-ish\n")

	fixedTime := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	uc := ImportLocal{
		Source: fsadapter.LocalNotesSource{Dir: notesDir, Source: "local-etl"},
		Store:  &fsadapter.SessionStore{BaseDir: baseDir},
		Now:    func() time.Time { return fixedTime },
	}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Discovered != 4 {
		t.Fatalf("Discovered = %d, want 4 (only .md/.log/.txt)", result.Discovered)
	}
	if result.Imported != 4 {
		t.Fatalf("Imported = %d, want 4", result.Imported)
	}
	if result.Redactions < 1 {
		t.Fatalf("Redactions = %d, want >= 1 (password should be redacted)", result.Redactions)
	}

	// index.json must exist and be valid.
	indexPath := filepath.Join(baseDir, "index.json")
	raw, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read index.json: %v", err)
	}
	var index ports.SessionIndex
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatalf("index.json is not valid JSON: %v", err)
	}
	if index.GeneratedAt != fixedTime.Format(time.RFC3339) {
		t.Fatalf("GeneratedAt = %q, want %q", index.GeneratedAt, fixedTime.Format(time.RFC3339))
	}
	if len(index.Entries) != 4 {
		t.Fatalf("index entries = %d, want 4", len(index.Entries))
	}

	for _, entry := range index.Entries {
		if entry.Source != "local-etl" {
			t.Errorf("entry source = %q, want local-etl", entry.Source)
		}
		if entry.OriginalPath == "" {
			t.Errorf("entry missing original_path")
		}
		if entry.NormalizedPath == "" {
			t.Errorf("entry missing normalized_path")
		}
		if entry.RawPath == "" {
			t.Errorf("entry missing raw_path")
		}
		if entry.Size <= 0 {
			t.Errorf("entry size = %d, want > 0", entry.Size)
		}
		if len(entry.SHA256) != 64 {
			t.Errorf("entry sha256 = %q, want 64 hex chars", entry.SHA256)
		}
		if entry.Modified == "" {
			t.Errorf("entry missing modified")
		}

		// Normalized file must exist on disk (path is relative to base dir).
		normPath := filepath.Join(baseDir, entry.NormalizedPath)
		content, err := os.ReadFile(normPath)
		if err != nil {
			t.Fatalf("normalized file %s missing: %v", entry.NormalizedPath, err)
		}
		text := string(content)
		if !strings.Contains(text, "# BQA Normalized Session") {
			t.Errorf("normalized %s missing BQA header", entry.NormalizedPath)
		}
		if !strings.Contains(text, "## QA signals") {
			t.Errorf("normalized %s missing QA signals section", entry.NormalizedPath)
		}
		if !strings.Contains(text, "## Raw note") {
			t.Errorf("normalized %s missing Raw note section", entry.NormalizedPath)
		}
		// No live secret survives into the normalized artifact.
		if strings.Contains(text, "supersecret123") {
			t.Errorf("normalized %s leaked a secret", entry.NormalizedPath)
		}
	}
}

func TestImportLocalEmptyDir(t *testing.T) {
	notesDir := t.TempDir()
	baseDir := t.TempDir()

	uc := ImportLocal{
		Source: fsadapter.LocalNotesSource{Dir: notesDir, Source: "local-etl"},
		Store:  &fsadapter.SessionStore{BaseDir: baseDir},
	}
	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Discovered != 0 || result.Imported != 0 {
		t.Fatalf("expected 0/0, got %d/%d", result.Discovered, result.Imported)
	}
	// An index.json must still be written (valid, empty entries) for bqa build.
	raw, err := os.ReadFile(filepath.Join(baseDir, "index.json"))
	if err != nil {
		t.Fatalf("read index.json: %v", err)
	}
	var index ports.SessionIndex
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatalf("index.json invalid: %v", err)
	}
	if len(index.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(index.Entries))
	}
}

func writeNote(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write note %s: %v", name, err)
	}
}
