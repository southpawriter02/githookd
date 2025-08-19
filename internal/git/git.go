package git

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetRepoRoot returns the root directory of the Git repository.
func GetRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository")
		}
		dir = parent
	}
}

// GetHooksDir returns the path to the .git/hooks directory.
func GetHooksDir() (string, error) {
	repoRoot, err := GetRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(repoRoot, ".git", "hooks"), nil
}
