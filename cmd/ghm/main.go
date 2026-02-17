package main

import (
	"fmt"
	"githookd/cmd/ghm/cmd"
	"os"
	"path/filepath"
	"strings"
)

// Build-time variables — set by GoReleaser via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Expose version info to the cmd package
	cmd.SetVersionInfo(version, commit, date)

	exeName := filepath.Base(os.Args[0])
	exeName = strings.TrimSuffix(exeName, ".exe")

	// Quick --version flag check (before cobra dispatch)
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Printf("ghm %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	if exeName != "main" && exeName != "ghm" {
		// Running as a hook
		args := []string{"run", exeName}
		args = append(args, os.Args[1:]...)
		cmd.ExecuteWithArgs(args)
	} else {
		// Running as ghm
		cmd.Execute()
	}
}
