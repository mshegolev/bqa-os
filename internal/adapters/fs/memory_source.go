package fs

import (
	"context"
	"fmt"
	"os"
	archivepath "path"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/pathsafe"
	"github.com/mshegolev/bqa-os/internal/ports"
)

type MemorySource struct{}

func (MemorySource) ReadMemory(ctx context.Context, root string, allow []string) ([]ports.ArchiveFile, error) {
	var out []ports.ArchiveFile
	for _, dir := range allow {
		base := filepath.Join(root, dir)
		info, err := os.Stat(base)
		if err != nil || !info.IsDir() {
			continue // a missing allow-listed dir is simply absent
		}
		err = filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if d.IsDir() || !d.Type().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			if _, ok := pathsafe.RelClean(rel); !ok {
				return fmt.Errorf("unsafe memory path %q", rel)
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			out = append(out, ports.ArchiveFile{Path: archivepath.Clean(filepath.ToSlash(rel)), Data: data})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return sortedArchiveFiles(out), nil
}
