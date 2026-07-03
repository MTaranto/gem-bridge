package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePathAllowsRelativePathInsideWorkspace(t *testing.T) {
	root := t.TempDir()

	workspace, err := NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	got, err := workspace.ResolvePath("src/main.go")
	if err != nil {
		t.Fatalf("expected path to be resolved: %v", err)
	}

	want := filepath.Join(workspace.Root, "src", "main.go")

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolvePathAllowsCurrentDirectory(t *testing.T) {
	root := t.TempDir()

	workspace, err := NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	got, err := workspace.ResolvePath(".")
	if err != nil {
		t.Fatalf("expected current directory to be resolved: %v", err)
	}

	if got != workspace.Root {
		t.Fatalf("expected %q, got %q", workspace.Root, got)
	}
}

func TestResolvePathRejectsEmptyPath(t *testing.T) {
	root := t.TempDir()

	workspace, err := NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	_, err = workspace.ResolvePath("   ")
	if err == nil {
		t.Fatal("expected empty path to be rejected")
	}
}

func TestResolvePathRejectsAbsolutePath(t *testing.T) {
	root := t.TempDir()

	workspace, err := NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	_, err = workspace.ResolvePath("/etc/passwd")
	if err == nil {
		t.Fatal("expected absolute path to be rejected")
	}
}

func TestResolvePathRejectsPathTraversal(t *testing.T) {
	root := t.TempDir()

	workspace, err := NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	_, err = workspace.ResolvePath("../outside.txt")
	if err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}

func TestResolvePathRejectsSymlinkEscapingWorkspace(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	outsideFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0600); err != nil {
		t.Fatalf("expected outside file to be created: %v", err)
	}

	linkPath := filepath.Join(root, "secret-link.txt")
	if err := os.Symlink(outsideFile, linkPath); err != nil {
		t.Skipf("skipping symlink test: %v", err)
	}

	workspace, err := NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	_, err = workspace.ResolvePath("secret-link.txt")
	if err == nil {
		t.Fatal("expected symlink escape to be rejected")
	}
}
