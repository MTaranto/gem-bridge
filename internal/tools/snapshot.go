package tools

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/mtaranto/gem-bridge/internal/security"
)

const (
	maxSnapshotFileBytes    = 64 * 1024
	maxSnapshotContentBytes = 512 * 1024
	maxSnapshotFiles        = 300
)

var errSnapshotLimitReached = errors.New("project snapshot limit reached")

// ProjectSnapshot represents a bounded and security-filtered view of a workspace.
type ProjectSnapshot struct {
	Workspace string             `json:"workspace"`
	Branch    string             `json:"branch"`
	GitStatus []string           `json:"gitStatus"`
	GitDiff   string             `json:"gitDiff"`
	Files     []SnapshotFile     `json:"files"`
	Omitted   []SnapshotOmission `json:"omitted"`
	Truncated bool               `json:"truncated"`
}

// SnapshotFile represents one text file included in a project snapshot.
type SnapshotFile struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Content string `json:"content"`
}

// SnapshotOmission records a file or directory excluded from a snapshot.
type SnapshotOmission struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// SnapshotTools builds safe and bounded project snapshots.
type SnapshotTools struct {
	Workspace *security.Workspace
}

// NewSnapshotTools creates snapshot tools bound to an authorized workspace.
func NewSnapshotTools(workspace *security.Workspace) *SnapshotTools {
	return &SnapshotTools{
		Workspace: workspace,
	}
}

// Build creates a bounded project snapshot for the authorized workspace.
func (s *SnapshotTools) Build() (ProjectSnapshot, error) {
	gitTools := NewGitTools(s.Workspace)

	branchOutput, err := gitTools.runGit("branch", "--show-current")
	if err != nil {
		return ProjectSnapshot{}, errors.New(
			"git branch failed: " + cleanGitOutput(branchOutput, err),
		)
	}

	gitStatus, err := gitTools.StatusShort()
	if err != nil {
		return ProjectSnapshot{}, err
	}

	gitDiff, err := gitTools.Diff()
	if err != nil {
		return ProjectSnapshot{}, err
	}

	snapshot := ProjectSnapshot{
		Workspace: filepath.Base(s.Workspace.Root),
		Branch:    strings.TrimSpace(normalizeLineEndings(branchOutput)),
		GitStatus: gitStatus,
		GitDiff:   gitDiff,
		Files:     []SnapshotFile{},
		Omitted:   []SnapshotOmission{},
	}

	totalContentBytes := 0
	examinedFiles := 0

	walkErr := filepath.WalkDir(
		s.Workspace.Root,
		func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			relativePath, err := filepath.Rel(s.Workspace.Root, path)
			if err != nil {
				return err
			}

			if relativePath == "." {
				return nil
			}

			relativePath = filepath.ToSlash(relativePath)

			if entry.Type()&os.ModeSymlink != 0 {
				snapshot.Omitted = append(
					snapshot.Omitted,
					SnapshotOmission{
						Path:   relativePath,
						Reason: "symbolic links are excluded",
					},
				)

				return nil
			}

			if entry.IsDir() {
				if isExcludedSnapshotDirectory(entry.Name()) {
					snapshot.Omitted = append(
						snapshot.Omitted,
						SnapshotOmission{
							Path:   relativePath + "/",
							Reason: "excluded directory",
						},
					)

					return fs.SkipDir
				}

				return nil
			}

			examinedFiles++
			if examinedFiles > maxSnapshotFiles {
				snapshot.Truncated = true
				return errSnapshotLimitReached
			}

			if !entry.Type().IsRegular() {
				snapshot.Omitted = append(
					snapshot.Omitted,
					SnapshotOmission{
						Path:   relativePath,
						Reason: "non-regular file",
					},
				)

				return nil
			}

			if isSensitiveSnapshotFile(relativePath) {
				snapshot.Omitted = append(
					snapshot.Omitted,
					SnapshotOmission{
						Path:   relativePath,
						Reason: "sensitive file pattern",
					},
				)

				return nil
			}

			if !isSupportedSnapshotTextFile(relativePath) {
				snapshot.Omitted = append(
					snapshot.Omitted,
					SnapshotOmission{
						Path:   relativePath,
						Reason: "unsupported file type",
					},
				)

				return nil
			}

			info, err := entry.Info()
			if err != nil {
				return err
			}

			if info.Size() > maxSnapshotFileBytes {
				snapshot.Omitted = append(
					snapshot.Omitted,
					SnapshotOmission{
						Path:   relativePath,
						Reason: "file exceeds snapshot size limit",
					},
				)

				return nil
			}

			if totalContentBytes+int(info.Size()) > maxSnapshotContentBytes {
				snapshot.Omitted = append(
					snapshot.Omitted,
					SnapshotOmission{
						Path:   relativePath,
						Reason: "total snapshot content limit reached",
					},
				)

				snapshot.Truncated = true
				return errSnapshotLimitReached
			}

			resolvedPath, err := s.Workspace.ResolvePath(relativePath)
			if err != nil {
				return err
			}

			content, err := os.ReadFile(resolvedPath)
			if err != nil {
				return err
			}

			if bytes.IndexByte(content, 0) >= 0 || !utf8.Valid(content) {
				snapshot.Omitted = append(
					snapshot.Omitted,
					SnapshotOmission{
						Path:   relativePath,
						Reason: "binary or non-UTF-8 content",
					},
				)

				return nil
			}

			snapshot.Files = append(
				snapshot.Files,
				SnapshotFile{
					Path:    relativePath,
					Size:    info.Size(),
					Content: string(content),
				},
			)

			totalContentBytes += len(content)

			return nil
		},
	)

	if errors.Is(walkErr, errSnapshotLimitReached) {
		return snapshot, nil
	}

	if walkErr != nil {
		return ProjectSnapshot{}, walkErr
	}

	return snapshot, nil
}

func isExcludedSnapshotDirectory(name string) bool {
	excludedDirectories := map[string]struct{}{
		".git":         {},
		"bin":          {},
		"build":        {},
		"coverage":     {},
		"dist":         {},
		"node_modules": {},
		"tmp":          {},
		"vendor":       {},
	}

	_, excluded := excludedDirectories[strings.ToLower(name)]

	return excluded
}

func isSensitiveSnapshotFile(path string) bool {
	baseName := strings.ToLower(filepath.Base(path))

	if baseName == ".env" || strings.HasPrefix(baseName, ".env.") {
		return true
	}

	sensitiveNames := map[string]struct{}{
		"id_ed25519": {},
		"id_rsa":     {},
	}

	if _, sensitive := sensitiveNames[baseName]; sensitive {
		return true
	}

	sensitiveExtensions := map[string]struct{}{
		".key": {},
		".p12": {},
		".pem": {},
		".pfx": {},
	}

	_, sensitive := sensitiveExtensions[strings.ToLower(filepath.Ext(baseName))]

	return sensitive
}

func isSupportedSnapshotTextFile(path string) bool {
	baseName := strings.ToLower(filepath.Base(path))

	knownNames := map[string]struct{}{
		".editorconfig":  {},
		".gitattributes": {},
		".gitignore":     {},
		"dockerfile":     {},
		"license":        {},
		"makefile":       {},
	}

	if _, supported := knownNames[baseName]; supported {
		return true
	}

	supportedExtensions := map[string]struct{}{
		".c":          {},
		".conf":       {},
		".cpp":        {},
		".cs":         {},
		".css":        {},
		".go":         {},
		".graphql":    {},
		".h":          {},
		".htm":        {},
		".html":       {},
		".ini":        {},
		".java":       {},
		".js":         {},
		".json":       {},
		".jsx":        {},
		".kt":         {},
		".md":         {},
		".mod":        {},
		".php":        {},
		".properties": {},
		".proto":      {},
		".py":         {},
		".rb":         {},
		".rs":         {},
		".sh":         {},
		".sql":        {},
		".sum":        {},
		".svelte":     {},
		".swift":      {},
		".toml":       {},
		".ts":         {},
		".tsx":        {},
		".txt":        {},
		".vue":        {},
		".xml":        {},
		".yaml":       {},
		".yml":        {},
	}

	_, supported := supportedExtensions[strings.ToLower(filepath.Ext(baseName))]

	return supported
}
