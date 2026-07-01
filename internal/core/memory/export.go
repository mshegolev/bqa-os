package memory

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type ExportOptions struct {
	SourceRoot string   // default ".bqa"
	Target     string   // "zip" | "tar" | "github"
	OutPath    string   // required for zip/tar
	Exclude    []string // extra glob patterns (matched against bundle paths)
	DryRun     bool
	Strict     bool
}

type ExportResult struct {
	Target  string
	OutPath string
	Files   []string
	Audit   ports.AuditReport
	DryRun  bool
}

func (u UseCase) Export(ctx context.Context, opts ExportOptions) (ExportResult, error) {
	root := opts.SourceRoot
	if root == "" {
		root = ".bqa"
	}
	payload, err := u.Source.ReadMemory(ctx, root, AllowList)
	if err != nil {
		return ExportResult{}, err
	}
	payload, err = applyExcludes(payload, opts.Exclude)
	if err != nil {
		return ExportResult{}, err
	}
	if len(payload) == 0 {
		return ExportResult{}, fmt.Errorf("nothing to export from %s (run `bqa build` first)", root)
	}

	stage, err := os.MkdirTemp("", "bqa-export-*")
	if err != nil {
		return ExportResult{}, err
	}
	defer os.RemoveAll(stage)
	if err := writeFilesToDir(stage, payload); err != nil {
		return ExportResult{}, err
	}
	audit, err := u.Auditor.Audit(ctx, stage)
	if err != nil {
		return ExportResult{}, err
	}
	if opts.Strict && audit.Candidates > 0 {
		return ExportResult{}, fmt.Errorf("export blocked by --strict: audit flagged %d file(s) with redaction candidates", audit.Candidates)
	}

	res := ExportResult{Target: opts.Target, OutPath: opts.OutPath, Files: pathsOf(payload), Audit: audit, DryRun: opts.DryRun}
	if opts.DryRun {
		return res, nil
	}

	switch opts.Target {
	case "zip", "tar":
		if opts.OutPath == "" {
			return ExportResult{}, errors.New("--out is required for zip/tar export")
		}
		w, ok := u.Writers[opts.Target]
		if !ok {
			return ExportResult{}, fmt.Errorf("no archive writer for target %q", opts.Target)
		}
		bundle, err := assembleBundle(payload, audit)
		if err != nil {
			return ExportResult{}, err
		}
		if err := w.WriteArchive(ctx, opts.OutPath, bundle); err != nil {
			return ExportResult{}, err
		}
	case "github":
		if u.Brain == nil {
			return ExportResult{}, errors.New("github target is not configured")
		}
		if err := u.Brain.Push(ctx, payload, true); err != nil {
			return ExportResult{}, err
		}
	default:
		return ExportResult{}, fmt.Errorf("unknown target %q (want zip|tar|github)", opts.Target)
	}
	return res, nil
}

func applyExcludes(files []ports.ArchiveFile, patterns []string) ([]ports.ArchiveFile, error) {
	if len(patterns) == 0 {
		return files, nil
	}
	var out []ports.ArchiveFile
	for _, f := range files {
		drop := false
		for _, pat := range patterns {
			ok, err := path.Match(pat, f.Path)
			if err != nil {
				return nil, fmt.Errorf("invalid --exclude pattern %q: %w", pat, err)
			}
			if ok {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, f)
		}
	}
	return out, nil
}

func writeFilesToDir(dir string, files []ports.ArchiveFile) error {
	for _, f := range files {
		dst := filepath.Join(dir, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, f.Data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func pathsOf(files []ports.ArchiveFile) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.Path)
	}
	return out
}
