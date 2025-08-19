package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRepoRoot(t *testing.T) {
	// The test is running inside the git repo, so it should find the root.
	root, err := GetRepoRoot()
	if err != nil {
		t.Fatalf("GetRepoRoot() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		t.Errorf(".git directory not found in the root: %v", err)
	}
}

func TestGetHooksDir(t *testing.T) {
	hooksDir, err := GetHooksDir()
	if err != nil {
		t.Fatalf("GetHooksDir() error = %v", err)
	}

	root, err := GetRepoRoot()
	if err != nil {
		t.Fatal(err)
	}

	expectedHooksDir := filepath.Join(root, ".git", "hooks")
	if hooksDir != expectedHooksDir {
		t.Errorf("expected hooks dir to be '%s', got '%s'", expectedHooksDir, hooksDir)
	}
}
