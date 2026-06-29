package github

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestIssueJSONSourceParsesGitHubIssueSnapshot(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "issue.json")
	body := "{\n" +
		"  \"body\": \"## Manual verification\\n\\n```bash\\ngo test ./...\\n```\",\n" +
		"  \"labels\": [\n" +
		"    {\"name\": \"bqa:arch-approved\", \"color\": \"0E8A16\"},\n" +
		"    {\"name\": \"bqa:ready-dev\", \"color\": \"FBCA04\"},\n" +
		"    {\"name\": \"bqa:codex-team\", \"color\": \"BFD4F2\"}\n" +
		"  ],\n" +
		"  \"title\": \"Codex Team Pipeline MVP\"\n" +
		"}"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	source := IssueJSONSource{Path: path}
	issue, err := source.LoadTeamIssue(context.Background(), ports.TeamIssueRef{Repo: "mshegolev/bqa-os", Number: 27})
	if err != nil {
		t.Fatalf("LoadTeamIssue returned error: %v", err)
	}

	if issue.Number != 27 {
		t.Fatalf("expected fallback issue number 27, got %d", issue.Number)
	}
	if issue.Title != "Codex Team Pipeline MVP" {
		t.Fatalf("unexpected title %q", issue.Title)
	}
	for _, label := range []string{"bqa:arch-approved", "bqa:ready-dev", "bqa:codex-team"} {
		if !issueHasLabel(issue, label) {
			t.Fatalf("expected parsed label %q, got %#v", label, issue.Labels)
		}
	}
}

func issueHasLabel(issue ports.TeamIssue, name string) bool {
	for _, label := range issue.Labels {
		if label.Name == name {
			return true
		}
	}
	return false
}
