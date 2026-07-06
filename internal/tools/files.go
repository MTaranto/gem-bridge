package tools

import (
	"errors"
	"os"

	"github.com/mtaranto/gem-bridge/internal/security"
)

const maxWriteFileBytes = 64 * 1024

// FileTools exposes safe filesystem operations for the local workspace.
//
// This layer should remain small and predictable. It should not decide what
// the AI wants to do; it only executes explicitly requested file operations
// after the path has been validated by the security layer.
type FileTools struct {
	Workspace *security.Workspace
}

// NewFileTools creates a new file toolset bound to a restricted workspace.
func NewFileTools(workspace *security.Workspace) *FileTools {
	return &FileTools{
		Workspace: workspace,
	}
}

// ReadFile reads a text file from inside the authorized workspace.
func (f *FileTools) ReadFile(path string) (string, error) {
	resolvedPath, err := f.Workspace.ResolvePath(path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// WriteFile creates a new text file inside the authorized workspace.
//
// This first implementation is intentionally conservative: it refuses to
// overwrite existing files and limits content size. All paths must pass through
// the workspace security layer before any write operation touches the disk.
func (f *FileTools) WriteFile(path string, content string) error {
	if len([]byte(content)) > maxWriteFileBytes {
		return errors.New("file content exceeds maximum allowed size")
	}

	resolvedPath, err := f.Workspace.ResolvePath(path)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(resolvedPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return errors.New("file already exists")
		}

		return err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return err
	}

	return nil
}

// ListDirectory returns the entries of a directory inside the workspace.
//
// Directory names receive a trailing slash to make them easier to distinguish
// from files in plain JSON responses.
func (f *FileTools) ListDirectory(path string) ([]string, error) {
	resolvedPath, err := f.Workspace.ResolvePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(resolvedPath)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(entries))

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() {
			name += "/"
		}

		result = append(result, name)
	}

	return result, nil
}
