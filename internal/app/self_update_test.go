package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

// fakeLatestReleaseServer serves a GitHub-style release payload containing an
// asset for the running OS/arch. It never serves the asset bytes because the
// --check path must not download anything.
func fakeLatestReleaseServer(t *testing.T, tag string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		name := "bqa-" + runtime.GOOS + "-" + runtime.GOARCH
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		payload := map[string]any{
			"tag_name": tag,
			"assets": []map[string]string{
				{
					"name":                 name,
					"browser_download_url": "http://" + r.Host + "/download/" + name,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestSelfUpdateCheckReportsVersionsWithoutDownloading(t *testing.T) {
	tmp := t.TempDir()
	// Sentinel file: --check must not write any binary into the workspace.
	before, _ := os.ReadDir(tmp)

	srv := fakeLatestReleaseServer(t, "v9.9.9")

	cmd := selfUpdateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--check", "--base-url", srv.URL})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := out.String()
	for _, expected := range []string{
		"Current version:",
		"Latest version:  v9.9.9",
		"Run `bqa self-update` to install v9.9.9.",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
	if strings.Contains(output, "Downloading") {
		t.Fatalf("--check must not download, got:\n%s", output)
	}

	after, _ := os.ReadDir(tmp)
	if len(after) != len(before) {
		t.Fatalf("--check wrote files into workspace: before=%v after=%v", before, after)
	}
}
