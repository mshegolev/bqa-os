// Package selfupdate implements GitHub Releases based self-update for the bqa
// binary. It uses only the standard library (net/http, encoding/json, runtime,
// os, io) so the project keeps cobra as its only direct dependency.
package selfupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DefaultBaseURL is the GitHub API endpoint used to resolve the latest release.
const DefaultBaseURL = "https://api.github.com/repos/mshegolev/bqa-os"

// Updater resolves and applies updates from GitHub Releases. The base URL and
// HTTP client are injectable so tests can point at an httptest.Server.
type Updater struct {
	// BaseURL is the GitHub API repo base (e.g. DefaultBaseURL). The path
	// "/releases/latest" is appended to it.
	BaseURL string
	// Client is the HTTP client used for all requests. If nil, a client with a
	// sane timeout is used.
	Client *http.Client
	// GOOS / GOARCH default to the runtime values; overridable for tests.
	GOOS   string
	GOARCH string
}

// executable resolves the path of the running binary. It is a package variable
// so tests can override it without touching the real process executable.
var executable = os.Executable

// New returns an Updater configured for production use.
func New() *Updater {
	return &Updater{
		BaseURL: DefaultBaseURL,
		Client:  &http.Client{Timeout: 60 * time.Second},
		GOOS:    runtime.GOOS,
		GOARCH:  runtime.GOARCH,
	}
}

func (u *Updater) baseURL() string {
	if u.BaseURL != "" {
		return strings.TrimRight(u.BaseURL, "/")
	}
	return DefaultBaseURL
}

func (u *Updater) client() *http.Client {
	if u.Client != nil {
		return u.Client
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (u *Updater) goos() string {
	if u.GOOS != "" {
		return u.GOOS
	}
	return runtime.GOOS
}

func (u *Updater) goarch() string {
	if u.GOARCH != "" {
		return u.GOARCH
	}
	return runtime.GOARCH
}

// asset is a single downloadable release artifact.
type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// release is the subset of the GitHub release payload we care about.
type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

// Release is the resolved latest release for the current platform.
type Release struct {
	// TagName is the release tag, e.g. "v1.2.3".
	TagName string
	// AssetName is the platform-specific asset chosen for this OS/arch.
	AssetName string
	// DownloadURL is the browser download URL for AssetName.
	DownloadURL string
}

// AssetName returns the expected release asset name for the given OS/arch,
// matching the naming contract in .github/workflows/release.yml:
// bqa-<goos>-<goarch>  (with a .exe suffix on windows).
func AssetName(goos, goarch string) string {
	name := fmt.Sprintf("bqa-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name
}

// LatestRelease fetches the latest release and selects the asset matching the
// updater's OS/arch.
func (u *Updater) LatestRelease() (*Release, error) {
	url := u.baseURL() + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := u.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch latest release: unexpected status %s", resp.Status)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("latest release has no tag_name")
	}

	want := AssetName(u.goos(), u.goarch())
	for _, a := range rel.Assets {
		if a.Name == want {
			return &Release{
				TagName:     rel.TagName,
				AssetName:   a.Name,
				DownloadURL: a.BrowserDownloadURL,
			}, nil
		}
	}
	return nil, fmt.Errorf("no asset %q in release %s for %s/%s", want, rel.TagName, u.goos(), u.goarch())
}

// IsNewer reports whether the release tag differs from the currently running
// version. A "dev" build is always considered updatable. Comparison is a plain
// string inequality on the normalized tag (releases are immutable per tag, so a
// differing tag means a different build).
func IsNewer(currentVersion, tagName string) bool {
	if currentVersion == "dev" || currentVersion == "" {
		return true
	}
	return normalize(currentVersion) != normalize(tagName)
}

func normalize(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// Apply downloads the release asset and atomically replaces the running binary.
// It returns the path of the binary that was replaced.
func (u *Updater) Apply(rel *Release) (string, error) {
	exe, err := executable()
	if err != nil {
		return "", fmt.Errorf("resolve running binary: %w", err)
	}
	if resolved, rerr := filepath.EvalSymlinks(exe); rerr == nil {
		exe = resolved
	}

	dir := filepath.Dir(exe)

	tmp, err := os.CreateTemp(dir, ".bqa-update-*")
	if err != nil {
		return "", fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup if anything below fails before the rename.
	defer func() {
		if _, statErr := os.Stat(tmpName); statErr == nil {
			_ = os.Remove(tmpName)
		}
	}()

	if err := u.download(rel.DownloadURL, tmp); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Chmod(tmpName, 0o755); err != nil {
		return "", fmt.Errorf("chmod temp file: %w", err)
	}

	// On Windows a running executable cannot be renamed over, so move the old
	// binary aside first. On unix this branch is harmless but unnecessary, so we
	// keep the simpler atomic rename path there.
	if u.goos() == "windows" {
		old := exe + ".old"
		_ = os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			return "", fmt.Errorf("move current binary aside: %w", err)
		}
		if err := os.Rename(tmpName, exe); err != nil {
			// Try to restore the old binary on failure.
			_ = os.Rename(old, exe)
			return "", fmt.Errorf("install new binary: %w", err)
		}
		_ = os.Remove(old)
		return exe, nil
	}

	if err := os.Rename(tmpName, exe); err != nil {
		return "", fmt.Errorf("install new binary: %w", err)
	}
	return exe, nil
}

func (u *Updater) download(url string, dst io.Writer) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}
	resp, err := u.client().Do(req)
	if err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download asset: unexpected status %s", resp.Status)
	}
	if _, err := io.Copy(dst, resp.Body); err != nil {
		return fmt.Errorf("write asset: %w", err)
	}
	return nil
}
