package memgov

import (
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
)

// parseItems parses an items file rendered by renderItems. Unknown lines and the
// envelope header are ignored; each "- id:" starts a new item.
func parseItems(content string) []MemoryItem {
	var items []MemoryItem
	var cur *MemoryItem
	flush := func() {
		if cur != nil {
			items = append(items, *cur)
			cur = nil
		}
	}
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "- id:"):
			flush()
			cur = &MemoryItem{ID: strings.TrimSpace(strings.TrimPrefix(line, "- id:"))}
		case cur == nil:
			continue
		case strings.HasPrefix(line, "name:"):
			cur.Name = textutil.UnquoteYAML(strings.TrimPrefix(line, "name:"))
		case strings.HasPrefix(line, "domain:"):
			cur.Domain = textutil.UnquoteYAML(strings.TrimPrefix(line, "domain:"))
		case strings.HasPrefix(line, "evidence:"):
			cur.Evidence = textutil.UnquoteYAML(strings.TrimPrefix(line, "evidence:"))
		case strings.HasPrefix(line, "source:"):
			cur.Source = textutil.UnquoteYAML(strings.TrimPrefix(line, "source:"))
		case strings.HasPrefix(line, "status:"):
			cur.Status = strings.TrimSpace(strings.TrimPrefix(line, "status:"))
		}
	}
	flush()
	return items
}

// parseLog parses a decision_log file rendered by renderLog.
func parseLog(content string) []DecisionEntry {
	var entries []DecisionEntry
	var cur *DecisionEntry
	flush := func() {
		if cur != nil {
			entries = append(entries, *cur)
			cur = nil
		}
	}
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "- id:"):
			flush()
			cur = &DecisionEntry{ID: strings.TrimSpace(strings.TrimPrefix(line, "- id:"))}
		case cur == nil:
			continue
		case strings.HasPrefix(line, "action:"):
			cur.Action = strings.TrimSpace(strings.TrimPrefix(line, "action:"))
		case strings.HasPrefix(line, "name:"):
			cur.Name = textutil.UnquoteYAML(strings.TrimPrefix(line, "name:"))
		}
	}
	flush()
	return entries
}
