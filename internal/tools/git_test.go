package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mtaranto/gem-bridge/internal/security"
)

func TestGitStatusShortReturnsEmptyStatusForCleanRepository(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	gitTools := NewGitTools(workspace)

	status, err := gitTools.StatusShort()
	if err != nil {
		t.Fatalf("expected git status to succeed: %v", err)
	}

	if len(status) != 0 {
		t.Fatalf("expected clean repository status to be empty, got %v", status)
	}
}

func TestGitStatusShortReportsUntrackedFile(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")

	filePath := filepath.Join(root, "notes.txt")
	if err := os.WriteFile(filePath, []byte("hello from gem bridge"), 0644); err != nil {
		t.Fatalf("expected test file creation to succeed: %v", err)
	}

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	gitTools := NewGitTools(workspace)

	status, err := gitTools.StatusShort()
	if err != nil {
		t.Fatalf("expected git status to succeed: %v", err)
	}

	if len(status) != 1 {
		t.Fatalf("expected one status line, got %v", status)
	}

	if status[0] != "?? notes.txt" {
		t.Fatalf("expected untracked file status, got %q", status[0])
	}
}

func TestGitStatusShortFailsOutsideRepository(t *testing.T) {
	requireGit(t)

	root := t.TempDir()

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	gitTools := NewGitTools(workspace)

	_, err = gitTools.StatusShort()
	if err == nil {
		t.Fatal("expected git status to fail outside a repository")
	}

	if !strings.Contains(err.Error(), "git status failed:") {
		t.Fatalf("expected git status failure message, got %q", err.Error())
	}
}

func requireGit(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable is not available")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}
