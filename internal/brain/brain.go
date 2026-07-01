package brain

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/mshegolev/bqa-os/internal/sanitize"
)

const configFileName = "config.yaml"

type config struct {
	repoURL  string
	cacheDir string
	branch   string
}

func Connect(repoURL string, branchOption string) error {
	if strings.TrimSpace(repoURL) == "" {
		return errors.New("brain repository URL is required")
	}
	branch, err := resolveBranchOption(branchOption)
	if err != nil {
		return err
	}

	root, err := bqaHome()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}

	configPath := filepath.Join(root, configFileName)
	content := fmt.Sprintf("brain_repository: %q\nbrain_cache: %q\n", repoURL, filepath.Join(root, "brain"))
	if branch != "" {
		content += fmt.Sprintf("brain_branch: %q\n", branch)
	}
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		return err
	}

	fmt.Printf("BQA Brain connected: %s\n", repoURL)
	if branch != "" {
		fmt.Printf("Brain branch: %s\n", branch)
	}
	fmt.Printf("Config: %s\n", configPath)
	return nil
}

func Pull() error {
	cfg, err := readBrainConfig()
	if err != nil {
		return err
	}
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git is required for brain pull")
	}

	if _, err := os.Stat(filepath.Join(cfg.cacheDir, ".git")); err == nil {
		fmt.Printf("Updating BQA Brain cache: %s\n", cfg.cacheDir)
		if cfg.branch != "" {
			return checkoutBrainBranch(cfg.cacheDir, cfg.branch)
		}
		return run("git", "-C", cfg.cacheDir, "pull", "--ff-only")
	}

	if err := os.MkdirAll(filepath.Dir(cfg.cacheDir), 0o755); err != nil {
		return err
	}
	fmt.Printf("Cloning BQA Brain: %s -> %s\n", cfg.repoURL, cfg.cacheDir)
	if err := run("git", "clone", cfg.repoURL, cfg.cacheDir); err != nil {
		return err
	}
	if cfg.branch != "" {
		return checkoutBrainBranch(cfg.cacheDir, cfg.branch)
	}
	return nil
}

func Sync(runSanitize bool) error {
	cfg, err := readBrainConfig()
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(cfg.cacheDir, ".git")); err != nil {
		return errors.New("brain cache is missing; run: bqa brain pull")
	}
	if cfg.branch != "" {
		if err := checkoutBrainBranch(cfg.cacheDir, cfg.branch); err != nil {
			return err
		}
	}

	if runSanitize {
		result, err := sanitize.Path(cfg.cacheDir, true)
		if err != nil {
			return err
		}
		fmt.Printf("Sanitize scanned=%d changed=%d redactions=%d write=true\n", result.FilesScanned, result.FilesChanged, result.Redactions)
	}

	if err := run("git", "-C", cfg.cacheDir, "status", "--short"); err != nil {
		return err
	}

	if err := run("git", "-C", cfg.cacheDir, "add", "."); err != nil {
		return err
	}

	message := fmt.Sprintf("AIQA-1: update BQA Brain %s", time.Now().Format(time.RFC3339))
	if err := runAllowNoChanges("git", "-C", cfg.cacheDir, "commit", "-m", message); err != nil {
		return err
	}
	if cfg.branch != "" {
		return run("git", "-C", cfg.cacheDir, "push", "-u", "origin", cfg.branch)
	}
	return run("git", "-C", cfg.cacheDir, "push")
}

func Status() error {
	cfg, err := readBrainConfig()
	if err != nil {
		return err
	}
	fmt.Printf("Brain repository: %s\n", cfg.repoURL)
	fmt.Printf("Brain cache: %s\n", cfg.cacheDir)
	if cfg.branch != "" {
		fmt.Printf("Brain branch: %s\n", cfg.branch)
	}
	if _, err := os.Stat(filepath.Join(cfg.cacheDir, ".git")); err != nil {
		fmt.Println("Cache status: missing")
		return nil
	}
	fmt.Println("Cache status: present")
	return run("git", "-C", cfg.cacheDir, "status", "--short")
}

// CacheDir returns the connected brain cache directory, or an error if the
// brain is not connected (used by the github export backend).
func CacheDir() (string, error) {
	_, cacheDir, err := readConfig()
	return cacheDir, err
}

func readConfig() (string, string, error) {
	cfg, err := readBrainConfig()
	if err != nil {
		return "", "", err
	}
	return cfg.repoURL, cfg.cacheDir, nil
}

func readBrainConfig() (config, error) {
	root, err := bqaHome()
	if err != nil {
		return config{}, err
	}
	configPath := filepath.Join(root, configFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config{}, fmt.Errorf("brain is not connected; run: bqa brain connect <repo-url>")
	}

	cfg := config{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "brain_repository:") {
			cfg.repoURL = unquote(strings.TrimSpace(strings.TrimPrefix(line, "brain_repository:")))
		}
		if strings.HasPrefix(line, "brain_cache:") {
			cfg.cacheDir = unquote(strings.TrimSpace(strings.TrimPrefix(line, "brain_cache:")))
		}
		if strings.HasPrefix(line, "brain_branch:") {
			cfg.branch = unquote(strings.TrimSpace(strings.TrimPrefix(line, "brain_branch:")))
		}
	}
	if cfg.repoURL == "" {
		return config{}, fmt.Errorf("brain_repository is missing in %s", configPath)
	}
	if cfg.cacheDir == "" {
		cfg.cacheDir = filepath.Join(root, "brain")
	}
	return cfg, nil
}

func bqaHome() (string, error) {
	if v := strings.TrimSpace(os.Getenv("BQA_HOME")); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".bqa"), nil
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"")
	return value
}

func resolveBranchOption(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if strings.EqualFold(value, "auto") {
		return autoBrainBranch(), nil
	}
	return policyBranchName(value), nil
}

func autoBrainBranch() string {
	for _, value := range []string{
		gitConfig("user.email"),
		gitConfig("user.name"),
		os.Getenv("USER"),
		hostname(),
		"local",
	} {
		if branch := policyBranchName(value); branch != "" {
			return branch
		}
	}
	return "fb-aiqa-local"
}

func policyBranchName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if isPolicyBranch(value) {
		return value
	}
	if strings.Contains(value, "@") {
		value = strings.Split(value, "@")[0]
	}
	slug := slugify(value)
	if slug == "" {
		return ""
	}
	return "fb-aiqa-" + slug
}

func isPolicyBranch(value string) bool {
	return value == "master" ||
		value == "commercials" ||
		strings.HasPrefix(value, "fb-") ||
		strings.HasPrefix(value, "hf-") ||
		strings.HasPrefix(value, "ch-pick-hf-")
}

func slugify(value string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(value) {
		var out rune
		switch {
		case r >= 'a' && r <= 'z':
			out = r
		case r >= '0' && r <= '9':
			out = r
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			out = '-'
		default:
			out = '-'
		}
		if out == '-' {
			if prevDash || b.Len() == 0 {
				continue
			}
			prevDash = true
			b.WriteRune(out)
			continue
		}
		prevDash = false
		b.WriteRune(out)
	}
	return strings.Trim(b.String(), "-")
}

func gitConfig(key string) string {
	out, err := exec.Command("git", "config", "--global", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

func checkoutBrainBranch(cacheDir string, branch string) error {
	if err := run("git", "-C", cacheDir, "fetch", "origin"); err != nil {
		return err
	}
	remoteBranch := "origin/" + branch
	if gitRefExists(cacheDir, "refs/remotes/"+remoteBranch) {
		if err := run("git", "-C", cacheDir, "checkout", "-B", branch, remoteBranch); err != nil {
			return err
		}
		return run("git", "-C", cacheDir, "pull", "--ff-only", "origin", branch)
	}
	if gitRefExists(cacheDir, "refs/heads/"+branch) {
		return run("git", "-C", cacheDir, "checkout", branch)
	}
	fmt.Printf("Brain branch %s has no remote yet; creating local branch.\n", branch)
	return run("git", "-C", cacheDir, "checkout", "-B", branch)
}

func gitRefExists(cacheDir string, ref string) bool {
	cmd := exec.Command("git", "-C", cacheDir, "show-ref", "--verify", "--quiet", ref)
	return cmd.Run() == nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runAllowNoChanges(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err == nil {
		return nil
	}
	fmt.Println("No commit created or git commit returned non-zero. Continuing to push if possible.")
	return nil
}
