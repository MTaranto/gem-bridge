package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mtaranto/gem-bridge/internal/nativemessaging"
	"github.com/mtaranto/gem-bridge/internal/tools"
)

func TestRunRespondsToPing(t *testing.T) {
	input := frameMessage(t, []byte(`{"type":"ping"}`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected native host to succeed: %v", err)
	}

	response := readResponse(t, &output)

	if !response.Success {
		t.Fatalf("expected successful response, got error %q", response.Error)
	}

	if response.Type != "pong" {
		t.Fatalf("expected response type pong, got %q", response.Type)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected response data object, got %#v", response.Data)
	}

	if data["host"] != "gem-bridge-native-host" {
		t.Fatalf(
			"expected host gem-bridge-native-host, got %#v",
			data["host"],
		)
	}
}

func TestRunReadsFileFromConfiguredWorkspace(t *testing.T) {
	workspaceRoot := t.TempDir()
	expectedContent := "conteúdo secreto do teste"

	filePath := filepath.Join(workspaceRoot, "teste.txt")
	if err := os.WriteFile(filePath, []byte(expectedContent), 0644); err != nil {
		t.Fatalf("expected test file creation to succeed: %v", err)
	}

	t.Setenv(workspaceEnvName, workspaceRoot)

	input := frameMessage(
		t,
		[]byte(`{"type":"readFile","path":"teste.txt"}`),
	)

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected native host to succeed: %v", err)
	}

	response := readResponse(t, &output)

	if !response.Success {
		t.Fatalf("expected successful response, got error %q", response.Error)
	}

	if response.Type != "fileContent" {
		t.Fatalf("expected response type fileContent, got %q", response.Type)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected response data object, got %#v", response.Data)
	}

	if data["path"] != "teste.txt" {
		t.Fatalf("expected path teste.txt, got %#v", data["path"])
	}

	if data["content"] != expectedContent {
		t.Fatalf(
			"expected content %q, got %#v",
			expectedContent,
			data["content"],
		)
	}

	expectedSize := float64(len([]byte(expectedContent)))
	if data["size"] != expectedSize {
		t.Fatalf(
			"expected size %.0f, got %#v",
			expectedSize,
			data["size"],
		)
	}
}

func TestRunRejectsReadFileWithoutPath(t *testing.T) {
	t.Setenv(workspaceEnvName, t.TempDir())

	input := frameMessage(t, []byte(`{"type":"readFile"}`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected invalid request to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	if response.Error != "path is required for readFile" {
		t.Fatalf("expected missing path error, got %q", response.Error)
	}
}

func TestRunRejectsReadFileWithoutConfiguredWorkspace(t *testing.T) {
	t.Setenv(workspaceEnvName, "")

	input := frameMessage(
		t,
		[]byte(`{"type":"readFile","path":"teste.txt"}`),
	)

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected invalid request to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	expectedError := workspaceEnvName + " is not configured"
	if response.Error != expectedError {
		t.Fatalf("expected error %q, got %q", expectedError, response.Error)
	}
}

func TestRunBlocksReadOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	workspaceRoot := filepath.Join(root, "workspace")

	if err := os.Mkdir(workspaceRoot, 0755); err != nil {
		t.Fatalf("expected workspace creation to succeed: %v", err)
	}

	outsideFile := filepath.Join(root, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatalf("expected outside file creation to succeed: %v", err)
	}

	t.Setenv(workspaceEnvName, workspaceRoot)

	input := frameMessage(
		t,
		[]byte(`{"type":"readFile","path":"../secret.txt"}`),
	)

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected blocked request to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	if response.Error != "access outside the workspace is blocked" {
		t.Fatalf("expected workspace escape error, got %q", response.Error)
	}
}

func TestRunReturnsProjectSnapshot(t *testing.T) {
	requireGitExecutable(t)

	workspaceRoot := t.TempDir()

	runGitCommand(t, workspaceRoot, "init")
	runGitCommand(t, workspaceRoot, "checkout", "-b", "snapshot-test")

	readmePath := filepath.Join(workspaceRoot, "README.md")
	if err := os.WriteFile(
		readmePath,
		[]byte("# Snapshot test\n"),
		0644,
	); err != nil {
		t.Fatalf("expected README creation to succeed: %v", err)
	}

	runGitCommand(t, workspaceRoot, "add", ".")
	runGitCommand(
		t,
		workspaceRoot,
		"-c", "user.name=Gem Bridge Tests",
		"-c", "user.email=tests@example.com",
		"commit",
		"-m", "initial commit",
	)

	if err := os.WriteFile(
		readmePath,
		[]byte("# Snapshot test\n\nUpdated content.\n"),
		0644,
	); err != nil {
		t.Fatalf("expected README update to succeed: %v", err)
	}

	t.Setenv(workspaceEnvName, workspaceRoot)

	input := frameMessage(t, []byte(`{"type":"projectSnapshot"}`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected native host to succeed: %v", err)
	}

	response := readResponse(t, &output)

	if !response.Success {
		t.Fatalf("expected successful response, got error %q", response.Error)
	}

	if response.Type != "projectSnapshot" {
		t.Fatalf(
			"expected response type projectSnapshot, got %q",
			response.Type,
		)
	}

	encodedSnapshot, err := json.Marshal(response.Data)
	if err != nil {
		t.Fatalf("expected snapshot encoding to succeed: %v", err)
	}

	var snapshot tools.ProjectSnapshot

	if err := json.Unmarshal(encodedSnapshot, &snapshot); err != nil {
		t.Fatalf("expected snapshot decoding to succeed: %v", err)
	}

	if snapshot.Workspace != filepath.Base(workspaceRoot) {
		t.Fatalf(
			"expected workspace %q, got %q",
			filepath.Base(workspaceRoot),
			snapshot.Workspace,
		)
	}

	if snapshot.Branch != "snapshot-test" {
		t.Fatalf("expected branch snapshot-test, got %q", snapshot.Branch)
	}

	if !strings.Contains(snapshot.GitDiff, "+Updated content.") {
		t.Fatalf(
			"expected snapshot diff to include README update:\n%s",
			snapshot.GitDiff,
		)
	}

	var readmeFound bool

	for _, file := range snapshot.Files {
		if file.Path != "README.md" {
			continue
		}

		readmeFound = true

		if file.Content != "# Snapshot test\n\nUpdated content.\n" {
			t.Fatalf("unexpected README content: %q", file.Content)
		}
	}

	if !readmeFound {
		t.Fatalf("expected README.md in snapshot files: %#v", snapshot.Files)
	}
}

func TestRunRejectsProjectSnapshotWithoutConfiguredWorkspace(t *testing.T) {
	t.Setenv(workspaceEnvName, "")

	input := frameMessage(t, []byte(`{"type":"projectSnapshot"}`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected rejected request to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	expectedError := workspaceEnvName + " is not configured"
	if response.Error != expectedError {
		t.Fatalf("expected error %q, got %q", expectedError, response.Error)
	}
}

func TestRunReturnsErrorResponseForInvalidJSON(t *testing.T) {
	input := frameMessage(t, []byte(`{"type":`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected malformed JSON to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	if response.Error != "invalid JSON request" {
		t.Fatalf("expected invalid JSON error, got %q", response.Error)
	}
}

func TestRunReturnsErrorResponseWhenTypeIsMissing(t *testing.T) {
	input := frameMessage(t, []byte(`{}`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected missing type to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	if response.Error != "request type is required" {
		t.Fatalf("expected missing type error, got %q", response.Error)
	}
}

func TestRunReturnsErrorResponseForUnsupportedType(t *testing.T) {
	input := frameMessage(t, []byte(`{"type":"unknown"}`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected unsupported type to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	expectedError := "unsupported request type: unknown"
	if response.Error != expectedError {
		t.Fatalf("expected error %q, got %q", expectedError, response.Error)
	}
}

func TestRunFailsForIncompleteFrame(t *testing.T) {
	input := bytes.NewBuffer([]byte{1, 2})

	var output bytes.Buffer

	err := run(input, &output)
	if err == nil {
		t.Fatal("expected incomplete frame to fail")
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}

	if !strings.Contains(err.Error(), "read request:") {
		t.Fatalf("expected read request context, got %q", err.Error())
	}
}

func frameMessage(t *testing.T, payload []byte) bytes.Buffer {
	t.Helper()

	var buffer bytes.Buffer

	if err := nativemessaging.WriteMessage(&buffer, payload); err != nil {
		t.Fatalf("expected message framing to succeed: %v", err)
	}

	return buffer
}

func readResponse(t *testing.T, reader io.Reader) Response {
	t.Helper()

	payload, err := nativemessaging.ReadMessage(reader)
	if err != nil {
		t.Fatalf("expected framed response read to succeed: %v", err)
	}

	var response Response

	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("expected response JSON decoding to succeed: %v", err)
	}

	return response
}

func requireGitExecutable(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable is not available")
	}
}

func runGitCommand(t *testing.T, directory string, args ...string) {
	t.Helper()

	command := exec.Command("git", args...)
	command.Dir = directory

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf(
			"git %s failed: %v\n%s",
			strings.Join(args, " "),
			err,
			string(output),
		)
	}
}
