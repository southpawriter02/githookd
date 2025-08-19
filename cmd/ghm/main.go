package main

import (
	"githookd/cmd/ghm/cmd"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	exeName := filepath.Base(os.Args[0])
	exeName = strings.TrimSuffix(exeName, ".exe")

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
