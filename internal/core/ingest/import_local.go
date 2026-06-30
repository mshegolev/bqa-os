package ingest

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/mshegolev/bqa-os/internal/sanitize"
)

// ImportLocal converts local ETL notes/log snippets discovered by Source into
// normalized BQA session markdown files and a canonical index.json, using the
// same SessionStore and SessionIndexEntry shape that `bqa ingest`/`bqa build`
// rely on. It differs from UseCase only in normalization: it redacts secrets
// in-memory and structures each note into QA-signal sections.
type ImportLocal struct {
	Source ports.SessionSource
	Store  ports.SessionStore
	Now    func() time.Time
}

// ImportResult reports how many files were discovered, imported, and how many
// secret redactions were applied across all imported files.
type ImportResult struct {
	Discovered int
	Imported   int
	Redactions int
}

func (u ImportLocal) Run(ctx context.Context) (ImportResult, error) {
	if u.Now == nil {
		u.Now = time.Now
	}
	refs, err := u.Source.Discover(ctx)
	if err != nil {
		return ImportResult{}, err
	}

	result := ImportResult{Discovered: len(refs)}
	index := ports.SessionIndex{GeneratedAt: u.Now().UTC().Format(time.RFC3339)}
	for _, ref := range refs {
		data, err := u.Source.Read(ctx, ref)
		if err != nil {
			continue
		}
		hash := sha(data)
		raw := ports.RawSession{Ref: ref, Bytes: data, SHA256: hash}
		rawPath, err := u.Store.SaveRaw(ctx, raw)
		if err != nil {
			continue
		}
		content, redactions := normalizeLocalNote(ref, rawPath, hash, data)
		result.Redactions += redactions
		normalized := ports.NormalizedSession{
			Ref:     ref,
			RawPath: rawPath,
			Content: content,
			SHA256:  hash,
		}
		normalizedPath, err := u.Store.SaveNormalized(ctx, normalized)
		if err != nil {
			continue
		}
		index.Entries = append(index.Entries, ports.SessionIndexEntry{
			Source:         ref.Source,
			OriginalPath:   ref.Path,
			RawPath:        rawPath,
			NormalizedPath: normalizedPath,
			Size:           ref.Size,
			SHA256:         hash,
			Modified:       ref.Modified,
		})
	}
	if err := u.Store.SaveIndex(ctx, index); err != nil {
		return ImportResult{}, err
	}
	result.Imported = len(index.Entries)
	return result, nil
}

// qaSignalSections is the ordered set of QA-signal sections every normalized
// ETL note exposes, matching issue #47's "useful QA signals" list. The raw note
// body is preserved verbatim (after secret redaction) under "Raw note".
var qaSignalSections = []string{
	"Pipeline / source / target summary",
	"Observed failures",
	"Successful checks",
	"Common bugs",
	"Useful prompts",
	"Project knowledge",
	"Data quality risks",
}

func normalizeLocalNote(ref ports.SessionRef, rawPath string, hash string, data []byte) (string, int) {
	body := strings.ReplaceAll(string(data), "\x00", "")
	if len(body) > 300000 {
		body = body[:300000] + "\n\n[truncated by BQA]\n"
	}
	// Redact secrets before they reach the normalized (committed) artifact.
	body, redactions := sanitize.Text(body)

	title := strings.TrimSuffix(filepath.Base(ref.Path), filepath.Ext(ref.Path))

	var b strings.Builder
	fmt.Fprintf(&b, "# BQA Normalized Session: %s\n\n", title)
	fmt.Fprintf(&b, "Source: %s\nOriginal path: %s\nRaw copy: %s\nModified: %s\nSize: %d\nSHA256: %s\n\n",
		ref.Source, ref.Path, rawPath, ref.Modified, ref.Size, hash)
	b.WriteString("> Imported from a local ETL note via `bqa ingest --from`. ")
	b.WriteString("Sanitize any client data before committing the `.bqa/` directory.\n\n")

	b.WriteString("## QA signals\n\n")
	for _, section := range qaSignalSections {
		fmt.Fprintf(&b, "### %s\n\n_(extract from the raw note below)_\n\n", section)
	}

	b.WriteString("## Raw note\n\n")
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	return b.String(), redactions
}
