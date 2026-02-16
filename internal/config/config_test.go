package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestLoad_WithEnabledField(t *testing.T) {
	yamlContent := `
hooks:
  pre-commit:
    - run: "npm run lint"
      enabled: false
    - run: "npm run test"
`
	tmpfile := writeTempFile(t, yamlContent)
	defer os.Remove(tmpfile)

	cfg, err := Load(tmpfile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	cmds := cfg.Hooks["pre-commit"]
	if cmds[0].IsEnabled() {
		t.Error("expected first command to be disabled")
	}
	if !cmds[1].IsEnabled() {
		t.Error("expected second command to be enabled (nil defaults to true)")
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  *bool
		expected bool
	}{
		{"nil defaults to true", nil, true},
		{"explicit true", BoolPtr(true), true},
		{"explicit false", BoolPtr(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := HookCommand{Enabled: tt.enabled}
			if got := hc.IsEnabled(); got != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBoolPtr(t *testing.T) {
	truePtr := BoolPtr(true)
	if truePtr == nil || *truePtr != true {
		t.Error("BoolPtr(true) should return pointer to true")
	}

	falsePtr := BoolPtr(false)
	if falsePtr == nil || *falsePtr != false {
		t.Error("BoolPtr(false) should return pointer to false")
	}
}

func TestSave(t *testing.T) {
	cfg := &Config{
		Timeout:  "10s",
		LogLevel: "info",
		Hooks: map[string][]HookCommand{
			"pre-commit": {
				{Run: "npm run lint", Description: "Run ESLint"},
				{Run: "npm run test", Description: "Run tests", Enabled: BoolPtr(false)},
			},
		},
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yml")

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load it back
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.Timeout != "10s" {
		t.Errorf("Timeout = %q, want %q", loaded.Timeout, "10s")
	}
	if loaded.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", loaded.LogLevel, "info")
	}

	cmds := loaded.Hooks["pre-commit"]
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	if cmds[0].Run != "npm run lint" {
		t.Errorf("first command Run = %q, want %q", cmds[0].Run, "npm run lint")
	}
	if cmds[1].IsEnabled() {
		t.Error("second command should be disabled after round-trip")
	}
}

func TestResolve_ValidConfig(t *testing.T) {
	cfg := &Config{
		Timeout:  "10s",
		LogLevel: "info",
		Hooks: map[string][]HookCommand{
			"pre-commit": {
				{Run: "npm run lint", Description: "Run ESLint"},
			},
		},
	}

	resolved, errs := cfg.Resolve()
	if len(errs) > 0 {
		t.Fatalf("Resolve() errors = %v", errs)
	}

	if resolved.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", resolved.Timeout)
	}
	if resolved.LogLevel != LogInfo {
		t.Errorf("LogLevel = %v, want LogInfo", resolved.LogLevel)
	}

	cmds := resolved.Hooks["pre-commit"]
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Timeout != 10*time.Second {
		t.Errorf("command timeout = %v, want 10s (inherited)", cmds[0].Timeout)
	}
}

func TestResolve_DefaultsWhenOmitted(t *testing.T) {
	cfg := &Config{
		Hooks: map[string][]HookCommand{
			"pre-commit": {
				{Run: "echo hello"},
			},
		},
	}

	resolved, errs := cfg.Resolve()
	if len(errs) > 0 {
		t.Fatalf("Resolve() errors = %v", errs)
	}

	if resolved.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", resolved.Timeout, DefaultTimeout)
	}
	if resolved.LogLevel != LogWarn {
		t.Errorf("LogLevel = %v, want LogWarn", resolved.LogLevel)
	}

	cmd := resolved.Hooks["pre-commit"][0]
	if cmd.Timeout != DefaultTimeout {
		t.Errorf("command timeout = %v, want %v", cmd.Timeout, DefaultTimeout)
	}
	if cmd.LogLevel != LogWarn {
		t.Errorf("command log level = %v, want LogWarn", cmd.LogLevel)
	}
}

func TestResolve_PerCommandOverrides(t *testing.T) {
	cfg := &Config{
		Timeout:  "10s",
		LogLevel: "warn",
		Hooks: map[string][]HookCommand{
			"pre-commit": {
				{Run: "fast-check", Timeout: "5s", LogLevel: "debug"},
				{Run: "slow-check", Timeout: "120s"},
				{Run: "no-timeout", Timeout: "none"},
			},
		},
	}

	resolved, errs := cfg.Resolve()
	if len(errs) > 0 {
		t.Fatalf("Resolve() errors = %v", errs)
	}

	cmds := resolved.Hooks["pre-commit"]
	if cmds[0].Timeout != 5*time.Second {
		t.Errorf("cmd[0] timeout = %v, want 5s", cmds[0].Timeout)
	}
	if cmds[0].LogLevel != LogDebug {
		t.Errorf("cmd[0] log level = %v, want LogDebug", cmds[0].LogLevel)
	}
	if cmds[1].Timeout != 120*time.Second {
		t.Errorf("cmd[1] timeout = %v, want 120s", cmds[1].Timeout)
	}
	if cmds[1].LogLevel != LogWarn {
		t.Errorf("cmd[1] log level = %v, want LogWarn (inherited)", cmds[1].LogLevel)
	}
	if cmds[2].Timeout != 0 {
		t.Errorf("cmd[2] timeout = %v, want 0 (none)", cmds[2].Timeout)
	}
}

func TestResolve_DisabledCommand(t *testing.T) {
	cfg := &Config{
		Hooks: map[string][]HookCommand{
			"pre-commit": {
				{Run: "echo enabled"},
				{Run: "echo disabled", Enabled: BoolPtr(false)},
			},
		},
	}

	resolved, errs := cfg.Resolve()
	if len(errs) > 0 {
		t.Fatalf("Resolve() errors = %v", errs)
	}

	cmds := resolved.Hooks["pre-commit"]
	if !cmds[0].Enabled {
		t.Error("first command should be enabled")
	}
	if cmds[1].Enabled {
		t.Error("second command should be disabled")
	}
}

func TestResolve_MissingRunField(t *testing.T) {
	cfg := &Config{
		Hooks: map[string][]HookCommand{
			"pre-commit": {
				{Description: "missing run"},
			},
		},
	}

	_, errs := cfg.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for missing run field")
	}
	if !strings.Contains(errs[0].Error(), "'run' field is required") {
		t.Errorf("error = %q, want it to mention 'run' field", errs[0])
	}
}

func TestResolve_InvalidHookName(t *testing.T) {
	cfg := &Config{
		Hooks: map[string][]HookCommand{
			"pre-comit": {
				{Run: "echo test"},
			},
		},
	}

	_, errs := cfg.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for invalid hook name")
	}
	if !strings.Contains(errs[0].Error(), "not a recognized Git hook name") {
		t.Errorf("error = %q, want it to mention 'not a recognized Git hook name'", errs[0])
	}
	if !strings.Contains(errs[0].Error(), "pre-commit") {
		t.Errorf("error = %q, want it to suggest 'pre-commit'", errs[0])
	}
}

func TestResolve_InvalidTimeout(t *testing.T) {
	cfg := &Config{
		Timeout: "banana",
	}

	_, errs := cfg.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for invalid timeout")
	}
	if !strings.Contains(errs[0].Error(), "invalid global timeout") {
		t.Errorf("error = %q, want it to mention invalid timeout", errs[0])
	}
}

func TestResolve_ZeroTimeout(t *testing.T) {
	cfg := &Config{
		Timeout: "0s",
	}

	_, errs := cfg.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for 0s timeout")
	}
	if !strings.Contains(errs[0].Error(), "use \"none\"") {
		t.Errorf("error = %q, want hint about 'none'", errs[0])
	}
}

func TestResolve_InvalidLogLevel(t *testing.T) {
	cfg := &Config{
		LogLevel: "verbose",
	}

	_, errs := cfg.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for invalid log level")
	}
	if !strings.Contains(errs[0].Error(), "valid levels are") {
		t.Errorf("error = %q, want it to list valid levels", errs[0])
	}
}

func TestResolve_MultipleErrors(t *testing.T) {
	cfg := &Config{
		Timeout:  "banana",
		LogLevel: "verbose",
		Hooks: map[string][]HookCommand{
			"pre-commit": {
				{Description: "no run"},
			},
		},
	}

	_, errs := cfg.Resolve()
	if len(errs) < 3 {
		t.Fatalf("expected at least 3 errors, got %d: %v", len(errs), errs)
	}
}

func TestResolve_EmptyConfig(t *testing.T) {
	cfg := &Config{}

	resolved, errs := cfg.Resolve()
	if len(errs) > 0 {
		t.Fatalf("Resolve() errors = %v", errs)
	}
	if resolved.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", resolved.Timeout, DefaultTimeout)
	}
	if resolved.LogLevel != LogWarn {
		t.Errorf("LogLevel = %v, want LogWarn", resolved.LogLevel)
	}
}

func TestValidateHookName(t *testing.T) {
	if err := ValidateHookName("pre-commit"); err != nil {
		t.Errorf("ValidateHookName('pre-commit') should be valid: %v", err)
	}
	if err := ValidateHookName("not-a-hook"); err == nil {
		t.Error("ValidateHookName('not-a-hook') should be invalid")
	}
}

func TestSuggestHookName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"pre-comit", "pre-commit"},
		{"pre-commti", "pre-commit"},
		{"comit-msg", "commit-msg"},
		{"xyzxyzxyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SuggestHookName(tt.input)
			if got != tt.expected {
				t.Errorf("SuggestHookName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "test-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	return tmpfile.Name()
}
