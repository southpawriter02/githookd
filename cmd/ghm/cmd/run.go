package cmd

import (
	"fmt"
	"githookd/internal/config"
	"githookd/internal/git"
	"githookd/internal/logging"
	"githookd/internal/runner"
	"os"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [hook] [args...]",
	Short: "Run the specified hook",
	Long: `This command runs the specified hook. This is useful for testing your hooks
or for running them in a CI/CD environment.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hookName := args[0]
		hookArgs := args[1:]

		cfg, err := config.Load(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		resolved, errs := cfg.Resolve()
		if len(errs) > 0 {
			fmt.Fprint(os.Stderr, runner.FormatErrors(errs))
			os.Exit(1)
		}

		// Set up logging
		slogLevel := logging.ConfigLevelToSlog(resolved.LogLevel)
		logging.Setup(slogLevel, os.Stderr)

		repoRoot, err := git.GetRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting repo root: %v\n", err)
			os.Exit(1)
		}

		commands, ok := resolved.Hooks[hookName]
		if !ok {
			// No commands for this hook, exit successfully.
			return nil
		}

		if err := runner.RunHook(hookName, commands, repoRoot, hookArgs); err != nil {
			if hookErr, ok := err.(*runner.HookError); ok {
				fmt.Fprint(os.Stderr, hookErr.FormatReport())
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
