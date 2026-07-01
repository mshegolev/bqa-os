package fs

import (
	"fmt"
	archivepath "path"
	"sort"
	"strings"
	"time"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// archiveModTime is a fixed timestamp so archives are byte-deterministic.
var archiveModTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// maxArchiveEntrySize caps a single decompressed archive entry so a malicious
// bundle (e.g. a zip bomb) can't exhaust memory during import.
const maxArchiveEntrySize = 64 << 20 // 64 MiB

func sortedArchiveFiles(files []ports.ArchiveFile) []ports.ArchiveFile {
	out := append([]ports.ArchiveFile(nil), files...)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func cleanArchivePath(value string) (string, error) {
	value = strings.ReplaceAll(value, "\\", "/")
	cleaned := archivepath.Clean(value)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("invalid archive path %q", value)
	}
	return cleaned, nil
}
