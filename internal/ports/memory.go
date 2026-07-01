package ports

import "context"

// MemorySource reads the allow-listed memory files under a root (e.g. ".bqa").
type MemorySource interface {
	ReadMemory(ctx context.Context, root string, allow []string) ([]ArchiveFile, error)
}

// InstalledSummary reports where a bundle was installed.
type InstalledSummary struct {
	Target string
	Files  []string
}

// MemoryInstaller places a verified bundle directory into <target>/.bqa.
type MemoryInstaller interface {
	InstallMemory(ctx context.Context, sourceDir, target string) (InstalledSummary, error)
}

// AuditReport summarizes a pre-export sensitivity scan.
type AuditReport struct {
	FilesScanned int
	Candidates   int
}

// MemoryAuditor scans a staged directory for secret/PII candidates.
type MemoryAuditor interface {
	Audit(ctx context.Context, dir string) (AuditReport, error)
}

// BrainStore pushes assembled memory files to the connected GitHub brain.
type BrainStore interface {
	Push(ctx context.Context, files []ArchiveFile, sanitize bool) error
}
