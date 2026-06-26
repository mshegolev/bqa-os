package ingest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mshegolev/bqa-os/internal/discovery"
)

type Options struct {
	Sources []string
	Global  bool
	Local   bool
	BaseDir string
}

type IndexEntry struct {
	Source         string `json:"source"`
	OriginalPath   string `json:"original_path"`
	RawPath        string `json:"raw_path"`
	NormalizedPath string `json:"normalized_path"`
	Size           int64  `json:"size"`
	Sha256         string `json:"sha256"`
	Modified       string `json:"modified"`
}

type Index struct {
	GeneratedAt string       `json:"generated_at"`
	Entries     []IndexEntry `json:"entries"`
}

func Run(opts Options) error {
	if opts.BaseDir == "" {
		opts.BaseDir = ".bqa/input/sessions"
	}

	manifest, err := discovery.Discover(discovery.Options{
		Sources: opts.Sources,
		Global:  opts.Global,
		Local:   opts.Local,
	})
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(opts.BaseDir, "manifest.json")
	if err := discovery.WriteManifest(manifest, manifestPath); err != nil {
		return err
	}

	var index Index
	index.GeneratedAt = time.Now().UTC().Format(time.RFC3339)

	for i, file := range manifest.Files {
		entry, err := ingestFile(opts.BaseDir, i, file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", file.Path, err)
			continue
		}
		index.Entries = append(index.Entries, entry)
	}

	indexPath := filepath.Join(opts.BaseDir, "index.json")
	if err := writeJSON(indexPath, index); err != nil {
		return err
	}

	discovery.PrintSummary(manifest)
	fmt.Printf("Manifest: %s\n", manifestPath)
	fmt.Printf("Index: %s\n", indexPath)
	fmt.Printf("Ingested files: %d\n", len(index.Entries))
	return nil
}

func ingestFile(baseDir string, n int, file discovery.SessionFile) (IndexEntry, error) {
	data, err := os.ReadFile(file.Path)
	if err != nil {
		return IndexEntry{}, err
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	short := hash[:12]
	ext := strings.ToLower(filepath.Ext(file.Path))
	if ext == "" {
		ext = ".txt"
	}

	rawRel := filepath.Join("raw", file.Source, fmt.Sprintf("%06d-%s%s", n+1, short, ext))
	normRel := filepath.Join("normalized", file.Source, fmt.Sprintf("%06d-%s.md", n+1, short))
	rawPath := filepath.Join(baseDir, rawRel)
	normPath := filepath.Join(baseDir, normRel)

	if err := os.MkdirAll(filepath.Dir(rawPath), 0o755); err != nil {
		return IndexEntry{}, err
	}
	if err := os.MkdirAll(filepath.Dir(normPath), 0o755); err != nil {
		return IndexEntry{}, err
	}
	if err := copyFile(file.Path, rawPath); err != nil {
		return IndexEntry{}, err
	}

	normalized := normalizeToMarkdown(file, rawRel, hash, data)
	if err := os.WriteFile(normPath, []byte(normalized), 0o600); err != nil {
		return IndexEntry{}, err
	}

	return IndexEntry{
		Source:         file.Source,
		OriginalPath:   file.Path,
		RawPath:        rawPath,
		NormalizedPath: normPath,
		Size:           file.Size,
		Sha256:         hash,
		Modified:       file.Modified,
	}, nil
}

func normalizeToMarkdown(file discovery.SessionFile, rawRel string, hash string, data []byte) string {
	text := string(data)
	if looksBinary(data) {
		text = "[binary or non-text content omitted]"
	}
	text = strings.ReplaceAll(text, "\x00", "")
	if len(text) > 300000 {
		text = text[:300000] + "\n\n[truncated by BQA ingest]\n"
	}
	return fmt.Sprintf(`# BQA Normalized Session

Source: %s
Original path: %s
Raw copy: %s
Modified: %s
Size: %d
SHA256: %s

---

%s
`, file.Source, file.Path, rawRel, file.Modified, file.Size, hash, text)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func looksBinary(data []byte) bool {
	limit := len(data)
	if limit > 4096 {
		limit = 4096
	}
	for i := 0; i < limit; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}
