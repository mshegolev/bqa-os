package ports

import "context"

type DemoArchiveFile struct {
	Path    string
	Content string
}

type DemoArchiveWriter interface {
	WriteDemoArchive(ctx context.Context, outputPath string, files []DemoArchiveFile) error
}
