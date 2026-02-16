package cmd

import (
	"fmt"
	"githookd/internal/config"
	"os"

	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage hook commands",
	Long: `Manage the hook commands configured in .githooksrc.yml.
Use subcommands to add, remove, enable, disable, or list hooks.`,
}

func init() {
	rootCmd.AddCommand(hooksCmd)
}

// requireGitRepo checks that we're inside a git repository.
// Prints an error and exits if not.
func requireGitRepo() {
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		// Try parent directories (simplified check — real validation is in git package)
		fmt.Fprintln(os.Stderr, "Error: not a git repository")
		os.Exit(1)
	}
}

// requireConfigFile checks that the config file exists.
// Prints an error with a hint and exits if not.
func requireConfigFile() {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: config file '%s' not found; run 'ghm install' first\n", configFile)
		os.Exit(1)
	}
}

// loadConfigOrFail loads and returns the config, exiting on error.
func loadConfigOrFail() *config.Config {
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse config file: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

// saveConfigOrFail saves the config, exiting on error.
func saveConfigOrFail(cfg *config.Config) {
	if err := config.Save(configFile, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save config file: %v\n", err)
		os.Exit(1)
	}
}
