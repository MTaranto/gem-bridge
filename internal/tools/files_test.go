package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mtaranto/gem-bridge/internal/security"
)

func TestWriteFileCreatesFileInsideWorkspace(t *testing.T) {
	root := t.TempDir()

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	fileTools := NewFileTools(workspace)

	if err := fileTools.WriteFile("notes.txt", "hello from gem bridge"); err != nil {
		t.Fatalf("expected file to be written: %v", err)
	}

	writtenPath := filepath.Join(workspace.Root, "notes.txt")

	content, err := os.ReadFile(writtenPath)
	if err != nil {
		t.Fatalf("expected written file to be readable: %v", err)
	}

	if string(content) != "hello from gem bridge" {
		t.Fatalf("expected written content to match, got %q", string(content))
	}
}

func TestWriteFileRejectsExistingFile(t *testing.T) {
	root := t.TempDir()

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	existingPath := filepath.Join(workspace.Root, "existing.txt")
	if err := os.WriteFile(existingPath, []byte("original"), 0644); err != nil {
		t.Fatalf("expected existing file to be created: %v", err)
	}

	fileTools := NewFileTools(workspace)

	err = fileTools.WriteFile("existing.txt", "changed")
	if err == nil {
		t.Fatal("expected existing file overwrite to be rejected")
	}

	content, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("expected existing file to be readable: %v", err)
	}

	if string(content) != "original" {
		t.Fatalf("expected existing content to remain unchanged, got %q", string(content))
	}
}

func TestWriteFileRejectsContentAboveLimit(t *testing.T) {
	root := t.TempDir()

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	fileTools := NewFileTools(workspace)
	largeContent := strings.Repeat("a", maxWriteFileBytes+1)

	err = fileTools.WriteFile("large.txt", largeContent)
	if err == nil {
		t.Fatal("expected large content to be rejected")
	}

	if _, err := os.Stat(filepath.Join(workspace.Root, "large.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected large file not to be created, got error: %v", err)
	}
}

func TestWriteFileRejectsUnsafePaths(t *testing.T) {
	root := t.TempDir()

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	fileTools := NewFileTools(workspace)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "path traversal with unix separator",
			path: "../outside.txt",
		},
		{
			name: "path traversal with windows separator",
			path: `..\outside.txt`,
		},
		{
			name: "unix absolute path",
			path: "/etc/passwd",
		},
		{
			name: "windows drive path with backslashes",
			path: `C:\Users\Marcio\.ssh\id_rsa`,
		},
		{
			name: "windows drive path with slashes",
			path: "C:/Users/Marcio/.ssh/id_rsa",
		},
		{
			name: "windows unc path",
			path: `\\server\share\secret.txt`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fileTools.WriteFile(tt.path, "unsafe")
			if err == nil {
				t.Fatalf("expected unsafe path %q to be rejected", tt.path)
			}
		})
	}
}

func TestWriteFileRejectsSymlinkParentEscapingWorkspace(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	linkPath := filepath.Join(root, "outside-link")
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Skipf("skipping symlink test: %v", err)
	}

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	fileTools := NewFileTools(workspace)

	err = fileTools.WriteFile("outside-link/created.txt", "secret")
	if err == nil {
		t.Fatal("expected symlink parent escape to be rejected")
	}

	if _, err := os.Stat(filepath.Join(outside, "created.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected file outside workspace not to be created, got error: %v", err)
	}
}
