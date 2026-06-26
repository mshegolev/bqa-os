package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type Result struct {
	Discovered int
	Ingested   int
}

type UseCase struct {
	Source ports.SessionSource
	Store  ports.SessionStore
	Now    func() time.Time
}

func (u UseCase) Run(ctx context.Context) (Result, error) {
	if u.Now == nil {
		u.Now = time.Now
	}
	refs, err := u.Source.Discover(ctx)
	if err != nil {
		return Result{}, err
	}

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
		normalized := ports.NormalizedSession{
			Ref:     ref,
			RawPath: rawPath,
			Content: normalize(ref, rawPath, hash, data),
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
		return Result{}, err
	}
	return Result{Discovered: len(refs), Ingested: len(index.Entries)}, nil
}

func sha(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func normalize(ref ports.SessionRef, rawPath string, hash string, data []byte) string {
	body := string(data)
	body = strings.ReplaceAll(body, "\x00", "")
	if len(body) > 300000 {
		body = body[:300000] + "\n\n[truncated by BQA]\n"
	}
	return fmt.Sprintf("# BQA Normalized Session\n\nSource: %s\nOriginal path: %s\nRaw copy: %s\nModified: %s\nSize: %d\nSHA256: %s\n\n---\n\n%s\n", ref.Source, ref.Path, rawPath, ref.Modified, ref.Size, hash, body)
}
