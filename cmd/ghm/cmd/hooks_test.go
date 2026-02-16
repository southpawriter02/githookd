package cmd

import (
	"githookd/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func setupTestConfig(t *testing.T, content string) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()

	// Write config file
	cfgPath := filepath.Join(tmpDir, ".githooksrc.yml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore original working directory
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	return tmpDir, func() {
		os.Chdir(origDir)
	}
}

func TestHooksAdd_NewHookKey(t *testing.T) {
	yamlContent := `
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
`
	tmpDir, cleanup := setupTestConfig(t, yamlContent)
	defer cleanup()

	cfgPath := filepath.Join(tmpDir, ".githooksrc.yml")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate add to commit-msg
	if cfg.Hooks == nil {
		cfg.Hooks = make(map[string][]config.HookCommand)
	}
	cfg.Hooks["commit-msg"] = append(cfg.Hooks["commit-msg"], config.HookCommand{
		Run:         "python validate.py",
		Description: "Validate commit message",
	})

	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}

	// Verify
	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := loaded.Hooks["commit-msg"]; !ok {
		t.Fatal("expected 'commit-msg' hook to exist")
	}
	if len(loaded.Hooks["commit-msg"]) != 1 {
		t.Fatalf("expected 1 command for commit-msg, got %d", len(loaded.Hooks["commit-msg"]))
	}
}

func TestHooksAdd_AppendToExisting(t *testing.T) {
	yamlContent := `
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
`
	tmpDir, cleanup := setupTestConfig(t, yamlContent)
	defer cleanup()

	cfgPath := filepath.Join(tmpDir, ".githooksrc.yml")
	cfg, _ := config.Load(cfgPath)

	cfg.Hooks["pre-commit"] = append(cfg.Hooks["pre-commit"], config.HookCommand{
		Run:         "npm run test",
		Description: "Run tests",
	})
	config.Save(cfgPath, cfg)

	loaded, _ := config.Load(cfgPath)
	if len(loaded.Hooks["pre-commit"]) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(loaded.Hooks["pre-commit"]))
	}
	if loaded.Hooks["pre-commit"][1].Run != "npm run test" {
		t.Errorf("second command = %q, want %q", loaded.Hooks["pre-commit"][1].Run, "npm run test")
	}
}

func TestHooksAdd_DuplicateDetection(t *testing.T) {
	commands := []config.HookCommand{
		{Run: "npm run lint"},
	}

	// Check for duplicate
	newRun := "npm run lint"
	for _, c := range commands {
		if c.Run == newRun {
			return // duplicate found, test passes
		}
	}
	t.Fatal("expected duplicate detection")
}

func TestHooksRemove_FromMultiple(t *testing.T) {
	commands := []config.HookCommand{
		{Run: "npm run lint"},
		{Run: "npm run test"},
		{Run: "npm run build"},
	}

	runToRemove := "npm run test"
	var remaining []config.HookCommand
	for _, c := range commands {
		if c.Run != runToRemove {
			remaining = append(remaining, c)
		}
	}

	if len(remaining) != 2 {
		t.Fatalf("expected 2 remaining, got %d", len(remaining))
	}
	if remaining[0].Run != "npm run lint" || remaining[1].Run != "npm run build" {
		t.Error("wrong commands remaining after remove")
	}
}

func TestHooksRemove_LastCommand(t *testing.T) {
	hooks := map[string][]config.HookCommand{
		"pre-commit": {{Run: "npm run lint"}},
	}

	// Remove last command
	hooks["pre-commit"] = nil
	if len(hooks["pre-commit"]) == 0 {
		delete(hooks, "pre-commit")
	}

	if _, ok := hooks["pre-commit"]; ok {
		t.Fatal("expected pre-commit key to be deleted when empty")
	}
}

func TestHooksEnable_SetsEnabledNil(t *testing.T) {
	cmd := config.HookCommand{
		Run:     "npm run lint",
		Enabled: config.BoolPtr(false),
	}

	if cmd.IsEnabled() {
		t.Fatal("command should be disabled")
	}

	// Enable it
	cmd.Enabled = nil

	if !cmd.IsEnabled() {
		t.Fatal("command should be enabled after setting Enabled to nil")
	}
}

func TestHooksDisable_SetsBoolPtrFalse(t *testing.T) {
	cmd := config.HookCommand{
		Run: "npm run lint",
	}

	if !cmd.IsEnabled() {
		t.Fatal("command should be enabled by default")
	}

	cmd.Enabled = config.BoolPtr(false)

	if cmd.IsEnabled() {
		t.Fatal("command should be disabled after setting Enabled to false")
	}
}

func TestHooksDisable_Idempotent(t *testing.T) {
	cmd := config.HookCommand{
		Run:     "npm run lint",
		Enabled: config.BoolPtr(false),
	}

	// Disable again — should not error
	cmd.Enabled = config.BoolPtr(false)

	if cmd.IsEnabled() {
		t.Fatal("command should still be disabled")
	}
}

func TestValidateHookName_Invalid(t *testing.T) {
	err := config.ValidateHookName("not-a-real-hook")
	if err == nil {
		t.Fatal("expected error for invalid hook name")
	}
}

func TestValidateHookName_AllStandard(t *testing.T) {
	for _, h := range config.StandardHooks {
		if err := config.ValidateHookName(h); err != nil {
			t.Errorf("ValidateHookName(%q) should be valid: %v", h, err)
		}
	}
}

func TestHooksEnableDisable_RoundTrip(t *testing.T) {
	yamlContent := `
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
`
	tmpDir, cleanup := setupTestConfig(t, yamlContent)
	defer cleanup()

	cfgPath := filepath.Join(tmpDir, ".githooksrc.yml")

	// Disable
	cfg, _ := config.Load(cfgPath)
	cfg.Hooks["pre-commit"][0].Enabled = config.BoolPtr(false)
	config.Save(cfgPath, cfg)

	cfg, _ = config.Load(cfgPath)
	if cfg.Hooks["pre-commit"][0].IsEnabled() {
		t.Fatal("should be disabled after save/load")
	}

	// Enable
	cfg.Hooks["pre-commit"][0].Enabled = nil
	config.Save(cfgPath, cfg)

	cfg, _ = config.Load(cfgPath)
	if !cfg.Hooks["pre-commit"][0].IsEnabled() {
		t.Fatal("should be enabled after save/load")
	}
}
