package tools

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/mtaranto/gem-bridge/internal/security"
)

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
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = g.Workspace.Root

	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}

		return nil, errors.New("git status failed: " + message)
	}

	status := strings.TrimRight(string(output), "\r\n")
	if status == "" {
		return []string{}, nil
	}

	status = strings.ReplaceAll(status, "\r\n", "\n")

	return strings.Split(status, "\n"), nil
}
