package brain

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mshegolev/bqa-os/internal/pathsafe"
)

// installArtifactDirs lists the top-level brain artifact directories that are
// safe to copy into a client project. Anything not in this list (raw sessions,
// secrets, caches, VCS metadata) is deliberately ignored.
var installArtifactDirs = []string{
	"knowledge",
	"agents",
	"skills",
	"workflows",
	"prompts",
	"registry",
}

// installRequiredDirs are the directories the source package must contain to be
// considered a valid brain export. The registry is mandatory and at least one
// artifact directory must be present.
var installRequiredDirs = []string{"registry"}

// InstallResult summarizes a completed install for reporting.
type InstallResult struct {
	Source      string
	Target      string
	BqaDir      string
	Directories []string
	Files       []string
}

// Install copies a generated brain package from source into <target>/.bqa/.
//
// It validates that source is an existing directory holding the expected brain
// structure and that target is an existing directory, then copies only the
// known-safe artifact directories. Unrelated files in target are never touched.
func Install(source, target string) (*InstallResult, error) {
	source = strings.TrimSpace(source)
	target = strings.TrimSpace(target)
	if source == "" {
		return nil, fmt.Errorf("--from source brain package directory is required")
	}
	if target == "" {
		return nil, fmt.Errorf("--target client project directory is required")
	}

	if err := mustBeDir(source, "source brain package"); err != nil {
		return nil, err
	}
	if err := mustBeDir(target, "target project"); err != nil {
		return nil, err
	}

	if err := validateBrainPackage(source); err != nil {
		return nil, err
	}

	bqaDir := filepath.Join(target, ".bqa")
	if err := os.MkdirAll(bqaDir, 0o755); err != nil {
		return nil, err
	}

	result := &InstallResult{Source: source, Target: target, BqaDir: bqaDir}

	for _, name := range installArtifactDirs {
		src := filepath.Join(source, name)
		info, err := os.Stat(src)
		if err != nil || !info.IsDir() {
			continue
		}
		dst := filepath.Join(bqaDir, name)
		files, err := copyTree(src, dst)
		if err != nil {
			return nil, fmt.Errorf("install %s: %w", name, err)
		}
		result.Directories = append(result.Directories, name)
		result.Files = append(result.Files, files...)
	}

	sort.Strings(result.Files)
	return result, nil
}

// validateBrainPackage ensures source looks like a brain export: it must contain
// the required directories and at least one artifact directory with content.
func validateBrainPackage(source string) error {
	for _, name := range installRequiredDirs {
		info, err := os.Stat(filepath.Join(source, name))
		if err != nil || !info.IsDir() {
			return fmt.Errorf("invalid brain package: missing required %q directory in %s", name, source)
		}
	}

	for _, name := range installArtifactDirs {
		info, err := os.Stat(filepath.Join(source, name))
		if err == nil && info.IsDir() {
			return nil
		}
	}
	return fmt.Errorf("invalid brain package: no artifact directories found in %s", source)
}

func mustBeDir(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist: %s", label, path)
		}
		return fmt.Errorf("%s: %w", label, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory: %s", label, path)
	}
	return nil
}

// copyTree recursively copies regular files from src into dst, returning the
// relative paths (under the destination's parent .bqa directory style) that were
// written. Symlinks and other non-regular files are skipped for safety, and
// every relative path is validated with pathsafe before being joined.
func copyTree(src, dst string) ([]string, error) {
	base := filepath.Base(dst)
	var written []string

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			if info.IsDir() {
				return os.MkdirAll(dst, 0o755)
			}
			return nil
		}

		cleaned, ok := pathsafe.RelClean(rel)
		if !ok {
			return fmt.Errorf("unsafe path in source package: %s", rel)
		}

		if info.IsDir() {
			return os.MkdirAll(filepath.Join(dst, cleaned), 0o755)
		}
		if !info.Mode().IsRegular() {
			// Skip symlinks, sockets, devices: never follow them into a target repo.
			return nil
		}

		dstFile := filepath.Join(dst, cleaned)
		if err := os.MkdirAll(filepath.Dir(dstFile), 0o755); err != nil {
			return err
		}
		if err := copyFile(path, dstFile); err != nil {
			return err
		}
		written = append(written, filepath.ToSlash(filepath.Join(base, cleaned)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return written, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
