package brain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPolicyBranchNameUsesGitLabCompliantPrefix(t *testing.T) {
	tests := map[string]string{
		"m.v.shchegolev@ringcentral.com": "fb-aiqa-m-v-shchegolev",
		"Mikhail V. Shchegolev":          "fb-aiqa-mikhail-v-shchegolev",
		"fb-aiqa-existing":               "fb-aiqa-existing",
		"master":                         "master",
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			if got := policyBranchName(input); got != expected {
				t.Fatalf("policyBranchName(%q) = %q, want %q", input, got, expected)
			}
		})
	}
}

func TestResolveBranchOptionAutoReturnsPolicyBranch(t *testing.T) {
	got, err := resolveBranchOption("auto")
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("auto branch should not be empty")
	}
	if !isPolicyBranch(got) {
		t.Fatalf("auto branch %q should satisfy policy branch shape", got)
	}
}

func TestConnectWritesResolvedBrainBranch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BQA_HOME", home)

	if err := Connect("git@git.ringcentral.com:BIAnalyticsPlatform/aiqa/bqa-brain.git", "Mikhail V. Shchegolev"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(home, configFileName))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	for _, expected := range []string{
		`brain_repository: "git@git.ringcentral.com:BIAnalyticsPlatform/aiqa/bqa-brain.git"`,
		`brain_cache: "` + filepath.Join(home, "brain") + `"`,
		`brain_branch: "fb-aiqa-mikhail-v-shchegolev"`,
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("config missing %q in:\n%s", expected, content)
		}
	}
}
