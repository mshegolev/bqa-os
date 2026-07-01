package fs

import (
	"context"

	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/mshegolev/bqa-os/internal/sanitize"
)

type MemoryAuditor struct{}

// Audit runs a non-destructive sanitize scan; Candidates is the number of files
// that contain redaction candidates (potential secrets/PII).
func (MemoryAuditor) Audit(ctx context.Context, dir string) (ports.AuditReport, error) {
	res, err := sanitize.Path(dir, false)
	if err != nil {
		return ports.AuditReport{}, err
	}
	return ports.AuditReport{FilesScanned: res.FilesScanned, Candidates: res.FilesChanged}, nil
}
