package fs

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"os"
	archivepath "path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type DemoArchiveWriter struct{}

func (w DemoArchiveWriter) WriteDemoArchive(ctx context.Context, outputPath string, files []ports.DemoArchiveFile) error {
	if outputPath == "" {
		return errors.New("demo archive output path is required")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(out)
	if err := writeDemoArchiveFiles(ctx, zipWriter, files); err != nil {
		_ = zipWriter.Close()
		_ = out.Close()
		return err
	}
	if err := zipWriter.Close(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func writeDemoArchiveFiles(ctx context.Context, zipWriter *zip.Writer, files []ports.DemoArchiveFile) error {
	files = append([]ports.DemoArchiveFile(nil), files...)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	seen := map[string]bool{}
	for _, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		name, err := cleanDemoArchivePath(file.Path)
		if err != nil {
			return err
		}
		if seen[name] {
			return fmt.Errorf("duplicate demo archive path %q", name)
		}
		seen[name] = true

		header := &zip.FileHeader{Name: name, Method: zip.Store}
		header.SetModTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
		header.SetMode(0o600)

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := writer.Write([]byte(file.Content)); err != nil {
			return err
		}
	}
	return nil
}

func cleanDemoArchivePath(value string) (string, error) {
	value = strings.ReplaceAll(value, "\\", "/")
	cleaned := archivepath.Clean(value)
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("invalid demo archive path %q", value)
	}
	return cleaned, nil
}
