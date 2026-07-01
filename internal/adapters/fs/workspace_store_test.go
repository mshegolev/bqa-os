package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceStoreExistsLoadSave(t *testing.T) {
	base := filepath.Join(t.TempDir(), ".bqa")
	store := WorkspaceStore{}

	exists, err := store.Exists(context.Background(), base)
	if err != nil || exists {
		t.Fatalf("expected not-exists, got exists=%v err=%v", exists, err)
	}
	if err := store.Save(context.Background(), base, "hello\n"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	exists, err = store.Exists(context.Background(), base)
	if err != nil || !exists {
		t.Fatalf("expected exists after save, got exists=%v err=%v", exists, err)
	}
	content, err := store.Load(context.Background(), base)
	if err != nil || content != "hello\n" {
		t.Fatalf("Load mismatch: content=%q err=%v", content, err)
	}
}

func TestPathInspectorIsDir(t *testing.T) {
	dir := t.TempDir()
	insp := PathInspector{}

	ok, err := insp.IsDir(dir)
	if err != nil || !ok {
		t.Fatalf("expected dir true, got %v err=%v", ok, err)
	}
	file := filepath.Join(dir, "f")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if ok, _ := insp.IsDir(file); ok {
		t.Fatalf("a regular file should not be a dir")
	}
	if ok, _ := insp.IsDir(filepath.Join(dir, "missing")); ok {
		t.Fatalf("missing path should not be a dir")
	}
}

func TestPathInspectorIsGitRepo(t *testing.T) {
	insp := PathInspector{}
	plain := t.TempDir()
	if ok, _ := insp.IsGitRepo(plain); ok {
		t.Fatalf("plain dir should not be a git repo")
	}

	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if ok, err := insp.IsGitRepo(repo); err != nil || !ok {
		t.Fatalf("expected git repo true, got %v err=%v", ok, err)
	}
	// A subdirectory of the repo is inside the working tree.
	sub := filepath.Join(repo, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	if ok, err := insp.IsGitRepo(sub); err != nil || !ok {
		t.Fatalf("expected subdir of repo to be inside git tree, got %v err=%v", ok, err)
	}
}

func TestWorkspaceStoreCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := WorkspaceStore{}
	dir := t.TempDir()
	if _, err := store.Exists(ctx, dir); err == nil {
		t.Fatalf("expected error on cancelled ctx (Exists)")
	}
	if _, err := store.Load(ctx, dir); err == nil {
		t.Fatalf("expected error on cancelled ctx (Load)")
	}
	if err := store.Save(ctx, dir, "x"); err == nil {
		t.Fatalf("expected error on cancelled ctx (Save)")
	}
}

func TestPathInspectorIsGitRepoWorktreeFile(t *testing.T) {
	// A linked git worktree has a `.git` FILE (not a directory). IsGitRepo must
	// treat that as inside a git working tree too.
	insp := PathInspector{}
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, ".git"), []byte("gitdir: /somewhere/.git/worktrees/x\n"), 0o600); err != nil {
		t.Fatalf("write .git file: %v", err)
	}
	if ok, err := insp.IsGitRepo(repo); err != nil || !ok {
		t.Fatalf("expected worktree .git file to count as git repo, got %v err=%v", ok, err)
	}
}
