package fs

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type ZipArchive struct{}

func (ZipArchive) WriteArchive(ctx context.Context, outPath string, files []ports.ArchiveFile) error {
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
	zw := zip.NewWriter(out)
	if err := writeZipFiles(ctx, zw, files); err != nil {
		_ = zw.Close()
		_ = out.Close()
		_ = os.Remove(outPath) // don't leave a partial archive behind
		return err
	}
	if err := zw.Close(); err != nil {
		_ = out.Close()
		_ = os.Remove(outPath)
		return err
	}
	return out.Close()
}

func writeZipFiles(ctx context.Context, zw *zip.Writer, files []ports.ArchiveFile) error {
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
		h := &zip.FileHeader{Name: name, Method: zip.Store}
		h.SetModTime(archiveModTime)
		h.SetMode(0o600)
		w, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		if _, err := w.Write(f.Data); err != nil {
			return err
		}
	}
	return nil
}

func (ZipArchive) ReadArchive(ctx context.Context, inPath string) ([]ports.ArchiveFile, error) {
	r, err := zip.OpenReader(inPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []ports.ArchiveFile
	for _, zf := range r.File {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if zf.FileInfo().IsDir() {
			continue
		}
		name, err := cleanArchivePath(zf.Name)
		if err != nil {
			return nil, err
		}
		if zf.UncompressedSize64 > maxArchiveEntrySize {
			return nil, fmt.Errorf("archive entry %q exceeds size limit (%d bytes)", name, zf.UncompressedSize64)
		}
		rc, err := zf.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(rc, maxArchiveEntrySize+1))
		rc.Close()
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
