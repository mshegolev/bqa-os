package selfupdate

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAssetName(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"linux", "amd64", "bqa-linux-amd64"},
		{"linux", "arm64", "bqa-linux-arm64"},
		{"darwin", "amd64", "bqa-darwin-amd64"},
		{"darwin", "arm64", "bqa-darwin-arm64"},
		{"windows", "amd64", "bqa-windows-amd64.exe"},
	}
	for _, c := range cases {
		if got := AssetName(c.goos, c.goarch); got != c.want {
			t.Errorf("AssetName(%q,%q) = %q, want %q", c.goos, c.goarch, got, c.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current, tag string
		want         bool
	}{
		{"dev", "v1.0.0", true},
		{"", "v1.0.0", true},
		{"v1.0.0", "v1.0.0", false},
		{"1.0.0", "v1.0.0", false}, // normalized equal
		{"v1.0.0", "v1.2.3", true},
		{"v1.2.3", "v1.0.0", true}, // differing tag => updatable
	}
	for _, c := range cases {
		if got := IsNewer(c.current, c.tag); got != c.want {
			t.Errorf("IsNewer(%q,%q) = %v, want %v", c.current, c.tag, got, c.want)
		}
	}
}

// fakeRelease builds a test server that serves a release JSON with an asset for
// each requested OS/arch, plus serves the asset bytes.
func fakeReleaseServer(t *testing.T, tag string, oses, arches []string, body []byte) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Asset download endpoint.
	mux.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		base := "http://" + r.Host
		var assets []asset
		for _, o := range oses {
			for _, a := range arches {
				name := AssetName(o, a)
				assets = append(assets, asset{
					Name:               name,
					BrowserDownloadURL: base + "/download/" + name,
				})
			}
		}
		rel := release{TagName: tag, Assets: assets}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rel)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestLatestReleaseSelectsAssetPerPlatform(t *testing.T) {
	srv := fakeReleaseServer(t, "v2.0.0",
		[]string{"linux", "darwin", "windows"},
		[]string{"amd64", "arm64"},
		[]byte("fake-binary"))

	cases := []struct {
		goos, goarch, wantAsset string
	}{
		{"linux", "amd64", "bqa-linux-amd64"},
		{"darwin", "arm64", "bqa-darwin-arm64"},
		{"windows", "amd64", "bqa-windows-amd64.exe"},
	}
	for _, c := range cases {
		u := &Updater{BaseURL: srv.URL, Client: srv.Client(), GOOS: c.goos, GOARCH: c.goarch}
		rel, err := u.LatestRelease()
		if err != nil {
			t.Fatalf("LatestRelease(%s/%s) error: %v", c.goos, c.goarch, err)
		}
		if rel.TagName != "v2.0.0" {
			t.Errorf("tag = %q, want v2.0.0", rel.TagName)
		}
		if rel.AssetName != c.wantAsset {
			t.Errorf("asset = %q, want %q", rel.AssetName, c.wantAsset)
		}
		if rel.DownloadURL == "" {
			t.Errorf("expected non-empty download URL for %s/%s", c.goos, c.goarch)
		}
	}
}

func TestLatestReleaseMissingAsset(t *testing.T) {
	srv := fakeReleaseServer(t, "v2.0.0",
		[]string{"linux"}, []string{"amd64"}, []byte("x"))

	u := &Updater{BaseURL: srv.URL, Client: srv.Client(), GOOS: "darwin", GOARCH: "arm64"}
	if _, err := u.LatestRelease(); err == nil {
		t.Fatal("expected error when no matching asset exists")
	}
}

func TestLatestReleaseNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	u := &Updater{BaseURL: srv.URL, Client: srv.Client(), GOOS: "linux", GOARCH: "amd64"}
	if _, err := u.LatestRelease(); err == nil {
		t.Fatal("expected error on non-200 response")
	}
}

func TestApplyReplacesBinary(t *testing.T) {
	content := []byte("new-binary-bytes")
	srv := fakeReleaseServer(t, "v3.0.0",
		[]string{"linux"}, []string{"amd64"}, content)

	// Stand in for os.Executable by pointing Apply at a temp "binary".
	dir := t.TempDir()
	exe := filepath.Join(dir, "bqa")
	if err := os.WriteFile(exe, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("seed binary: %v", err)
	}
	origExecutable := executable
	executable = func() (string, error) { return exe, nil }
	t.Cleanup(func() { executable = origExecutable })

	u := &Updater{BaseURL: srv.URL, Client: srv.Client(), GOOS: "linux", GOARCH: "amd64"}
	rel, err := u.LatestRelease()
	if err != nil {
		t.Fatalf("LatestRelease error: %v", err)
	}

	path, err := u.Apply(rel)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	// Apply resolves symlinks (e.g. macOS /var -> /private/var), so compare
	// against the resolved executable path.
	wantPath, _ := filepath.EvalSymlinks(exe)
	if path != wantPath {
		t.Errorf("replaced path = %q, want %q", path, wantPath)
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatalf("read replaced binary: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("binary content = %q, want %q", got, content)
	}

	// No leftover temp files.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "bqa" {
			t.Errorf("unexpected leftover file %q in install dir", e.Name())
		}
	}
}

func TestApplyCheckOnlyDoesNotWrite(t *testing.T) {
	// The --check path lives in the command layer and never calls Apply, so the
	// guarantee here is structural: LatestRelease must not write any file.
	srv := fakeReleaseServer(t, "v3.0.0",
		[]string{"linux"}, []string{"amd64"}, []byte("x"))

	dir := t.TempDir()
	u := &Updater{BaseURL: srv.URL, Client: srv.Client(), GOOS: "linux", GOARCH: "amd64"}
	if _, err := u.LatestRelease(); err != nil {
		t.Fatalf("LatestRelease error: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("LatestRelease wrote files: %v", entries)
	}
}
