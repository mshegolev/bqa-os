package memory

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type ImportOptions struct {
	FromPath string
	Target   string
	DryRun   bool
}

type ImportResult struct {
	Verified  bool
	Installed ports.InstalledSummary
	DryRun    bool
}

func (u UseCase) Import(ctx context.Context, opts ImportOptions) (ImportResult, error) {
	reader, err := u.readerFor(opts.FromPath)
	if err != nil {
		return ImportResult{}, err
	}
	files, err := reader.ReadArchive(ctx, opts.FromPath)
	if err != nil {
		return ImportResult{}, err
	}
	index := map[string][]byte{}
	for _, f := range files {
		index[f.Path] = f.Data
	}

	mdata, ok := index["manifest.json"]
	if !ok {
		return ImportResult{}, errors.New("bundle is missing manifest.json")
	}
	manifest, err := parseManifest(mdata)
	if err != nil {
		return ImportResult{}, err
	}
	if manifest.BundleVersion != BundleVersion {
		return ImportResult{}, fmt.Errorf("unsupported bundle_version %d (want %d)", manifest.BundleVersion, BundleVersion)
	}

	cdata, ok := index["metadata/checksums.json"]
	if !ok {
		return ImportResult{}, errors.New("bundle is missing metadata/checksums.json")
	}
	checksums, err := parseChecksums(cdata)
	if err != nil {
		return ImportResult{}, err
	}

	payload := payloadFiles(files)
	if err := verifyChecksums(payload, checksums); err != nil {
		return ImportResult{}, err
	}

	res := ImportResult{Verified: true, DryRun: opts.DryRun}
	if opts.DryRun {
		return res, nil
	}

	stage, err := os.MkdirTemp("", "bqa-import-*")
	if err != nil {
		return ImportResult{}, err
	}
	defer os.RemoveAll(stage)
	if err := writeFilesToDir(stage, payload); err != nil {
		return ImportResult{}, err
	}
	summary, err := u.Installer.InstallMemory(ctx, stage, opts.Target)
	if err != nil {
		return ImportResult{}, err
	}
	res.Installed = summary
	return res, nil
}

// readerFor selects the archive reader by file extension.
func (u UseCase) readerFor(fromPath string) (ports.ArchiveReader, error) {
	switch {
	case strings.HasSuffix(fromPath, ".zip"):
		if r, ok := u.Readers["zip"]; ok {
			return r, nil
		}
	case strings.HasSuffix(fromPath, ".tar"):
		if r, ok := u.Readers["tar"]; ok {
			return r, nil
		}
	}
	return nil, fmt.Errorf("unsupported bundle format for %q (want .zip or .tar)", fromPath)
}

// payloadFiles returns only the allow-listed payload (drops manifest, metadata, docs).
func payloadFiles(files []ports.ArchiveFile) []ports.ArchiveFile {
	allow := map[string]bool{}
	for _, d := range AllowList {
		allow[d] = true
	}
	var out []ports.ArchiveFile
	for _, f := range files {
		i := strings.IndexByte(f.Path, '/')
		if i <= 0 {
			continue // manifest.json, README.md, etc.
		}
		if allow[f.Path[:i]] {
			out = append(out, f)
		}
	}
	return out
}
