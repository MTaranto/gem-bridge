package security

import (
	"errors"
	"path/filepath"
	"strings"
)

// Workspace represents a restricted filesystem boundary.
//
// All file operations must resolve paths through this type before touching
// the disk. This prevents callers from accessing files outside the configured
// workspace root by using absolute paths or path traversal sequences.
type Workspace struct {
	Root string
}

// NewWorkspace creates a new restricted workspace from the provided root path.
//
// The root path is converted to an absolute, clean path to ensure all later
// comparisons are performed against a stable filesystem boundary.
func NewWorkspace(root string) (*Workspace, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	return &Workspace{
		Root: filepath.Clean(absRoot),
	}, nil
}

// ResolvePath converts a user-provided relative path into an absolute path
// inside the workspace.
//
// Absolute paths are rejected intentionally. This keeps the daemon focused on
// the authorized project directory and prevents tools from reading arbitrary
// locations on the user's machine.
func (w *Workspace) ResolvePath(userPath string) (string, error) {
	if strings.TrimSpace(userPath) == "" {
		return "", errors.New("path cannot be empty")
	}

	cleanUserPath := filepath.Clean(userPath)

	if filepath.IsAbs(cleanUserPath) {
		return "", errors.New("absolute paths are not allowed")
	}

	fullPath := filepath.Join(w.Root, cleanUserPath)

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(w.Root, absFullPath)
	if err != nil {
		return "", err
	}

	if relPath == ".." || strings.HasPrefix(relPath, "../") {
		return "", errors.New("access outside the workspace is blocked")
	}

	return absFullPath, nil
}
