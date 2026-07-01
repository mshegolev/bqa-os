package brain

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mshegolev/bqa-os/internal/sanitize"
)

const configFileName = "config.yaml"

func Connect(repoURL string) error {
	if strings.TrimSpace(repoURL) == "" {
		return errors.New("brain repository URL is required")
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
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		return err
	}

	fmt.Printf("BQA Brain connected: %s\n", repoURL)
	fmt.Printf("Config: %s\n", configPath)
	return nil
}

func Pull() error {
	repoURL, cacheDir, err := readConfig()
	if err != nil {
		return err
	}
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git is required for brain pull")
	}

	if _, err := os.Stat(filepath.Join(cacheDir, ".git")); err == nil {
		fmt.Printf("Updating BQA Brain cache: %s\n", cacheDir)
		return run("git", "-C", cacheDir, "pull", "--ff-only")
	}

	if err := os.MkdirAll(filepath.Dir(cacheDir), 0o755); err != nil {
		return err
	}
	fmt.Printf("Cloning BQA Brain: %s -> %s\n", repoURL, cacheDir)
	return run("git", "clone", repoURL, cacheDir)
}

func Sync(runSanitize bool) error {
	_, cacheDir, err := readConfig()
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(cacheDir, ".git")); err != nil {
		return errors.New("brain cache is missing; run: bqa brain pull")
	}

	if runSanitize {
		result, err := sanitize.Path(cacheDir, true)
		if err != nil {
			return err
		}
		fmt.Printf("Sanitize scanned=%d changed=%d redactions=%d write=true\n", result.FilesScanned, result.FilesChanged, result.Redactions)
	}

	if err := run("git", "-C", cacheDir, "status", "--short"); err != nil {
		return err
	}

	if err := run("git", "-C", cacheDir, "add", "."); err != nil {
		return err
	}

	message := fmt.Sprintf("Update BQA Brain %s", time.Now().Format(time.RFC3339))
	if err := runAllowNoChanges("git", "-C", cacheDir, "commit", "-m", message); err != nil {
		return err
	}
	return run("git", "-C", cacheDir, "push")
}

func Status() error {
	repoURL, cacheDir, err := readConfig()
	if err != nil {
		return err
	}
	fmt.Printf("Brain repository: %s\n", repoURL)
	fmt.Printf("Brain cache: %s\n", cacheDir)
	if _, err := os.Stat(filepath.Join(cacheDir, ".git")); err != nil {
		fmt.Println("Cache status: missing")
		return nil
	}
	fmt.Println("Cache status: present")
	return run("git", "-C", cacheDir, "status", "--short")
}

// CacheDir returns the connected brain cache directory, or an error if the
// brain is not connected (used by the github export backend).
func CacheDir() (string, error) {
	_, cacheDir, err := readConfig()
	return cacheDir, err
}

func readConfig() (string, string, error) {
	root, err := bqaHome()
	if err != nil {
		return "", "", err
	}
	configPath := filepath.Join(root, configFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", "", fmt.Errorf("brain is not connected; run: bqa brain connect <repo-url>")
	}

	var repoURL string
	var cacheDir string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "brain_repository:") {
			repoURL = unquote(strings.TrimSpace(strings.TrimPrefix(line, "brain_repository:")))
		}
		if strings.HasPrefix(line, "brain_cache:") {
			cacheDir = unquote(strings.TrimSpace(strings.TrimPrefix(line, "brain_cache:")))
		}
	}
	if repoURL == "" {
		return "", "", fmt.Errorf("brain_repository is missing in %s", configPath)
	}
	if cacheDir == "" {
		cacheDir = filepath.Join(root, "brain")
	}
	return repoURL, cacheDir, nil
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
