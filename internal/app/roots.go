package app

import (
	"os"
	"path/filepath"
	"strings"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
)

func sessionRoots(sources string, global bool, local bool) []fsadapter.Root {
	want := map[string]bool{}
	for _, part := range strings.Split(sources, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part != "" {
			want[part] = true
		}
	}
	if len(want) == 0 {
		want["claude"] = true
		want["codex"] = true
		want["opencode"] = true
		want["droid"] = true
	}

	var roots []fsadapter.Root
	add := func(source string, path string) {
		if !want[source] {
			return
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			roots = append(roots, fsadapter.Root{Source: source, Path: path})
		}
	}

	home, _ := os.UserHomeDir()
	if global && home != "" {
		add("claude", filepath.Join(home, ".claude"))
		add("claude", filepath.Join(home, ".config", "claude"))
		add("codex", filepath.Join(home, ".codex"))
		add("codex", filepath.Join(home, ".config", "codex"))
		add("opencode", filepath.Join(home, ".opencode"))
		add("opencode", filepath.Join(home, ".config", "opencode"))
		add("droid", filepath.Join(home, ".droid"))
	}
	if local {
		add("claude", ".claude")
		add("codex", ".codex")
		add("opencode", ".opencode")
		add("droid", ".droid")
	}
	return roots
}
