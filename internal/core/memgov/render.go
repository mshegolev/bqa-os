package memgov

import (
	"fmt"
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
)

// header renders the shared v1 envelope.
func header(kind string) string {
	return fmt.Sprintf("schema_version: %d\nkind: %s\ngenerated_by: %s\n", SchemaVersion, kind, generatedBy())
}

// renderItems renders a kind's items file. id and status are barewords (safe by
// construction); free-text fields are YAML-quoted.
func renderItems(kind string, items []MemoryItem) string {
	var b strings.Builder
	b.WriteString(header(kind))
	if len(items) == 0 {
		b.WriteString("items: []\n")
		return b.String()
	}
	b.WriteString("items:\n")
	for _, it := range items {
		b.WriteString("  - id: " + it.ID + "\n")
		b.WriteString("    name: " + textutil.QuoteYAML(it.Name) + "\n")
		b.WriteString("    domain: " + textutil.QuoteYAML(it.Domain) + "\n")
		b.WriteString("    evidence: " + textutil.QuoteYAML(it.Evidence) + "\n")
		b.WriteString("    source: " + textutil.QuoteYAML(it.Source) + "\n")
		b.WriteString("    status: " + it.Status + "\n")
	}
	return b.String()
}

// renderLog renders the decision_log file.
func renderLog(entries []DecisionEntry) string {
	var b strings.Builder
	b.WriteString(header(KindDecisionLog))
	if len(entries) == 0 {
		b.WriteString("entries: []\n")
		return b.String()
	}
	b.WriteString("entries:\n")
	for _, e := range entries {
		b.WriteString("  - id: " + e.ID + "\n")
		b.WriteString("    action: " + e.Action + "\n")
		b.WriteString("    name: " + textutil.QuoteYAML(e.Name) + "\n")
	}
	return b.String()
}
