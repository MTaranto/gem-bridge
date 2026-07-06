package security

import (
	"errors"
	"path/filepath"
	"strings"
)

const windowsDriveSeparator = ':'

// Workspace represents a restricted filesystem boundary.
//
// All file operations must resolve paths through this type before touching
// the disk. This prevents callers from accessing files outside the configured
// workspace root by using absolute paths, path traversal sequences, or unsafe
// symbolic links.
type Workspace struct {
	Root string
}

// NewWorkspace creates a new restricted workspace from the provided root path.
//
// The root path is converted to an absolute, clean, symlink-resolved path to
// ensure all later comparisons are performed against a stable filesystem
// boundary.
func NewWorkspace(root string) (*Workspace, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("workspace root cannot be empty")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	resolvedRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return nil, err
	}

	return &Workspace{
		Root: filepath.Clean(resolvedRoot),
	}, nil
}

// ResolvePath converts a user-provided relative path into an absolute path
// inside the workspace.
//
// Absolute paths are rejected intentionally. This keeps the daemon focused on
// the authorized project directory and prevents tools from reading arbitrary
// locations on the user's machine.
func (w *Workspace) ResolvePath(userPath string) (string, error) {
	trimmedPath := strings.TrimSpace(userPath)
	if trimmedPath == "" {
		return "", errors.New("path cannot be empty")
	}

	if isUnsafeAbsolutePath(trimmedPath) {
		return "", errors.New("absolute paths are not allowed")
	}

	cleanUserPath := filepath.Clean(normalizePathSeparators(trimmedPath))

	fullPath := filepath.Join(w.Root, cleanUserPath)

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	resolvedPath := resolveExistingPathOrParent(absFullPath)

	if !w.isInside(resolvedPath) {
		return "", errors.New("access outside the workspace is blocked")
	}

	return resolvedPath, nil
}

// isInside checks whether a resolved path is still inside the workspace root.
func (w *Workspace) isInside(path string) bool {
	relPath, err := filepath.Rel(w.Root, path)
	if err != nil {
		return false
	}

	return relPath != ".." && !strings.HasPrefix(relPath, ".."+string(filepath.Separator))
}

// isUnsafeAbsolutePath rejects absolute path shapes from Unix and Windows.
//
// Gem Bridge receives paths from browser-based clients, not only from the local
// operating system shell. Because of that, validation must reject Unix absolute
// paths, Windows drive-qualified paths, and UNC/network paths regardless of the
// operating system currently running the daemon.
func isUnsafeAbsolutePath(path string) bool {
	normalizedPath := strings.ReplaceAll(path, "\\", "/")

	if strings.HasPrefix(normalizedPath, "/") {
		return true
	}

	return hasWindowsDrivePrefix(normalizedPath)
}

// hasWindowsDrivePrefix reports whether a path starts with a Windows drive
// prefix such as C:, C:\, or C:/.
func hasWindowsDrivePrefix(path string) bool {
	if len(path) < 2 {
		return false
	}

	first := path[0]
	isASCIIAlpha := (first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z')

	return isASCIIAlpha && path[1] == windowsDriveSeparator
}

// normalizePathSeparators converts client-provided separators to the current
// operating system separator before cleaning and joining paths.
func normalizePathSeparators(path string) string {
	slashPath := strings.ReplaceAll(path, "\\", "/")

	return filepath.FromSlash(slashPath)
}

// resolveExistingPathOrParent resolves symlinks for an existing path.
//
// If the final path does not exist yet, it attempts to resolve the parent
// directory. This helps prevent future write operations from escaping through
// symlinked directories.
func resolveExistingPathOrParent(path string) string {
	if resolvedPath, err := filepath.EvalSymlinks(path); err == nil {
		return filepath.Clean(resolvedPath)
	}

	parentDir := filepath.Dir(path)

	if resolvedParent, err := filepath.EvalSymlinks(parentDir); err == nil {
		return filepath.Join(resolvedParent, filepath.Base(path))
	}

	return filepath.Clean(path)
}
