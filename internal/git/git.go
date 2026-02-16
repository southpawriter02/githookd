package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetRepoRoot returns the root directory of the Git repository
// using git rev-parse --show-toplevel.
func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository (or any parent): %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetHooksDir returns the path to the Git hooks directory,
// respecting core.hooksPath if configured.
func GetHooksDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-path", "hooks/")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to determine hooks directory: %w", err)
	}

	hooksDir := strings.TrimSpace(string(output))

	// Ensure absolute path
	if !filepath.IsAbs(hooksDir) {
		repoRoot, err := GetRepoRoot()
		if err != nil {
			return "", err
		}
		hooksDir = filepath.Join(repoRoot, hooksDir)
	}

	return filepath.Clean(hooksDir), nil
}
