package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Options struct {
	Sources []string
	Global  bool
	Local   bool
}

type SessionFile struct {
	Source   string `json:"source"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

type Manifest struct {
	GeneratedAt string        `json:"generated_at"`
	Roots       []string      `json:"roots"`
	Files       []SessionFile `json:"files"`
}

func Discover(opts Options) (Manifest, error) {
	var manifest Manifest
	manifest.GeneratedAt = time.Now().UTC().Format(time.RFC3339)

	roots := candidateRoots(opts)
	manifest.Roots = roots

	for _, root := range roots {
		source := sourceFromRoot(root)
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == "node_modules" || name == "cache" || name == "Cache" {
					return filepath.SkipDir
				}
				return nil
			}
			if !looksLikeSessionFile(path) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			manifest.Files = append(manifest.Files, SessionFile{
				Source:   source,
				Path:     path,
				Size:     info.Size(),
				Modified: info.ModTime().UTC().Format(time.RFC3339),
			})
			return nil
		})
	}

	sort.Slice(manifest.Files, func(i, j int) bool {
		return manifest.Files[i].Modified > manifest.Files[j].Modified
	})
	return manifest, nil
}

func WriteManifest(manifest Manifest, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func PrintSummary(manifest Manifest) {
	counts := map[string]int{}
	for _, file := range manifest.Files {
		counts[file.Source]++
	}
	fmt.Printf("Discovered session-like files: %d\n", len(manifest.Files))
	for _, source := range []string{"claude", "codex", "opencode"} {
		fmt.Printf("%-8s %d\n", source, counts[source])
	}
}

func candidateRoots(opts Options) []string {
	sources := normalizeSources(opts.Sources)
	seen := map[string]bool{}
	var roots []string
	add := func(path string) {
		path = filepath.Clean(path)
		if seen[path] {
			return
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			seen[path] = true
			roots = append(roots, path)
		}
	}

	home, _ := os.UserHomeDir()
	if opts.Global && home != "" {
		if sources["claude"] {
			add(filepath.Join(home, ".claude"))
			add(filepath.Join(home, ".config", "claude"))
		}
		if sources["codex"] {
			add(filepath.Join(home, ".codex"))
			add(filepath.Join(home, ".config", "codex"))
		}
		if sources["opencode"] {
			add(filepath.Join(home, ".opencode"))
			add(filepath.Join(home, ".config", "opencode"))
		}
	}

	if opts.Local {
		if sources["claude"] {
			add(".claude")
		}
		if sources["codex"] {
			add(".codex")
		}
		if sources["opencode"] {
			add(".opencode")
		}
	}

	return roots
}

func normalizeSources(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(strings.ToLower(part))
			if part != "" {
				result[part] = true
			}
		}
	}
	if len(result) == 0 {
		result["claude"] = true
		result["codex"] = true
		result["opencode"] = true
	}
	return result
}

func sourceFromRoot(root string) string {
	lower := strings.ToLower(root)
	switch {
	case strings.Contains(lower, "claude"):
		return "claude"
	case strings.Contains(lower, "codex"):
		return "codex"
	case strings.Contains(lower, "opencode"):
		return "opencode"
	default:
		return "unknown"
	}
}

func looksLikeSessionFile(path string) bool {
	lower := strings.ToLower(path)
	ext := strings.ToLower(filepath.Ext(path))
	if !(ext == ".json" || ext == ".jsonl" || ext == ".md" || ext == ".txt" || ext == ".log") {
		return false
	}
	keywords := []string{"session", "conversation", "transcript", "chat", "history", "messages", "projects", "logs"}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}
