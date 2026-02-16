package cmd

import (
	"fmt"
	"os"

	"githookd/internal/config"

	"github.com/spf13/cobra"
)

var hooksRemoveCmd = &cobra.Command{
	Use:   "remove <hook-name>",
	Short: "Remove a command from a hook",
	Long:  `Remove a specific command from the specified Git hook by its exact run value.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hookName := args[0]
		runFlag, _ := cmd.Flags().GetString("run")

		if err := config.ValidateHookName(hookName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: unknown hook '%s'\n", hookName)
			os.Exit(1)
		}

		if runFlag == "" {
			fmt.Fprintln(os.Stderr, "Error: required flag \"run\" not set")
			os.Exit(1)
		}

		requireConfigFile()
		cfg := loadConfigOrFail()

		commands, ok := cfg.Hooks[hookName]
		if !ok || len(commands) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no commands configured for hook '%s'\n", hookName)
			os.Exit(1)
		}

		// Find and remove
		found := false
		var remaining []config.HookCommand
		for _, c := range commands {
			if c.Run == runFlag {
				found = true
				continue
			}
			remaining = append(remaining, c)
		}

		if !found {
			fmt.Fprintf(os.Stderr, "Error: command not found for hook '%s': %s\n", hookName, runFlag)
			os.Exit(1)
		}

		// Clean up empty key
		if len(remaining) == 0 {
			delete(cfg.Hooks, hookName)
		} else {
			cfg.Hooks[hookName] = remaining
		}

		saveConfigOrFail(cfg)
		fmt.Printf("Removed command from hook \"%s\": %s\n", hookName, runFlag)
		return nil
	},
}

func init() {
	hooksCmd.AddCommand(hooksRemoveCmd)
	hooksRemoveCmd.Flags().StringP("run", "r", "", "Exact run value of the command to remove (required)")
	hooksRemoveCmd.MarkFlagRequired("run")
}
