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

type SessionSource struct {
	Roots []Root
}

type Root struct {
	Source string
	Path   string
}

func (s SessionSource) Discover(ctx context.Context) ([]ports.SessionRef, error) {
	var refs []ports.SessionRef
	for _, root := range s.Roots {
		select {
		case <-ctx.Done():
			return refs, ctx.Err()
		default:
		}
		if info, err := os.Stat(root.Path); err != nil || !info.IsDir() {
			continue
		}
		_ = filepath.WalkDir(root.Path, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				switch d.Name() {
				case ".git", "node_modules", "cache", "Cache":
					return filepath.SkipDir
				}
				return nil
			}
			if !looksUseful(path) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			refs = append(refs, ports.SessionRef{
				Source:   root.Source,
				Path:     path,
				Size:     info.Size(),
				Modified: info.ModTime().UTC().Format(time.RFC3339),
			})
			return nil
		})
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].Modified > refs[j].Modified })
	return refs, nil
}

func (s SessionSource) Read(ctx context.Context, ref ports.SessionRef) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return os.ReadFile(ref.Path)
}

func looksUseful(path string) bool {
	lower := strings.ToLower(path)
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".json" && ext != ".jsonl" && ext != ".md" && ext != ".txt" && ext != ".log" {
		return false
	}
	for _, key := range []string{"session", "conversation", "transcript", "chat", "history", "messages", "projects", "logs"} {
		if strings.Contains(lower, key) {
			return true
		}
	}
	return false
}
