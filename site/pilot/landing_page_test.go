package pilot_test

import (
	"os"
	"strings"
	"testing"
)

func TestLandingPageStaticFilesExist(t *testing.T) {
	for _, name := range []string{"index.html", "styles.css", "app.js"} {
		content, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
		if len(strings.TrimSpace(string(content))) == 0 {
			t.Fatalf("expected %s to be non-empty", name)
		}
	}
}

func TestLandingPageCommunicatesPilotOffer(t *testing.T) {
	html := readFile(t, "index.html")

	requiredCopy := []string{
		"BQA-OS 2-week QA Memory Pilot",
		"QA Leads",
		"QA Automation Leads",
		"B2B SaaS, fintech, and data-heavy platforms",
		"test notes",
		"bug reports",
		"prompts",
		"regression checklists",
		"sanitized session notes",
		"reusable QA knowledge base",
		"3-5 AI-assisted QA workflows",
		"Start pilot",
	}

	for _, expected := range requiredCopy {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected landing page to contain %q", expected)
		}
	}
}

func TestLandingPageUsesLocalStaticAssetsOnly(t *testing.T) {
	for _, name := range []string{"index.html", "styles.css", "app.js"} {
		content := readFile(t, name)
		for _, forbidden := range []string{"http://", "https://", "//cdn.", "analytics"} {
			if strings.Contains(strings.ToLower(content), forbidden) {
				t.Fatalf("expected %s to avoid external or tracking references, found %q", name, forbidden)
			}
		}
	}
}

func TestLandingPageDoesNotIncludePrivateDataMarkers(t *testing.T) {
	combined := strings.ToLower(readFile(t, "index.html") + readFile(t, "styles.css") + readFile(t, "app.js"))

	for _, forbidden := range []string{
		"ringcentral",
		"password",
		"secret",
		"token",
		"bqa-brain",
		"data-123",
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("expected landing page to avoid private-data marker %q", forbidden)
		}
	}
}

func TestLandingPageCtaTargetsStartSection(t *testing.T) {
	html := readFile(t, "index.html")

	if !strings.Contains(html, `href="#start"`) {
		t.Fatal("expected primary CTA to link to #start")
	}
	if !strings.Contains(html, `id="start"`) {
		t.Fatal("expected page to include a start section with id=start")
	}
}

func readFile(t *testing.T, name string) string {
	t.Helper()

	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(content)
}
