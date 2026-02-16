package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"githookd/internal/config"
)

// RunHook executes all enabled commands for a hook in sequence.
// Returns on first failure (abort semantics).
func RunHook(hookName string, commands []config.ResolvedHookCommand, repoRoot string, hookArgs []string) error {
	for _, command := range commands {
		if !command.Enabled {
			slog.Info("Skipping disabled command", "hook", hookName, "command", command.Run)
			continue
		}

		slog.Info("Running command", "hook", hookName, "command", command.Run)
		if command.Description != "" {
			slog.Info("Description", "description", command.Description)
		}

		hookErr := runCommand(hookName, command, repoRoot, hookArgs)
		if hookErr != nil {
			return hookErr
		}
	}
	return nil
}

// runCommand executes a single hook command with timeout and output capture.
func runCommand(hookName string, command config.ResolvedHookCommand, repoRoot string, hookArgs []string) *HookError {
	// Build the shell command with hook arguments
	script := command.Run
	if len(hookArgs) > 0 {
		script += " " + strings.Join(hookArgs, " ")
	}

	// Set up context with timeout
	var ctx context.Context
	var cancel context.CancelFunc

	if command.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), command.Timeout)
		slog.Debug("Timeout set", "duration", command.Timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
		slog.Debug("No timeout (none)")
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", script)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(),
		"GHM_HOOK_NAME="+hookName,
		"GHM_ROOT="+repoRoot,
	)

	slog.Debug("Execution environment",
		"dir", repoRoot,
		"script", script,
	)

	// Stream and capture stdout/stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	slog.Debug("Command completed", "elapsed", elapsed)

	if err != nil {
		hookErr := &HookError{
			HookName: hookName,
			Command:  command.Run,
			Stdout:   stdoutBuf.String(),
			Stderr:   stderrBuf.String(),
		}

		// Check for timeout
		if ctx.Err() == context.DeadlineExceeded {
			hookErr.TimedOut = true
			hookErr.TimeoutDur = command.Timeout
		} else {
			// Extract exit code
			if exitErr, ok := err.(*exec.ExitError); ok {
				hookErr.ExitCode = exitErr.ExitCode()
			} else {
				hookErr.ExitCode = 1
			}
		}

		return hookErr
	}

	return nil
}

// FormatErrors formats multiple config validation errors into a single string.
func FormatErrors(errs []error) string {
	var b strings.Builder
	b.WriteString("Error: invalid configuration in .githooksrc.yml\n\n")
	for _, err := range errs {
		b.WriteString(fmt.Sprintf("  - %s\n", err))
	}
	return b.String()
}
