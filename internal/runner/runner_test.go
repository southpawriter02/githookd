package runner

import (
	"strings"
	"testing"
	"time"
)

func TestHookError_Error(t *testing.T) {
	err := &HookError{
		HookName: "pre-commit",
		Command:  "npm run lint",
		ExitCode: 2,
	}

	got := err.Error()
	if !strings.Contains(got, "pre-commit") {
		t.Errorf("Error() = %q, want it to contain 'pre-commit'", got)
	}
	if !strings.Contains(got, "exit code 2") {
		t.Errorf("Error() = %q, want it to contain 'exit code 2'", got)
	}
}

func TestHookError_Error_Timeout(t *testing.T) {
	err := &HookError{
		HookName:   "pre-commit",
		Command:    "sleep 100",
		TimedOut:   true,
		TimeoutDur: 5 * time.Second,
	}

	got := err.Error()
	if !strings.Contains(got, "timed out") {
		t.Errorf("Error() = %q, want it to contain 'timed out'", got)
	}
	if !strings.Contains(got, "5s") {
		t.Errorf("Error() = %q, want it to contain '5s'", got)
	}
}

func TestHookError_FormatReport(t *testing.T) {
	err := &HookError{
		HookName: "pre-commit",
		Command:  "npm run lint",
		ExitCode: 2,
		Stdout:   "some output\n",
		Stderr:   "some error\n",
	}

	report := err.FormatReport()

	checks := []string{
		"HOOK FAILED",
		"pre-commit",
		"npm run lint",
		"Exit Code: 2",
		"STDOUT",
		"some output",
		"STDERR",
		"some error",
	}

	for _, check := range checks {
		if !strings.Contains(report, check) {
			t.Errorf("FormatReport() missing %q", check)
		}
	}
}

func TestHookError_FormatReport_Timeout(t *testing.T) {
	err := &HookError{
		HookName:   "pre-commit",
		Command:    "sleep 100",
		TimedOut:   true,
		TimeoutDur: 5 * time.Second,
	}

	report := err.FormatReport()

	if !strings.Contains(report, "timed out after 5s") {
		t.Errorf("FormatReport() missing timeout status, got:\n%s", report)
	}
	if !strings.Contains(report, "Hint:") {
		t.Errorf("FormatReport() missing timeout hint, got:\n%s", report)
	}
}

func TestHookError_FormatReport_NoOutput(t *testing.T) {
	err := &HookError{
		HookName: "pre-commit",
		Command:  "false",
		ExitCode: 1,
	}

	report := err.FormatReport()
	if strings.Contains(report, "STDOUT") {
		t.Error("FormatReport() should not show STDOUT section when output is empty")
	}
	if strings.Contains(report, "STDERR") {
		t.Error("FormatReport() should not show STDERR section when output is empty")
	}
}

func TestTruncate(t *testing.T) {
	short := "hello"
	if got := truncate(short); got != short {
		t.Errorf("truncate(%q) = %q, want %q", short, got, short)
	}

	long := strings.Repeat("a", maxOutputBytes+100)
	got := truncate(long)
	if len(got) > maxOutputBytes+50 {
		t.Errorf("truncate() output too long: %d bytes", len(got))
	}
	if !strings.Contains(got, "truncated") {
		t.Error("truncate() should append truncation notice")
	}
}

func TestFormatErrors(t *testing.T) {
	errs := []error{
		strings.NewReader("").(*strings.Reader), // This won't work, use fmt.Errorf
	}
	// Use proper errors
	errs = []error{
		&HookError{HookName: "test", Command: "cmd", ExitCode: 1},
	}

	got := FormatErrors(errs)
	if !strings.Contains(got, "invalid configuration") {
		t.Errorf("FormatErrors() = %q, want 'invalid configuration'", got)
	}
}
