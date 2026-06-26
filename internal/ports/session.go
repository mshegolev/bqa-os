package ports

import "context"

type SessionRef struct {
	Source   string
	Path     string
	Size     int64
	Modified string
}

type RawSession struct {
	Ref    SessionRef
	Bytes  []byte
	SHA256 string
}

type NormalizedSession struct {
	Ref     SessionRef
	RawPath string
	Content string
	SHA256  string
}

type SessionIndexEntry struct {
	Source         string `json:"source"`
	OriginalPath   string `json:"original_path"`
	RawPath        string `json:"raw_path"`
	NormalizedPath string `json:"normalized_path"`
	Size           int64  `json:"size"`
	SHA256         string `json:"sha256"`
	Modified       string `json:"modified"`
}

type SessionIndex struct {
	GeneratedAt string              `json:"generated_at"`
	Entries     []SessionIndexEntry `json:"entries"`
}

type SessionSource interface {
	Discover(ctx context.Context) ([]SessionRef, error)
	Read(ctx context.Context, ref SessionRef) ([]byte, error)
}

type SessionStore interface {
	SaveRaw(ctx context.Context, session RawSession) (string, error)
	SaveNormalized(ctx context.Context, session NormalizedSession) (string, error)
	SaveIndex(ctx context.Context, index SessionIndex) error
}
