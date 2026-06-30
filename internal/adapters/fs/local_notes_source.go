package fs

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// LocalNotesSource discovers local ETL note and log files under a single
// directory. Unlike SessionSource it selects files purely by extension
// (.md/.log/.txt by default), so arbitrarily named notes such as etl_1.md or
// etl_1.log are picked up. It implements ports.SessionSource.
type LocalNotesSource struct {
	// Dir is the directory to scan (e.g. ./etl-notes).
	Dir string
	// Source labels every discovered ref; it becomes the index entry "source".
	Source string
	// Exts is the set of accepted lowercase extensions (with leading dot). When
	// empty, defaults to .md, .log and .txt.
	Exts []string
}

func (s LocalNotesSource) acceptedExts() map[string]bool {
	exts := s.Exts
	if len(exts) == 0 {
		exts = []string{".md", ".log", ".txt"}
	}
	set := make(map[string]bool, len(exts))
	for _, ext := range exts {
		set[strings.ToLower(ext)] = true
	}
	return set
}

func (s LocalNotesSource) Discover(ctx context.Context) ([]ports.SessionRef, error) {
	source := s.Source
	if source == "" {
		source = "local"
	}
	accepted := s.acceptedExts()

	if info, err := os.Stat(s.Dir); err != nil || !info.IsDir() {
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	var refs []ports.SessionRef
	err := filepath.WalkDir(s.Dir, func(path string, d os.DirEntry, walkErr error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", ".venv", "venv":
				return filepath.SkipDir
			}
			return nil
		}
		if !accepted[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		refs = append(refs, ports.SessionRef{
			Source:   source,
			Path:     path,
			Size:     info.Size(),
			Modified: info.ModTime().UTC().Format(time.RFC3339),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Deterministic order: by path, so output is stable across runs.
	sort.Slice(refs, func(i, j int) bool { return refs[i].Path < refs[j].Path })
	return refs, nil
}

func (s LocalNotesSource) Read(ctx context.Context, ref ports.SessionRef) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return os.ReadFile(ref.Path)
}
