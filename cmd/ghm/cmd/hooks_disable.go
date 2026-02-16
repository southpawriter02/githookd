package cmd

import (
	"fmt"
	"os"

	"githookd/internal/config"

	"github.com/spf13/cobra"
)

var hooksDisableCmd = &cobra.Command{
	Use:   "disable <hook-name>",
	Short: "Disable a hook command",
	Long: `Disable a command for the specified Git hook without removing it.
Disabled commands are skipped during execution.
Use --run to target a specific command, or --all to disable all commands.
If the hook has only one command, no flag is needed.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hookName := args[0]
		runFlag, _ := cmd.Flags().GetString("run")
		allFlag, _ := cmd.Flags().GetBool("all")

		if err := config.ValidateHookName(hookName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: unknown hook '%s'\n", hookName)
			os.Exit(1)
		}

		if runFlag != "" && allFlag {
			fmt.Fprintln(os.Stderr, "Error: --run and --all are mutually exclusive")
			os.Exit(1)
		}

		requireConfigFile()
		cfg := loadConfigOrFail()

		commands, ok := cfg.Hooks[hookName]
		if !ok || len(commands) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no commands configured for hook '%s'\n", hookName)
			os.Exit(1)
		}

		// Single-command shorthand
		if runFlag == "" && !allFlag {
			if len(commands) == 1 {
				allFlag = true
			} else {
				fmt.Fprintf(os.Stderr, "Error: hook '%s' has %d commands; use --run <command> or --all\n", hookName, len(commands))
				os.Exit(1)
			}
		}

		changed := false
		for i := range cfg.Hooks[hookName] {
			c := &cfg.Hooks[hookName][i]
			if allFlag || c.Run == runFlag {
				if c.Enabled == nil || *c.Enabled {
					c.Enabled = config.BoolPtr(false)
					changed = true
					fmt.Printf("Disabled command for hook \"%s\": %s\n", hookName, c.Run)
				} else {
					fmt.Printf("Command already disabled for hook \"%s\": %s\n", hookName, c.Run)
				}
				if !allFlag {
					break
				}
			}
		}

		if !allFlag && !changed {
			found := false
			for _, c := range cfg.Hooks[hookName] {
				if c.Run == runFlag {
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintf(os.Stderr, "Error: command not found for hook '%s': %s\n", hookName, runFlag)
				os.Exit(1)
			}
		}

		saveConfigOrFail(cfg)
		return nil
	},
}

func init() {
	hooksCmd.AddCommand(hooksDisableCmd)
	hooksDisableCmd.Flags().StringP("run", "r", "", "Specific command to disable")
	hooksDisableCmd.Flags().BoolP("all", "a", false, "Disable all commands under this hook")
}
