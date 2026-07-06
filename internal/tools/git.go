package tools

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/mtaranto/gem-bridge/internal/security"
)

const maxGitDiffBytes = 256 * 1024

// GitTools exposes safe Git operations for the authorized workspace.
//
// This toolset must never accept arbitrary shell commands from clients. Each
// Git operation should be implemented as an explicit, allowlisted command.
type GitTools struct {
	Workspace *security.Workspace
}

// NewGitTools creates a new Git toolset bound to a restricted workspace.
func NewGitTools(workspace *security.Workspace) *GitTools {
	return &GitTools{
		Workspace: workspace,
	}
}

// StatusShort returns the short Git status for the authorized workspace.
func (g *GitTools) StatusShort() ([]string, error) {
	output, err := g.runGit("status", "--short")
	if err != nil {
		return nil, errors.New("git status failed: " + cleanGitOutput(output, err))
	}

	status := strings.TrimRight(output, "\r\n")
	if status == "" {
		return []string{}, nil
	}

	status = normalizeLineEndings(status)

	return strings.Split(status, "\n"), nil
}

// Diff returns the current tracked Git diff for the authorized workspace.
//
// The command is intentionally narrow: it compares the workspace against HEAD,
// disables external diff helpers and text conversion commands, and does not
// accept client-provided Git arguments.
func (g *GitTools) Diff() (string, error) {
	output, err := g.runGit("diff", "--no-ext-diff", "--no-textconv", "HEAD", "--")
	if err != nil {
		return "", errors.New("git diff failed: " + cleanGitOutput(output, err))
	}

	if len([]byte(output)) > maxGitDiffBytes {
		return "", errors.New("git diff output exceeds maximum allowed size")
	}

	return normalizeLineEndings(output), nil
}

func (g *GitTools) runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.Workspace.Root

	output, err := cmd.CombinedOutput()

	return string(output), err
}

func cleanGitOutput(output string, err error) string {
	message := strings.TrimSpace(output)
	if message == "" {
		message = err.Error()
	}

	return message
}

func normalizeLineEndings(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}
