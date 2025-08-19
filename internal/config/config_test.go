package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	yamlContent := `
timeout: 10s
log_level: info
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
    - run: "npm run test"
      description: "Run unit tests"
  commit-msg:
    - run: "python scripts/validate_commit_msg.py $1"
      description: "Validate commit message format"
`
	tmpfile, err := os.CreateTemp("", "test.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Timeout != "10s" {
		t.Errorf("expected timeout to be '10s', got '%s'", cfg.Timeout)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected log_level to be 'info', got '%s'", cfg.LogLevel)
	}

	if len(cfg.Hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(cfg.Hooks))
	}

	preCommit, ok := cfg.Hooks["pre-commit"]
	if !ok {
		t.Fatal("expected 'pre-commit' hook")
	}
	if len(preCommit) != 2 {
		t.Fatalf("expected 2 commands for 'pre-commit', got %d", len(preCommit))
	}
	if preCommit[0].Run != "npm run lint" {
		t.Errorf("expected command to be 'npm run lint', got '%s'", preCommit[0].Run)
	}

	commitMsg, ok := cfg.Hooks["commit-msg"]
	if !ok {
		t.Fatal("expected 'commit-msg' hook")
	}
	if len(commitMsg) != 1 {
		t.Fatalf("expected 1 command for 'commit-msg', got %d", len(commitMsg))
	}
}
