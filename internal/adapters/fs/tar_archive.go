package fs

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type TarArchive struct{}

func (TarArchive) WriteArchive(ctx context.Context, outPath string, files []ports.ArchiveFile) error {
	if outPath == "" {
		return errors.New("archive output path is required")
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	tw := tar.NewWriter(out)
	if err := writeTarFiles(ctx, tw, files); err != nil {
		_ = tw.Close()
		_ = out.Close()
		_ = os.Remove(outPath) // don't leave a partial archive behind
		return err
	}
	if err := tw.Close(); err != nil {
		_ = out.Close()
		_ = os.Remove(outPath)
		return err
	}
	return out.Close()
}

func writeTarFiles(ctx context.Context, tw *tar.Writer, files []ports.ArchiveFile) error {
	files = sortedArchiveFiles(files)
	seen := map[string]bool{}
	for _, f := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		name, err := cleanArchivePath(f.Path)
		if err != nil {
			return err
		}
		if seen[name] {
			return fmt.Errorf("duplicate archive path %q", name)
		}
		seen[name] = true
		h := &tar.Header{Name: name, Mode: 0o600, Size: int64(len(f.Data)), ModTime: archiveModTime, Format: tar.FormatUSTAR}
		if err := tw.WriteHeader(h); err != nil {
			return err
		}
		if _, err := tw.Write(f.Data); err != nil {
			return err
		}
	}
	return nil
}

func (TarArchive) ReadArchive(ctx context.Context, inPath string) ([]ports.ArchiveFile, error) {
	in, err := os.Open(inPath)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	tr := tar.NewReader(in)
	var out []ports.ArchiveFile
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeDir {
			continue
		}
		name, err := cleanArchivePath(h.Name)
		if err != nil {
			return nil, err
		}
		// h.Size can be spoofed, so cap the actual read as well.
		data, err := io.ReadAll(io.LimitReader(tr, maxArchiveEntrySize+1))
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > maxArchiveEntrySize {
			return nil, fmt.Errorf("archive entry %q exceeds size limit (%d bytes)", name, maxArchiveEntrySize)
		}
		out = append(out, ports.ArchiveFile{Path: name, Data: data})
	}
	return sortedArchiveFiles(out), nil
}
