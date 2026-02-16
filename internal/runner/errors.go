package runner

import (
	"fmt"
	"strings"
	"time"
)

const maxOutputBytes = 10240 // 10KB

// HookError contains structured information about a hook command failure.
type HookError struct {
	HookName   string
	Command    string
	ExitCode   int
	Stdout     string
	Stderr     string
	TimedOut   bool
	TimeoutDur time.Duration
}

// Error implements the error interface.
func (e *HookError) Error() string {
	if e.TimedOut {
		return fmt.Sprintf("hook %q timed out after %s: %s", e.HookName, e.TimeoutDur, e.Command)
	}
	return fmt.Sprintf("hook %q failed (exit code %d): %s", e.HookName, e.ExitCode, e.Command)
}

// FormatReport returns a structured, human-readable error report.
func (e *HookError) FormatReport() string {
	var b strings.Builder

	b.WriteString("===========================================================\n")
	b.WriteString(" HOOK FAILED\n")
	b.WriteString("===========================================================\n")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Hook:      %s\n", e.HookName))
	b.WriteString(fmt.Sprintf("  Command:   %s\n", e.Command))

	if e.TimedOut {
		b.WriteString(fmt.Sprintf("  Status:    timed out after %s\n", e.TimeoutDur))
	} else {
		b.WriteString(fmt.Sprintf("  Exit Code: %d\n", e.ExitCode))
	}

	stdout := truncate(e.Stdout)
	if stdout != "" {
		b.WriteString("\n")
		b.WriteString("-----------------------------------------------------------\n")
		b.WriteString(" STDOUT\n")
		b.WriteString("-----------------------------------------------------------\n")
		b.WriteString("  " + strings.ReplaceAll(strings.TrimRight(stdout, "\n"), "\n", "\n  ") + "\n")
	}

	stderr := truncate(e.Stderr)
	if stderr != "" {
		b.WriteString("\n")
		b.WriteString("-----------------------------------------------------------\n")
		b.WriteString(" STDERR\n")
		b.WriteString("-----------------------------------------------------------\n")
		b.WriteString("  " + strings.ReplaceAll(strings.TrimRight(stderr, "\n"), "\n", "\n  ") + "\n")
	}

	b.WriteString("\n")
	b.WriteString("===========================================================\n")

	if e.TimedOut {
		b.WriteString(fmt.Sprintf("\n  Hint: To increase the timeout for this command, add 'timeout: 60s'\n"))
		b.WriteString("        to the command entry in .githooksrc.yml, or increase the\n")
		b.WriteString("        global 'timeout' setting.\n")
	}

	return b.String()
}

// truncate limits a string to maxOutputBytes, appending a truncation notice if needed.
func truncate(s string) string {
	if len(s) <= maxOutputBytes {
		return s
	}
	return s[:maxOutputBytes] + "\n... (output truncated)"
}
