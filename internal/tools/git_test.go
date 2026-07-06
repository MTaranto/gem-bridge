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

func TestGitDiffReturnsEmptyDiffForCleanRepository(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")
	writeFile(t, root, "notes.txt", "hello from gem bridge\n")
	commitAll(t, root, "initial commit")

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	gitTools := NewGitTools(workspace)

	diff, err := gitTools.Diff()
	if err != nil {
		t.Fatalf("expected git diff to succeed: %v", err)
	}

	if diff != "" {
		t.Fatalf("expected clean repository diff to be empty, got %q", diff)
	}
}

func TestGitDiffReportsTrackedFileModification(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")
	writeFile(t, root, "notes.txt", "hello from gem bridge\n")
	commitAll(t, root, "initial commit")

	writeFile(t, root, "notes.txt", "hello from gem bridge\nupdated line\n")

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	gitTools := NewGitTools(workspace)

	diff, err := gitTools.Diff()
	if err != nil {
		t.Fatalf("expected git diff to succeed: %v", err)
	}

	expectedParts := []string{
		"diff --git a/notes.txt b/notes.txt",
		"+updated line",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(diff, expected) {
			t.Fatalf("expected diff to contain %q, got:\n%s", expected, diff)
		}
	}
}

func TestGitDiffFailsOutsideRepository(t *testing.T) {
	requireGit(t)

	root := t.TempDir()

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	gitTools := NewGitTools(workspace)

	_, err = gitTools.Diff()
	if err == nil {
		t.Fatal("expected git diff to fail outside a repository")
	}

	if !strings.Contains(err.Error(), "git diff failed:") {
		t.Fatalf("expected git diff failure message, got %q", err.Error())
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

func commitAll(t *testing.T, dir string, message string) {
	t.Helper()

	runGit(t, dir, "add", ".")
	runGit(
		t,
		dir,
		"-c", "user.name=Gem Bridge Tests",
		"-c", "user.email=tests@example.com",
		"commit",
		"-m", message,
	)
}

func writeFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	filePath := filepath.Join(root, relativePath)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("expected test file write to succeed: %v", err)
	}
}
