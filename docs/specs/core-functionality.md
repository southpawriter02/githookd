# githookd Core Functionality: Design Specification

## Context

This specification addresses all Core Functionality gaps identified in the roadmap-vs-implementation gap analysis. The project is at early MVP stage with only `ghm install` and `ghm run <hook>` implemented. The `timeout` and `log_level` config fields are parsed but never enforced, there are no hook CRUD commands, no uninstall command, no config validation, no structured error reporting, and several bugs in the execution environment.

This spec is organized into seven sections, each independently implementable:

1. [Hook Lifecycle Management](#1-hook-lifecycle-management) — `hooks add/remove/enable/disable/list`
2. [Timeout Enforcement](#2-timeout-enforcement) — Making the `timeout` config field functional
3. [Logging System](#3-logging-system) — Making the `log_level` config field functional
4. [Hook-Specific Settings](#4-hook-specific-settings) — Per-command overrides for global config
5. [Execution Environment Fixes](#5-execution-environment-fixes) — Bugs and missing behaviors in `run.go` and `git.go`
6. [Config Validation](#6-config-validation) — Schema enforcement and error reporting
7. [Installation Improvements & Uninstall](#7-installation-improvements--uninstall) — Idempotency, `--dry-run`, `--force`, and `ghm uninstall`

---

## 0. Prerequisite: Config Layer Changes

All sections depend on changes to [internal/config/config.go](../../internal/config/config.go). These changes must land first.

### 0.1 Updated Structs

```go
type Config struct {
    Timeout  string                   `yaml:"timeout"`
    LogLevel string                   `yaml:"log_level"`
    Hooks    map[string][]HookCommand `yaml:"hooks"`
}

type HookCommand struct {
    Run         string `yaml:"run"`
    Description string `yaml:"description"`
    Enabled     *bool  `yaml:"enabled,omitempty"`    // NEW
    Timeout     string `yaml:"timeout,omitempty"`    // NEW
    LogLevel    string `yaml:"log_level,omitempty"`  // NEW
}
```

**Why `*bool` for Enabled:** A pointer distinguishes "not set" (nil → defaults to enabled) from "explicitly set to false" (disabled). Using `omitempty` means the field is omitted from YAML output when nil, keeping config files clean.

### 0.2 New Methods and Functions

```go
// IsEnabled returns true if the command is enabled (nil defaults to true).
func (hc HookCommand) IsEnabled() bool {
    if hc.Enabled == nil {
        return true
    }
    return *hc.Enabled
}

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool {
    return &b
}

// Save writes a Config to the given file path as YAML.
func Save(path string, cfg *Config) error { ... }

// Resolve validates and resolves a raw Config into a ResolvedConfig.
// Returns ALL validation errors, not just the first.
func (c *Config) Resolve() (*ResolvedConfig, []error) { ... }
```

### 0.3 Resolved Runtime Types

```go
type LogLevel int

const (
    LogDebug LogLevel = iota
    LogInfo
    LogWarn
    LogError
)

const DefaultTimeout = 30 * time.Second

type ResolvedConfig struct {
    Timeout  time.Duration
    LogLevel LogLevel
    Hooks    map[string][]ResolvedHookCommand
}

type ResolvedHookCommand struct {
    Run         string
    Description string
    Timeout     time.Duration // 0 = no timeout (only via "none")
    LogLevel    LogLevel
    Enabled     bool
}
```

---

## 1. Hook Lifecycle Management

### 1.0 Parent Command and Shared Infrastructure

A new `hooks` parent command groups all subcommands:

```
ghm hooks <subcommand>
```

**New files to create:**

| File | Purpose |
|------|---------|
| `cmd/ghm/cmd/hooks.go` | Parent command, shared validation helpers |
| `cmd/ghm/cmd/hooks_add.go` | `hooks add` subcommand |
| `cmd/ghm/cmd/hooks_remove.go` | `hooks remove` subcommand |
| `cmd/ghm/cmd/hooks_enable.go` | `hooks enable` subcommand |
| `cmd/ghm/cmd/hooks_disable.go` | `hooks disable` subcommand |
| `cmd/ghm/cmd/hooks_list.go` | `hooks list` subcommand |
| `cmd/ghm/cmd/hooks_test.go` | Tests for all hooks subcommands |

**Shared helpers** (in `hooks.go`):

The `standardHooks` slice currently in [cmd/ghm/cmd/install.go](../../cmd/ghm/cmd/install.go) must be accessible to all hooks subcommands. Since they share the `cmd` package, no move is required. A validation helper:

```go
func validateHookName(name string) error {
    for _, h := range standardHooks {
        if h == name {
            return nil
        }
    }
    return fmt.Errorf("unknown hook '%s'; must be one of: %s",
        name, strings.Join(standardHooks, ", "))
}
```

### 1.1 `ghm hooks add`

#### User Stories

- **US-1.1:** As a developer, I want to add a linting command to my `pre-commit` hook so that my code is checked before every commit.
- **US-1.2:** As a team lead, I want to add a commit message validation script to the `commit-msg` hook so that all commits follow our convention.
- **US-1.3:** As a developer, I want to add a second command to an existing hook so that multiple checks run in sequence.

#### Command Signature

```
ghm hooks add <hook-name> --run <command> [--description <text>]
```

| Argument/Flag | Required | Type | Description |
|---|---|---|---|
| `<hook-name>` | Yes | arg | One of the 19 standard Git hook names |
| `--run`, `-r` | Yes | string | The shell command to execute |
| `--description`, `-d` | No | string | Human-readable label |

#### Desired Behavior

1. Validate `<hook-name>` against `standardHooks`. Reject if invalid.
2. Validate `--run` is non-empty.
3. Verify cwd is inside a git repo via `git.GetRepoRoot()`.
4. Verify `.githooksrc.yml` exists. If not, error with hint to run `ghm install`.
5. Load config via `config.Load(configFile)`.
6. Initialize `cfg.Hooks` map if nil.
7. Check for duplicate: if an identical `Run` value already exists under this hook, reject.
8. Append `HookCommand{Run: runFlag, Description: descFlag}` to `cfg.Hooks[hookName]`. `Enabled` is left nil (enabled by default).
9. Save config via `config.Save(configFile, cfg)`.
10. Print: `Added command to hook "<hook-name>": <run-value>`

#### Decision Tree

```
Is <hook-name> in standardHooks?
├── NO → ERROR: "unknown hook '<name>'"
└── YES
    Is --run non-empty?
    ├── NO → ERROR: "required flag 'run' not set"
    └── YES
        Is cwd inside a git repo?
        ├── NO → ERROR: "not a git repository"
        └── YES
            Does .githooksrc.yml exist?
            ├── NO → ERROR: "config file not found; run 'ghm install' first"
            └── YES
                Load config → OK?
                ├── NO → ERROR: "failed to parse config file: <detail>"
                └── YES
                    Duplicate Run value for this hook?
                    ├── YES → ERROR: "command already exists for hook '<name>': <run>"
                    └── NO → Append, Save → SUCCESS
```

#### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Valid hook and command adds entry to `.githooksrc.yml` under the correct key |
| AC-2 | Adding to a hook with no prior entries creates the map key |
| AC-3 | Adding to an existing hook appends at the end |
| AC-4 | Invalid hook name → exit 1 with error listing valid names |
| AC-5 | Missing `--run` → exit 1 |
| AC-6 | Duplicate command → exit 1 |
| AC-7 | Missing config file → exit 1 with `ghm install` hint |
| AC-8 | Omitting `--description` produces a HookCommand with empty Description |

#### Error Scenarios

| Scenario | Message |
|---|---|
| Invalid hook name | `Error: unknown hook 'my-hook'; must be one of: applypatch-msg, pre-applypatch, ...` |
| Empty `--run` | `Error: required flag "run" not set` |
| Not a git repo | `Error: not a git repository` |
| Config missing | `Error: config file '.githooksrc.yml' not found; run 'ghm install' first` |
| Duplicate command | `Error: command already exists for hook 'pre-commit': npm run lint` |

#### Unit Test Examples

```go
func TestHooksAdd_NewHookKey(t *testing.T)
// Given config with only pre-commit, adding to commit-msg creates a new key

func TestHooksAdd_AppendToExistingHook(t *testing.T)
// Given pre-commit with 1 command, adding a second puts it at index 1

func TestHooksAdd_DuplicateCommand(t *testing.T)
// Given pre-commit with "npm run lint", adding "npm run lint" again → error

func TestHooksAdd_InvalidHookName(t *testing.T)
// "not-a-real-hook" → error containing "unknown hook"

func TestHooksAdd_EdgeCases(t *testing.T)
// Table-driven: empty description, special chars in run, nil Hooks map
```

#### Config File Mutations

**Before:**
```yaml
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
```

**Command:** `ghm hooks add pre-commit --run "npm run test" --description "Run unit tests"`

**After:**
```yaml
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
    - run: "npm run test"
      description: "Run unit tests"
```

---

### 1.2 `ghm hooks remove`

#### User Stories

- **US-2.1:** As a developer, I want to remove a specific command from a hook so that it no longer runs.
- **US-2.2:** As a team lead, I want to remove an outdated validation script from our commit-msg hook.
- **US-2.3:** As a developer, I want removing the last command from a hook to clean up the empty key.

#### Command Signature

```
ghm hooks remove <hook-name> --run <command>
```

| Argument/Flag | Required | Type | Description |
|---|---|---|---|
| `<hook-name>` | Yes | arg | Standard Git hook name |
| `--run`, `-r` | Yes | string | Exact `run` value of the command to remove |

Matching is exact string equality on the `Run` field.

#### Desired Behavior

1. Validate hook name and `--run` flag.
2. Verify git repo and config file.
3. Load config.
4. Verify `cfg.Hooks[hookName]` exists and is non-empty.
5. Find matching `HookCommand` by `Run` field. If not found, error.
6. Remove the entry. If the slice is now empty, delete the map key entirely.
7. Save config.
8. Print: `Removed command from hook "<hook-name>": <run-value>`

#### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Removes matching command from config |
| AC-2 | Preserves other commands under the same hook |
| AC-3 | Cleans up empty hook key when last command removed |
| AC-4 | Hook key not found → error |
| AC-5 | Command not found → error |
| AC-6 | Exact string matching (similar but non-identical values don't match) |

#### Unit Test Examples

```go
func TestHooksRemove_FromMultipleCommands(t *testing.T)
func TestHooksRemove_LastCommand(t *testing.T)        // verifies key deletion
func TestHooksRemove_CommandNotFound(t *testing.T)
func TestHooksRemove_HookNotConfigured(t *testing.T)
func TestHooksRemove_ExactMatchOnly(t *testing.T)
```

---

### 1.3 `ghm hooks enable`

#### User Stories

- **US-3.1:** As a developer, I want to re-enable a previously disabled command.
- **US-3.2:** As a developer, I want to enable all commands under a hook at once with `--all`.

#### Command Signature

```
ghm hooks enable <hook-name> [--run <command>] [--all]
```

| Argument/Flag | Required | Type | Description |
|---|---|---|---|
| `<hook-name>` | Yes | arg | Standard Git hook name |
| `--run`, `-r` | No | string | Specific command to enable |
| `--all`, `-a` | No | bool | Enable all commands under this hook |

**Flag logic:** Exactly one of `--run` or `--all` must be provided. Exception: if the hook has exactly one command, no flag is needed (convenience shorthand). If the hook has multiple commands and no flag is given, error with guidance.

#### Desired Behavior

For each targeted command:
- If already enabled (Enabled == nil or true): print notice, no error, no change.
- If disabled (Enabled == false): set `Enabled = nil` (which omits the field from YAML via `omitempty`).

**Design rationale:** Setting Enabled to nil (not `BoolPtr(true)`) keeps config files clean — only disabled commands show `enabled: false`.

#### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Enables a disabled command (removes `enabled: false` from YAML) |
| AC-2 | Already-enabled command is idempotent (no error) |
| AC-3 | `--all` enables all disabled commands |
| AC-4 | Single-command shorthand works without flags |
| AC-5 | Multi-command without flag → error with guidance |
| AC-6 | `--run` and `--all` together → error (mutually exclusive) |

#### Config File Mutations

**Before:**
```yaml
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
      enabled: false
    - run: "npm run test"
      description: "Run unit tests"
```

**Command:** `ghm hooks enable pre-commit --run "npm run lint"`

**After:**
```yaml
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
    - run: "npm run test"
      description: "Run unit tests"
```

---

### 1.4 `ghm hooks disable`

#### User Stories

- **US-4.1:** As a developer, I want to temporarily disable a slow test suite from my pre-commit hook to iterate faster, without losing the configuration.
- **US-4.2:** As a developer, I want to disable all commands for a hook before a large rebase.

#### Command Signature

```
ghm hooks disable <hook-name> [--run <command>] [--all]
```

Same flag semantics as `hooks enable`.

#### Desired Behavior

For each targeted command:
- If already disabled: print notice, no error.
- If enabled: set `Enabled = BoolPtr(false)`.

#### Integration with `run.go`

The execution loop in [cmd/ghm/cmd/run.go](../../cmd/ghm/cmd/run.go) must be updated to check `IsEnabled()`:

```go
for _, command := range commands {
    if !command.IsEnabled() {
        slog.Info("Skipping disabled command", "description", command.Description)
        continue
    }
    // ... existing execution logic ...
}
```

#### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Disables an enabled command (adds `enabled: false` to YAML) |
| AC-2 | Already-disabled is idempotent |
| AC-3 | `--all` disables all commands |
| AC-4 | Disabled commands are skipped by `ghm run` |
| AC-5 | Single-command shorthand works |

#### Config File Mutations

**Before:**
```yaml
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
```

**Command:** `ghm hooks disable pre-commit --run "npm run lint"`

**After:**
```yaml
hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
      enabled: false
```

---

### 1.5 `ghm hooks list`

#### User Stories

- **US-5.1:** As a developer, I want to see all configured hooks and their commands to understand what runs during each Git event.
- **US-5.2:** As a developer, I want to see enabled/disabled status at a glance.
- **US-5.3:** As a new team member, I want to list hooks to learn the project's automated checks.

#### Command Signature

```
ghm hooks list [--hook <hook-name>] [--json] [--quiet]
```

| Argument/Flag | Required | Type | Description |
|---|---|---|---|
| `--hook`, `-H` | No | string | Filter to a specific hook |
| `--json` | No | bool | JSON output for scripting |
| `--quiet`, `-q` | No | bool | Hook names only, one per line |

#### Output Formats

**Default (human-readable):**
```
pre-commit (2 commands)
  [enabled]  npm run lint          Run ESLint
  [disabled] npm run test          Run unit tests

commit-msg (1 command)
  [enabled]  python scripts/validate_commit_msg.py $1   Validate commit message format
```

**JSON (`--json`):**
```json
{
  "pre-commit": [
    {"run": "npm run lint", "description": "Run ESLint", "enabled": true},
    {"run": "npm run test", "description": "Run unit tests", "enabled": false}
  ]
}
```

**Quiet (`--quiet`):**
```
commit-msg
pre-commit
```

Hooks are sorted alphabetically. Empty config → `No hooks configured.` (exit 0).

#### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Lists all hooks with commands and enabled/disabled annotations |
| AC-2 | Alphabetical ordering by hook name |
| AC-3 | `--hook` filter shows only the specified hook |
| AC-4 | `--json` produces valid JSON with all data |
| AC-5 | `--quiet` shows only hook names |
| AC-6 | Empty config handled gracefully (exit 0) |
| AC-7 | Invalid `--hook` name rejected |

**This command is read-only. It does not mutate the config file.**

---

## 2. Timeout Enforcement

### User Stories

- **US-A1:** As a developer, I want a global timeout so that a runaway hook does not block my Git workflow indefinitely.
- **US-A2:** As a project lead, I want to set a short global timeout but override it per-command for legitimately long-running tasks.
- **US-A3:** As a developer, when a hook times out I want a clear error message explaining what happened and how to fix it.

### Config Schema

The global `timeout` field already exists on `Config`. A new per-command `timeout` field is added to `HookCommand` (see Section 0.1).

Valid values: any string parseable by `time.ParseDuration` (e.g., `"5s"`, `"2m30s"`, `"500ms"`), or the special string `"none"` (per-command only, disables timeout).

**Default:** `30s` when no timeout is specified at any level.

### Resolution Decision Tree

```
command.timeout == "none"?
├── YES → no timeout enforced (duration = 0 sentinel)
└── NO
    command.timeout non-empty?
    ├── YES → parse and use command timeout
    └── NO
        global timeout non-empty?
        ├── YES → parse and use global timeout
        └── NO → use DefaultTimeout (30s)
```

### Implementation

Timeout enforcement uses `context.WithTimeout` and `exec.CommandContext` in [cmd/ghm/cmd/run.go](../../cmd/ghm/cmd/run.go). When the context expires, the process is killed.

### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Command running past global timeout is killed with timeout-specific error |
| AC-2 | Per-command timeout overrides global timeout |
| AC-3 | `timeout: none` disables timeout for that command |
| AC-4 | No timeout specified at any level → 30s default applied |
| AC-5 | Timeout error message includes: hook name, command, timeout duration, and hint |
| AC-6 | Timeout fires → subsequent commands NOT executed (abort semantics preserved) |
| AC-7 | Invalid timeout string → validation error at config load, not at execution |
| AC-8 | `timeout: "0s"` → validation error with hint to use `"none"` |

### Error Output on Timeout

```
Error: hook "pre-commit" timed out

  Command:  npm run test
  Timeout:  5s
  Status:   killed (context deadline exceeded)

  Hint: To increase the timeout for this command, add 'timeout: 60s'
        to the command entry in .githooksrc.yml, or increase the
        global 'timeout' setting.
```

### Config YAML Example

```yaml
timeout: 10s          # global default

hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
      # inherits 10s

    - run: "npm run test"
      description: "Run unit tests"
      timeout: 120s    # override: 2 minutes

  pre-push:
    - run: "./scripts/full-integration-test.sh"
      description: "Full integration test suite"
      timeout: none    # no timeout
```

### Unit Test Examples

```go
func TestResolveTimeout_GlobalOnly(t *testing.T)           // inherits global
func TestResolveTimeout_PerCommandOverride(t *testing.T)   // command wins
func TestResolveTimeout_NoneDisablesTimeout(t *testing.T)  // "none" → 0
func TestResolveTimeout_InvalidGlobal(t *testing.T)        // "banana" → error
func TestResolveTimeout_DefaultWhenOmitted(t *testing.T)   // empty → 30s
```

---

## 3. Logging System

### User Stories

- **US-B1:** As a developer, I want `log_level: debug` to see exactly what commands are being resolved, what timeouts apply, and what env vars are set.
- **US-B2:** As a CI engineer, I want the default log level to be `warn` so that successful hooks produce no output.
- **US-B3:** As a project lead, I want to set `log_level: debug` on a specific problematic hook without drowning in output from all others.

### Log Level Definitions

| Level | Purpose | Examples |
|---|---|---|
| `debug` | Internal state, troubleshooting | "Resolved timeout: 10s (global)", env var list |
| `info` | Normal operational messages | "Running command: npm run lint", completion summary |
| `warn` | Unexpected but non-fatal | "Config field 'timeout' empty, using default 30s" |
| `error` | Failures (always shown) | "Command failed: exit code 1", "Timed out after 10s" |

**Error-level messages are ALWAYS emitted regardless of configured log level.**

### Implementation Approach: `log/slog` (standard library)

**Rationale:** Zero new dependencies. The project has only 3 deps (cobra, pflag, yaml.v3). `log/slog` is part of Go's standard library since 1.21, supports structured key-value logging, and has built-in level filtering.

**New package:** `internal/logging/logging.go`

```go
func Setup(level slog.Level, w io.Writer)      // initialize global logger
func ParseLevel(s string) (slog.Level, error)  // "debug"→LevelDebug, etc.
```

**Critical:** Logging goes to `os.Stderr`. Hook script stdout stays on `os.Stdout`. This separation ensures Git receives expected output from hooks.

All existing `fmt.Printf` calls in `run.go` and `install.go` are replaced with `slog.Debug()`, `slog.Info()`, `slog.Warn()`, or `slog.Error()`.

### What Gets Logged at Each Level

| Level | On Success | On Failure |
|---|---|---|
| `debug` | Config resolution, env vars, command start/stop, timing, full stdout/stderr | All of the above + structured error report |
| `info` | "Running command: ...", completion summary | Structured error report |
| `warn` (default) | Nothing | Warnings + structured error report |
| `error` | Nothing | Structured error report only |

### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | `debug` shows config resolution, env vars, timing |
| AC-2 | `info` shows "Running command: ..." and completion summary |
| AC-3 | `warn` (default) produces no output on success |
| AC-4 | `error` shows only failures |
| AC-5 | Error messages always shown regardless of level |
| AC-6 | All log output goes to stderr, not stdout |
| AC-7 | Invalid `log_level` value → validation error |
| AC-8 | Per-command `log_level` overrides global |

### Unit Test Examples

```go
func TestParseLevel_ValidLevels(t *testing.T)
func TestParseLevel_EmptyDefaultsToWarn(t *testing.T)
func TestParseLevel_InvalidReturnsError(t *testing.T)
func TestSetup_DebugLevelShowsAllMessages(t *testing.T)
func TestSetup_ErrorLevelSuppressesInfoAndWarn(t *testing.T)
```

---

## 4. Hook-Specific Settings

### Resolution Pattern

Both `timeout` and `log_level` follow the same override pattern:

```
command-level value set?
├── YES → validate and use it
└── NO
    global-level value set?
    ├── YES → validate and use it
    └── NO → use default
```

**Defaults:** `timeout` → `30s`, `log_level` → `warn`

Resolution happens at config load time in `Resolve()`, not at execution time. This surfaces all validation errors before any hook runs.

### Config YAML Example

```yaml
timeout: 10s
log_level: warn

hooks:
  pre-commit:
    - run: "npm run lint"
      # inherits timeout=10s, log_level=warn

    - run: "npm run test"
      timeout: 120s        # override
      log_level: debug     # override

  pre-push:
    - run: "./scripts/deploy-check.sh"
      timeout: none
      log_level: info
```

### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Per-command timeout overrides global; absent inherits global; both absent uses default |
| AC-2 | Per-command log_level overrides global; same inheritance |
| AC-3 | `Resolve()` collects ALL validation errors across all hooks/commands |
| AC-4 | `Description` field is printed at `info` level when a command starts |

---

## 5. Execution Environment Fixes

### 5.1 The `cobra.ExactArgs(1)` Bug

**File:** [cmd/ghm/cmd/run.go:19](../../cmd/ghm/cmd/run.go)

**Problem:** `Args: cobra.ExactArgs(1)` means Cobra rejects any extra arguments. When Git invokes `commit-msg .git/COMMIT_EDITMSG`, the binary receives `["run", "commit-msg", ".git/COMMIT_EDITMSG"]`. Cobra sees 2 args for `run` and fails.

Meanwhile, `hookArgs := args[1:]` on line 22 is always empty because args always has exactly 1 element.

**Fix:** Change to `cobra.MinimumNArgs(1)`.

| # | Acceptance Criterion |
|---|---|
| AC-1 | `ghm run pre-commit` (no extra args) works |
| AC-2 | `ghm run commit-msg .git/COMMIT_EDITMSG` passes the path to the hook command |
| AC-3 | Symlink invocation passes Git-provided args through |
| AC-4 | `ghm run` (no args at all) still produces usage error |

### 5.2 `GetRepoRoot()` — Use `git rev-parse`

**File:** [internal/git/git.go](../../internal/git/git.go)

**Problem:** Pure filesystem walk doesn't respect `$GIT_DIR`, `--separate-git-dir`, or worktrees (where `.git` is a file, not a directory).

**Fix:** Use `git rev-parse --show-toplevel`. This handles all edge cases and is the canonical approach.

```go
func GetRepoRoot() (string, error) {
    cmd := exec.Command("git", "rev-parse", "--show-toplevel")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("not a git repository (or any parent): %w", err)
    }
    return strings.TrimSpace(string(output)), nil
}
```

| # | Acceptance Criterion |
|---|---|
| AC-1 | Correct path from subdirectory |
| AC-2 | Correct path with `$GIT_DIR` set |
| AC-3 | Correct path inside a `git worktree` |
| AC-4 | Descriptive error outside any repo |

### 5.3 `GetHooksDir()` — Respect `core.hooksPath`

**File:** [internal/git/git.go](../../internal/git/git.go)

**Problem:** Hardcoded to `.git/hooks`. Breaks when user has `core.hooksPath` configured.

**Fix:** Use `git rev-parse --git-path hooks/`.

```go
func GetHooksDir() (string, error) {
    cmd := exec.Command("git", "rev-parse", "--git-path", "hooks/")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to determine hooks directory: %w", err)
    }
    hooksDir := strings.TrimSpace(string(output))
    if !filepath.IsAbs(hooksDir) {
        repoRoot, err := GetRepoRoot()
        if err != nil {
            return "", err
        }
        hooksDir = filepath.Join(repoRoot, hooksDir)
    }
    return filepath.Clean(hooksDir), nil
}
```

| # | Acceptance Criterion |
|---|---|
| AC-1 | Returns `.git/hooks` for standard repos |
| AC-2 | Returns configured path when `core.hooksPath` is set |
| AC-3 | Always returns an absolute path |

### 5.4 Working Directory Enforcement

**File:** [cmd/ghm/cmd/run.go](../../cmd/ghm/cmd/run.go)

**Problem:** `exec.Command` does not set `cmd.Dir`. Hooks may not execute from the repo root.

**Fix:** Set `cmd.Dir = repoRoot` on the spawned process.

| # | Acceptance Criterion |
|---|---|
| AC-1 | A hook running `pwd` outputs the repo root |
| AC-2 | Relative paths in commands resolve relative to repo root |

### 5.5 Structured Error Reporting

**File:** [cmd/ghm/cmd/run.go](../../cmd/ghm/cmd/run.go)

**Problem:** Failures report only `"Command failed: exit status 1"`. No hook name, command string, exit code, or captured output.

**Fix:** Use `io.MultiWriter` to stream AND capture stdout/stderr. On failure, produce a structured report.

**Stdout/stderr capture:**

```go
var stdoutBuf, stderrBuf bytes.Buffer
cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
```

**New type:** `HookError` in a new package `internal/runner/errors.go`

```go
type HookError struct {
    HookName   string
    Command    string
    ExitCode   int
    Stdout     string
    Stderr     string
    TimedOut   bool
    TimeoutDur time.Duration
}

func (e *HookError) FormatReport() string { ... }
```

**Error output format:**

```
===========================================================
 HOOK FAILED
===========================================================

  Hook:      pre-commit
  Command:   npm run lint
  Exit Code: 2

-----------------------------------------------------------
 STDOUT
-----------------------------------------------------------
  src/index.ts: warning: unused variable 'x'

-----------------------------------------------------------
 STDERR
-----------------------------------------------------------
  ESLint found 1 error and 3 warnings.

===========================================================
```

**Security:** Error reports NEVER include environment variables (which may contain tokens). Env vars are only logged at `debug` level with a sensitivity warning.

| # | Acceptance Criterion |
|---|---|
| AC-1 | Failure report includes: hook name, command, exit code, stdout, stderr |
| AC-2 | Stdout/stderr stream in real-time AND are captured |
| AC-3 | Success produces no error report |
| AC-4 | Environment variables never in error report |
| AC-5 | Timeout produces "timed out after Xs" instead of numeric exit code |
| AC-6 | Stdout/stderr truncated at 10KB each in the report |

### Unit Test Examples for Section 5

```go
func TestRunCmd_MinimumNArgs(t *testing.T)
func TestRunCmd_NoExtraArgs(t *testing.T)
func TestGetRepoRoot_RespectsGitDir(t *testing.T)
func TestGetHooksDir_RespectsHooksPath(t *testing.T)
func TestGetHooksDir_ReturnsAbsolutePath(t *testing.T)
func TestRunCommand_SetsWorkingDirectory(t *testing.T)
func TestRunCommand_CapturesStdoutStderr(t *testing.T)
func TestHookError_FormatReport(t *testing.T)
func TestHookError_TimeoutReport(t *testing.T)
```

---

## 6. Config Validation

### What Gets Validated

| # | Rule | Severity |
|---|---|---|
| V1 | `HookCommand.Run` must be non-empty | Error |
| V2 | Hook names in `Config.Hooks` must be in `standardHooks` | Error |
| V3 | Timeout must be valid `time.ParseDuration` or `"none"`, positive, not `"0s"` | Error |
| V4 | Log level must be `debug`, `info`, `warn`, `error`, or empty | Error |
| V5 | Empty `hooks:` block (key exists, no hooks defined) | Warning |
| V6 | Duplicate `run` values within a single hook | Warning |

### Error Messages

**V1 — Missing run field:**
```
Error: invalid configuration in .githooksrc.yml
  Hook:     pre-commit
  Command:  #2 (index 1)
  Problem:  'run' field is required but missing or empty
```

**V2 — Invalid hook name (with typo suggestion):**
```
Error: invalid configuration in .githooksrc.yml
  Hook:     "pre-comit"
  Problem:  not a recognized Git hook name
  Hint:     Did you mean "pre-commit"?
```

Uses Levenshtein distance to suggest the closest valid name.

**V3 — Invalid timeout:**
```
Error: invalid configuration in .githooksrc.yml
  Field:    timeout
  Value:    "banana"
  Problem:  cannot parse as a duration
  Hint:     Valid formats: "10s", "2m30s", "500ms"
```

**V4 — Invalid log level:**
```
Error: invalid configuration in .githooksrc.yml
  Field:    log_level
  Value:    "verbose"
  Hint:     Valid log levels: debug, info, warn, error
```

### When Validation Runs

1. `Load()` catches YAML syntax errors (malformed file).
2. `Resolve()` validates all fields and returns ALL errors (not just the first).
3. In `ghm run`, validation runs before any command executes. If validation fails, exit 1 with all errors listed.

### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Missing `run` field → error identifying hook and command index |
| AC-2 | Unrecognized hook name → error with "did you mean?" if close match exists |
| AC-3 | Multiple errors → ALL reported in single output |
| AC-4 | Empty hooks / duplicate commands → warnings, not errors, don't block execution |
| AC-5 | Validation runs before any hook command executes |

### Unit Test Examples

```go
func TestResolve_MissingRunField(t *testing.T)
func TestResolve_InvalidHookName(t *testing.T)
func TestResolve_MultipleErrors(t *testing.T)          // all collected
func TestResolve_ValidConfig(t *testing.T)             // no errors
func TestResolve_DuplicateCommandWarning(t *testing.T) // warning, not error
```

---

## 7. Installation Improvements & Uninstall

### 7.1 Install Command Improvements

**File:** [cmd/ghm/cmd/install.go](../../cmd/ghm/cmd/install.go)

#### Updated Command Signature

```
ghm install [--dry-run] [--force]
```

| Flag | Description |
|---|---|
| `--dry-run` | Preview actions without making changes |
| `--force` | Reinstall all hooks even if already symlinked to ghm |

#### Idempotency Decision Tree (per hook)

```
Hook path P exists? (os.Lstat)
├── NO → Create symlink → "Installed: H"
└── YES
    Is P a symlink?
    ├── NO (regular file) → Backup P, create symlink → "Backed up + Installed: H"
    └── YES
        Symlink target resolves to ghmPath?
        ├── YES
        │   --force flag?
        │   ├── YES → Remove + recreate symlink → "Reinstalled: H (forced)"
        │   └── NO → Skip → "Skipped: H (already managed by ghm)"
        └── NO (foreign symlink) → Backup P, create symlink → "Backed up + Installed: H"
```

**Backup naming:** If `.bak` already exists, use timestamped name: `<hook>.bak.YYYYMMDDHHMMSS` (via `time.Now().Format("20060102150405")`).

**Symlink target comparison:** Uses `filepath.EvalSymlinks()` on both the readlink target and `os.Executable()` to compare canonical paths.

#### Summary Output

```
Installation complete: 19 hooks installed, 0 skipped, 3 backed up.
```

With `--dry-run`:
```
Dry run complete: 19 hooks would be installed, 0 would be skipped, 3 would be backed up.
```

#### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Not in a git repo → exit 1, no files created |
| AC-2 | Fresh repo → creates `.githooks/`, `.githooksrc.yml`, 19 symlinks |
| AC-3 | Second run (no changes) → all 19 skipped, 0 mutations |
| AC-4 | Pre-existing regular file → backed up to `.bak`, replaced with symlink |
| AC-5 | `.bak` already exists → timestamped backup name |
| AC-6 | `--dry-run` → shows plan, zero filesystem changes |
| AC-7 | `--force` → removes and recreates all 19 symlinks |
| AC-8 | Summary line shows correct counts |

#### Testability

Extract core logic into `doInstall(hooksDir, ghmPath string, dryRun, force bool) (installed, skipped, backedUp int, err error)` so tests can pass controlled paths.

#### Unit Test Examples

```go
func TestInstallFreshRepo(t *testing.T)
func TestInstallIdempotency(t *testing.T)
func TestInstallBacksUpExistingHook(t *testing.T)
func TestInstallTimestampedBackupWhenBakExists(t *testing.T)
func TestInstallSkipsForeignSymlink(t *testing.T)
func TestInstallDryRunMakesNoChanges(t *testing.T)
func TestInstallNotGitRepo(t *testing.T)
func TestInstallForceReplacesGhmSymlink(t *testing.T)
```

---

### 7.2 Uninstall Command (`ghm uninstall`)

**New file:** `cmd/ghm/cmd/uninstall.go`

#### User Stories

- **US-B1:** As a developer who no longer wants githookd, I want to run `ghm uninstall` to restore my repo to pre-githookd state.
- **US-B2:** As a developer, I want uninstall to preserve my config by default so reinstalling later is easy.
- **US-B3:** As a developer doing full cleanup, I want `--remove-config` to delete everything.

#### Command Signature

```
ghm uninstall [--dry-run] [--remove-config]
```

| Flag | Description |
|---|---|
| `--dry-run` | Preview actions without making changes |
| `--remove-config` | Also remove `.githooksrc.yml` and `.githooks/` |

#### Decision Tree (per hook)

```
Hook path P exists? (os.Lstat)
├── NO → Skip: "not present"
└── YES
    Is P a symlink?
    ├── NO (regular file) → Skip: "regular file, not managed by ghm"
    └── YES
        Target resolves to ghmPath?
        ├── NO → Skip: "symlink to <target>, not managed by ghm"
        └── YES
            Remove symlink → "Removed: H"
            P.bak exists?
            ├── YES → Rename P.bak to P → "Restored: H (from backup)"
            └── NO → "No backup to restore for: H"
```

#### What Gets Removed vs. Preserved

| Artifact | Default `ghm uninstall` | `--remove-config` |
|---|---|---|
| `.git/hooks/<name>` symlinks to ghm | **Removed** | **Removed** |
| `.git/hooks/<name>` not pointing to ghm | Preserved | Preserved |
| `.git/hooks/<name>.bak` backup files | **Restored** to original name | **Restored** |
| `.githooksrc.yml` | **Preserved** | **Removed** |
| `.githooks/` directory | **Preserved** | **Removed** |
| `.git/hooks/` directory | Preserved (git's) | Preserved |

#### Summary Output

```
Uninstall complete: 19 hooks removed, 3 restored from backup, 0 skipped.
```

Without `--remove-config`:
```
Preserved: .githooksrc.yml (use --remove-config to remove)
Preserved: .githooks/ (use --remove-config to remove)
```

#### Cobra Help Text

```go
var uninstallCmd = &cobra.Command{
    Use:   "uninstall",
    Short: "Remove githookd hooks from the current repository",
    Long: `Remove all githookd-managed hook symlinks from the current Git repository.
Hooks not managed by ghm are left untouched. Backed-up hooks are restored.
By default, .githooksrc.yml and .githooks/ are preserved.`,
}
```

#### Acceptance Criteria

| # | Criterion |
|---|-----------|
| AC-1 | Removes all 19 ghm symlinks |
| AC-2 | Preserves `.githooksrc.yml` and `.githooks/` by default |
| AC-3 | `--remove-config` deletes config and hooks dir |
| AC-4 | Restores `.bak` files to original names |
| AC-5 | No ghm hooks installed → "Nothing to uninstall" (exit 0) |
| AC-6 | `--dry-run` shows plan, zero changes |
| AC-7 | Not a git repo → exit 1 |
| AC-8 | Non-ghm symlinks and regular files are never touched |

#### Unit Test Examples

```go
func TestUninstallRemovesGhmSymlinks(t *testing.T)
func TestUninstallRestoresBackups(t *testing.T)
func TestUninstallPreservesNonGhmHooks(t *testing.T)
func TestUninstallPreservesConfigByDefault(t *testing.T)
func TestUninstallRemoveConfigFlag(t *testing.T)
func TestUninstallNotGitRepo(t *testing.T)
```

---

## 8. Example Workflows

### New Project Setup
```
$ cd my-project && git init
$ ghm install
  Created .githooks/ directory.
  Created .githooksrc.yml file.
  Installing hooks... 19 installed, 0 skipped, 0 backed up.

$ ghm hooks add pre-commit --run "go vet ./..." --description "Go vet"
  Added command to hook "pre-commit": go vet ./...

$ ghm hooks add pre-commit --run "go test ./..." --description "Go tests"
  Added command to hook "pre-commit": go test ./...

$ ghm hooks list
  pre-commit (2 commands)
    [enabled]  go vet ./...    Go vet
    [enabled]  go test ./...   Go tests
```

### Temporarily Disable a Hook
```
$ ghm hooks disable pre-commit --run "go test ./..."
  Disabled command for hook "pre-commit": go test ./...

$ ghm hooks list
  pre-commit (2 commands)
    [enabled]  go vet ./...    Go vet
    [disabled] go test ./...   Go tests

$ git commit -m "quick fix"   # only go vet runs, tests skipped

$ ghm hooks enable pre-commit --run "go test ./..."
  Enabled command for hook "pre-commit": go test ./...
```

### Remove a Hook Command
```
$ ghm hooks remove pre-commit --run "go vet ./..."
  Removed command from hook "pre-commit": go vet ./...
```

### Full Uninstall and Reinstall
```
$ ghm uninstall
  19 hooks removed, 0 restored from backup, 0 skipped.
  Preserved: .githooksrc.yml (use --remove-config to remove)

$ ghm install
  .githooksrc.yml file already exists.
  19 hooks installed, 0 skipped, 0 backed up.
```

---

## 9. New Package Structure

```
internal/
  config/
    config.go          # Updated: Enabled field, Save(), Resolve(), BoolPtr()
    config_test.go     # Updated: validation and resolution tests
    validation.go      # NEW: hook name validation, Levenshtein distance
  git/
    git.go             # Updated: git rev-parse for GetRepoRoot/GetHooksDir
    git_test.go        # Updated: $GIT_DIR and core.hooksPath tests
  logging/
    logging.go         # NEW: slog setup, ParseLevel()
    logging_test.go
  runner/
    runner.go          # NEW: extracted execution logic from run.go
    errors.go          # NEW: HookError type and FormatReport()
    runner_test.go

cmd/ghm/cmd/
    hooks.go           # NEW: parent command, validateHookName()
    hooks_add.go       # NEW
    hooks_remove.go    # NEW
    hooks_enable.go    # NEW
    hooks_disable.go   # NEW
    hooks_list.go      # NEW
    hooks_test.go      # NEW
    uninstall.go       # NEW
    uninstall_test.go  # NEW
    install.go         # MODIFIED: refactored, --dry-run, --force
    install_test.go    # NEW
    run.go             # MODIFIED: MinimumNArgs, cmd.Dir, structured errors, timeout
```

---

## 10. Implementation Sequencing

| Phase | Work | Rationale |
|---|---|---|
| 1 | Config layer: `Enabled`, `Save()`, `Resolve()`, validation | Everything depends on this |
| 2 | Logging: `internal/logging`, replace `fmt.Printf` | All subsequent work benefits |
| 3 | Execution fixes: `ExactArgs` bug, `GetRepoRoot`, `GetHooksDir`, `cmd.Dir`, structured errors | Independent fixes, can parallelize |
| 4 | Timeout enforcement: `context.WithTimeout` + `CommandContext` | Requires stable execution loop |
| 5 | Hook lifecycle: `hooks add/remove/enable/disable/list` | Depends on `Save()` and `Enabled` |
| 6 | Install improvements: idempotency, `--dry-run`, `--force` | Can happen in parallel with Phase 5 |
| 7 | Uninstall command | Depends on install refactoring for shared helpers |

---

## 11. Verification Plan

After implementation, verify end-to-end:

1. `go test ./...` — all unit tests pass
2. `go build -o ghm ./cmd/ghm` — binary compiles
3. In a test repo:
   - `ghm install` → 19 hooks installed
   - `ghm install` again → 19 skipped
   - `ghm hooks add pre-commit --run "echo hello"` → config updated
   - `ghm hooks list` → shows the hook
   - `ghm hooks disable pre-commit` → config shows `enabled: false`
   - `git commit --allow-empty -m "test"` → "echo hello" is skipped
   - `ghm hooks enable pre-commit` → re-enabled
   - `git commit --allow-empty -m "test2"` → "echo hello" runs
   - `ghm hooks remove pre-commit --run "echo hello"` → removed
   - `ghm uninstall` → hooks removed, config preserved
   - `ghm install --force` → all hooks reinstalled
4. Timeout test: add a hook with `run: "sleep 10"`, set `timeout: 2s`, verify it's killed with structured error
5. Logging test: set `log_level: debug`, verify verbose output on stderr
