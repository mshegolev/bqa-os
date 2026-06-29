package site_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsLandingPageCoversQAMemoryPilot(t *testing.T) {
	root := repoRoot(t)

	index, err := os.ReadFile(filepath.Join(root, "docs", "index.html"))
	if err != nil {
		t.Fatalf("ReadFile docs/index.html returned error: %v", err)
	}
	css, err := os.ReadFile(filepath.Join(root, "docs", "styles.css"))
	if err != nil {
		t.Fatalf("ReadFile docs/styles.css returned error: %v", err)
	}

	page := string(index)
	styles := string(css)
	requiredCopy := []string{
		"Turn scattered QA history into reusable QA memory",
		"For QA teams testing APIs, GraphQL and data pipelines",
		"10-30 sanitized QA artifacts",
		"3-5 reusable workflows",
		"project_profile.yaml",
		"common_bugs.yaml",
		"graphql_patterns.yaml",
		"etl_patterns.yaml",
		"successful_prompts.yaml",
		"bqa discover",
		"bqa ingest2",
		"bqa build",
		"local-first",
		"synthetic example",
		"Book a 20-minute QA Memory Audit",
		"Apply for 2-week paid pilot",
	}

	for _, want := range requiredCopy {
		if !strings.Contains(page, want) {
			t.Fatalf("docs/index.html should contain %q", want)
		}
	}

	for _, banned := range []string{
		"BQA-OS Agent Citadel",
		"fully autonomous QA",
		"replace QA engineers",
		"unlimited free pilot",
	} {
		if strings.Contains(page, banned) {
			t.Fatalf("docs/index.html should not contain %q", banned)
		}
	}

	if !strings.Contains(page, `href="styles.css"`) {
		t.Fatalf("docs/index.html should link docs/styles.css")
	}
	if !strings.Contains(styles, "@media") {
		t.Fatalf("docs/styles.css should include responsive rules")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root from %s", dir)
		}
		dir = parent
	}
}
