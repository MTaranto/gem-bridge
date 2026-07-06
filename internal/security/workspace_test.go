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

func TestResolvePathRejectsAbsolutePaths(t *testing.T) {
	root := t.TempDir()

	workspace, err := NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{
			name: "unix system path",
			path: "/etc/passwd",
		},
		{
			name: "unix home path",
			path: "/home/marcio/.ssh/id_rsa",
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
			name: "windows drive-relative path",
			path: `C:Users\Marcio\.ssh\id_rsa`,
		},
		{
			name: "windows unc path",
			path: `\\server\share\secret.txt`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := workspace.ResolvePath(tt.path)
			if err == nil {
				t.Fatalf("expected absolute path %q to be rejected", tt.path)
			}
		})
	}
}

func TestResolvePathRejectsPathTraversal(t *testing.T) {
	root := t.TempDir()

	workspace, err := NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace to be created: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{
			name: "unix separator traversal",
			path: "../outside.txt",
		},
		{
			name: "windows separator traversal",
			path: `..\outside.txt`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := workspace.ResolvePath(tt.path)
			if err == nil {
				t.Fatalf("expected path traversal %q to be rejected", tt.path)
			}
		})
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
