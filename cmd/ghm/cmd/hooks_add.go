package cmd

import (
	"fmt"
	"githookd/internal/config"
	"os"

	"github.com/spf13/cobra"
)

var hooksAddCmd = &cobra.Command{
	Use:   "add <hook-name>",
	Short: "Add a command to a hook",
	Long: `Add a new command to the specified Git hook.
The command will be appended to the end of the hook's command list.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hookName := args[0]

		runFlag, _ := cmd.Flags().GetString("run")
		descFlag, _ := cmd.Flags().GetString("description")

		// Validate hook name
		if err := config.ValidateHookName(hookName); err != nil {
			suggestion := config.SuggestHookName(hookName)
			msg := fmt.Sprintf("Error: unknown hook '%s'", hookName)
			if suggestion != "" {
				msg += fmt.Sprintf("; did you mean '%s'?", suggestion)
			}
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}

		// Validate run flag
		if runFlag == "" {
			fmt.Fprintln(os.Stderr, "Error: required flag \"run\" not set")
			os.Exit(1)
		}

		requireConfigFile()
		cfg := loadConfigOrFail()

		// Initialize hooks map if nil
		if cfg.Hooks == nil {
			cfg.Hooks = make(map[string][]config.HookCommand)
		}

		// Check for duplicate
		for _, existing := range cfg.Hooks[hookName] {
			if existing.Run == runFlag {
				fmt.Fprintf(os.Stderr, "Error: command already exists for hook '%s': %s\n", hookName, runFlag)
				os.Exit(1)
			}
		}

		// Append new command
		cfg.Hooks[hookName] = append(cfg.Hooks[hookName], config.HookCommand{
			Run:         runFlag,
			Description: descFlag,
		})

		saveConfigOrFail(cfg)
		fmt.Printf("Added command to hook \"%s\": %s\n", hookName, runFlag)
		return nil
	},
}

func init() {
	hooksCmd.AddCommand(hooksAddCmd)
	hooksAddCmd.Flags().StringP("run", "r", "", "The shell command to execute (required)")
	hooksAddCmd.Flags().StringP("description", "d", "", "Human-readable description")
	hooksAddCmd.MarkFlagRequired("run")
}
