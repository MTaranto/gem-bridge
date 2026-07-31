package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mtaranto/gem-bridge/internal/security"
)

func TestProjectSnapshotIncludesProjectStateAndTextFiles(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "checkout", "-b", "feature/snapshot")

	writeFile(t, root, "README.md", "# Test project\n")
	writeFile(t, root, "go.mod", "module example.com/test\n")
	commitAll(t, root, "initial commit")

	writeFile(t, root, "README.md", "# Test project\n\nUpdated documentation.\n")

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	snapshotTools := NewSnapshotTools(workspace)

	snapshot, err := snapshotTools.Build()
	if err != nil {
		t.Fatalf("expected snapshot creation to succeed: %v", err)
	}

	if snapshot.Workspace != filepath.Base(root) {
		t.Fatalf(
			"expected workspace name %q, got %q",
			filepath.Base(root),
			snapshot.Workspace,
		)
	}

	if snapshot.Branch != "feature/snapshot" {
		t.Fatalf(
			"expected branch feature/snapshot, got %q",
			snapshot.Branch,
		)
	}

	if !snapshotStatusContains(snapshot.GitStatus, "README.md") {
		t.Fatalf(
			"expected Git status to report README.md, got %v",
			snapshot.GitStatus,
		)
	}

	if !strings.Contains(snapshot.GitDiff, "+Updated documentation.") {
		t.Fatalf("expected Git diff to include README update:\n%s", snapshot.GitDiff)
	}

	readme := requireSnapshotFile(t, snapshot.Files, "README.md")
	if readme.Content != "# Test project\n\nUpdated documentation.\n" {
		t.Fatalf("unexpected README content: %q", readme.Content)
	}

	goMod := requireSnapshotFile(t, snapshot.Files, "go.mod")
	if goMod.Content != "module example.com/test\n" {
		t.Fatalf("unexpected go.mod content: %q", goMod.Content)
	}
}

func TestProjectSnapshotOmitsSensitiveAndUnsupportedFiles(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")

	writeFile(t, root, "README.md", "# Test project\n")
	commitAll(t, root, "initial commit")

	writeFile(t, root, ".env", "SECRET_TOKEN=hidden\n")
	writeFile(t, root, "private.pem", "private key content\n")
	writeFile(t, root, "image.png", "not a real image")

	nodeModulesPath := filepath.Join(root, "node_modules", "package")
	if err := os.MkdirAll(nodeModulesPath, 0755); err != nil {
		t.Fatalf("expected excluded directory creation to succeed: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(nodeModulesPath, "index.js"),
		[]byte("console.log('excluded');\n"),
		0644,
	); err != nil {
		t.Fatalf("expected excluded file creation to succeed: %v", err)
	}

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	snapshot, err := NewSnapshotTools(workspace).Build()
	if err != nil {
		t.Fatalf("expected snapshot creation to succeed: %v", err)
	}

	requireSnapshotOmission(t, snapshot.Omitted, ".env", "sensitive file pattern")
	requireSnapshotOmission(t, snapshot.Omitted, "private.pem", "sensitive file pattern")
	requireSnapshotOmission(t, snapshot.Omitted, "image.png", "unsupported file type")
	requireSnapshotOmission(
		t,
		snapshot.Omitted,
		"node_modules/",
		"excluded directory",
	)
}

func TestProjectSnapshotOmitsOversizedFile(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")

	writeFile(t, root, "README.md", "# Test project\n")
	commitAll(t, root, "initial commit")

	largeContent := strings.Repeat("a", maxSnapshotFileBytes+1)
	writeFile(t, root, "large.txt", largeContent)

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	snapshot, err := NewSnapshotTools(workspace).Build()
	if err != nil {
		t.Fatalf("expected snapshot creation to succeed: %v", err)
	}

	requireSnapshotOmission(
		t,
		snapshot.Omitted,
		"large.txt",
		"file exceeds snapshot size limit",
	)
}

func TestProjectSnapshotOmitsBinaryContentWithTextExtension(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")

	writeFile(t, root, "README.md", "# Test project\n")
	commitAll(t, root, "initial commit")

	binaryPath := filepath.Join(root, "binary.txt")
	if err := os.WriteFile(binaryPath, []byte{'a', 0, 'b'}, 0644); err != nil {
		t.Fatalf("expected binary test file creation to succeed: %v", err)
	}

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	snapshot, err := NewSnapshotTools(workspace).Build()
	if err != nil {
		t.Fatalf("expected snapshot creation to succeed: %v", err)
	}

	requireSnapshotOmission(
		t,
		snapshot.Omitted,
		"binary.txt",
		"binary or non-UTF-8 content",
	)
}

func TestProjectSnapshotOmitsSymbolicLinks(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")

	writeFile(t, root, "README.md", "# Test project\n")
	commitAll(t, root, "initial commit")

	linkPath := filepath.Join(root, "linked-readme.md")
	if err := os.Symlink("README.md", linkPath); err != nil {
		t.Skipf("symbolic links are not available in this environment: %v", err)
	}

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	snapshot, err := NewSnapshotTools(workspace).Build()
	if err != nil {
		t.Fatalf("expected snapshot creation to succeed: %v", err)
	}

	requireSnapshotOmission(
		t,
		snapshot.Omitted,
		"linked-readme.md",
		"symbolic links are excluded",
	)
}

func TestProjectSnapshotStopsAtTotalContentLimit(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	runGit(t, root, "init")

	writeFile(t, root, "README.md", "# Test project\n")
	commitAll(t, root, "initial commit")

	content := strings.Repeat("a", 60*1024)

	for index := 1; index <= 9; index++ {
		fileName := filepath.Join(
			root,
			"file-"+string(rune('a'+index-1))+".txt",
		)

		if err := os.WriteFile(fileName, []byte(content), 0644); err != nil {
			t.Fatalf("expected test file creation to succeed: %v", err)
		}
	}

	workspace, err := security.NewWorkspace(root)
	if err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	snapshot, err := NewSnapshotTools(workspace).Build()
	if err != nil {
		t.Fatalf("expected snapshot creation to succeed: %v", err)
	}

	if !snapshot.Truncated {
		t.Fatal("expected snapshot to be marked as truncated")
	}

	if !snapshotOmissionReasonExists(
		snapshot.Omitted,
		"total snapshot content limit reached",
	) {
		t.Fatalf(
			"expected total content limit omission, got %#v",
			snapshot.Omitted,
		)
	}
}

func requireSnapshotFile(
	t *testing.T,
	files []SnapshotFile,
	expectedPath string,
) SnapshotFile {
	t.Helper()

	for _, file := range files {
		if file.Path == expectedPath {
			return file
		}
	}

	t.Fatalf("expected snapshot file %q, got %#v", expectedPath, files)

	return SnapshotFile{}
}

func requireSnapshotOmission(
	t *testing.T,
	omissions []SnapshotOmission,
	expectedPath string,
	expectedReason string,
) {
	t.Helper()

	for _, omission := range omissions {
		if omission.Path == expectedPath && omission.Reason == expectedReason {
			return
		}
	}

	t.Fatalf(
		"expected omission %q with reason %q, got %#v",
		expectedPath,
		expectedReason,
		omissions,
	)
}

func snapshotStatusContains(status []string, expectedPath string) bool {
	for _, line := range status {
		if strings.Contains(line, expectedPath) {
			return true
		}
	}

	return false
}

func snapshotOmissionReasonExists(
	omissions []SnapshotOmission,
	expectedReason string,
) bool {
	for _, omission := range omissions {
		if omission.Reason == expectedReason {
			return true
		}
	}

	return false
}
