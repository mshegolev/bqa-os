package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runWorkspace(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := workspaceCmd()
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetArgs(args)
	err = cmd.ExecuteContext(context.Background())
	return out.String(), errb.String(), err
}

func TestWorkspaceInitAddListFlow(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, ".bqa")
	repo := filepath.Join(dir, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo/.git: %v", err)
	}

	if _, _, err := runWorkspace(t, "init", "--name", "bigdata", "--base-dir", base); err != nil {
		t.Fatalf("init: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(base, "workspace.yaml"))
	if err != nil {
		t.Fatalf("read workspace.yaml: %v", err)
	}
	if !strings.Contains(string(content), "name: \"bigdata\"") {
		t.Fatalf("workspace.yaml missing name:\n%s", content)
	}

	addOut, _, err := runWorkspace(t, "add", "main", repo, "--repo", "bigdata_testing", "--etl", "NS2", "--base-dir", base)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(addOut, "main") {
		t.Fatalf("add output missing project id:\n%s", addOut)
	}

	listOut, _, err := runWorkspace(t, "list", "--base-dir", base)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(listOut, "Workspace: bigdata") || !strings.Contains(listOut, "main") || !strings.Contains(listOut, "NS2") {
		t.Fatalf("list output missing expected content:\n%s", listOut)
	}
	if !strings.Contains(listOut, "branch_role=base") {
		t.Fatalf("list missing default branch_role=base:\n%s", listOut)
	}
}

func TestWorkspaceAddNonGitWarns(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, ".bqa")
	plain := filepath.Join(dir, "plain")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatalf("mkdir plain: %v", err)
	}
	if _, _, err := runWorkspace(t, "init", "--name", "w", "--base-dir", base); err != nil {
		t.Fatalf("init: %v", err)
	}
	stdout, stderr, err := runWorkspace(t, "add", "main", plain, "--repo", "bt", "--base-dir", base)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(stderr, "warning") {
		t.Fatalf("expected non-git warning on stderr:\n%s", stderr)
	}
	if strings.Contains(stdout, "warning") {
		t.Fatalf("warning must not be on stdout:\n%s", stdout)
	}
}

func TestWorkspaceListBeforeInitErrors(t *testing.T) {
	base := filepath.Join(t.TempDir(), ".bqa")
	if _, _, err := runWorkspace(t, "list", "--base-dir", base); err == nil {
		t.Fatalf("expected error listing before init")
	}
}

func TestWorkspaceAddBeforeInitErrors(t *testing.T) {
	base := filepath.Join(t.TempDir(), ".bqa")
	dir := t.TempDir()
	if _, _, err := runWorkspace(t, "add", "main", dir, "--repo", "bt", "--base-dir", base); err == nil {
		t.Fatalf("expected error adding before init")
	}
}
