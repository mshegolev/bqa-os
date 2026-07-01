package ports

import "context"

// ArchiveFile is one file in a memory bundle. Path is bundle-relative with
// forward slashes.
type ArchiveFile struct {
	Path string
	Data []byte
}

type ArchiveWriter interface {
	WriteArchive(ctx context.Context, outPath string, files []ArchiveFile) error
}

type ArchiveReader interface {
	ReadArchive(ctx context.Context, inPath string) ([]ArchiveFile, error)
}
