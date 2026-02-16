package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"githookd/internal/config"

	"github.com/spf13/cobra"
)

var hooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured hooks",
	Long:  `Display all hooks and their commands from .githooksrc.yml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		hookFilter, _ := cmd.Flags().GetString("hook")
		jsonFlag, _ := cmd.Flags().GetBool("json")
		quietFlag, _ := cmd.Flags().GetBool("quiet")

		// Validate hook filter if provided
		if hookFilter != "" {
			if err := config.ValidateHookName(hookFilter); err != nil {
				fmt.Fprintf(os.Stderr, "Error: unknown hook '%s'\n", hookFilter)
				os.Exit(1)
			}
		}

		requireConfigFile()
		cfg := loadConfigOrFail()

		if len(cfg.Hooks) == 0 {
			fmt.Println("No hooks configured.")
			return nil
		}

		// Get sorted hook names
		var hookNames []string
		for name := range cfg.Hooks {
			if hookFilter != "" && name != hookFilter {
				continue
			}
			hookNames = append(hookNames, name)
		}
		sort.Strings(hookNames)

		if len(hookNames) == 0 {
			if hookFilter != "" {
				fmt.Printf("No commands configured for hook '%s'.\n", hookFilter)
			} else {
				fmt.Println("No hooks configured.")
			}
			return nil
		}

		if jsonFlag {
			return printJSON(cfg, hookNames)
		}

		if quietFlag {
			for _, name := range hookNames {
				fmt.Println(name)
			}
			return nil
		}

		// Human-readable output
		for _, name := range hookNames {
			commands := cfg.Hooks[name]
			noun := "commands"
			if len(commands) == 1 {
				noun = "command"
			}
			fmt.Printf("%s (%d %s)\n", name, len(commands), noun)
			for _, c := range commands {
				status := "enabled"
				if !c.IsEnabled() {
					status = "disabled"
				}
				desc := ""
				if c.Description != "" {
					desc = "  " + c.Description
				}
				fmt.Printf("  [%s]  %-40s%s\n", status, c.Run, desc)
			}
			fmt.Println()
		}

		return nil
	},
}

type jsonHookCommand struct {
	Run         string `json:"run"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

func printJSON(cfg *config.Config, hookNames []string) error {
	output := make(map[string][]jsonHookCommand)

	for _, name := range hookNames {
		var cmds []jsonHookCommand
		for _, c := range cfg.Hooks[name] {
			cmds = append(cmds, jsonHookCommand{
				Run:         c.Run,
				Description: c.Description,
				Enabled:     c.IsEnabled(),
			})
		}
		output[name] = cmds
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func init() {
	hooksCmd.AddCommand(hooksListCmd)
	hooksListCmd.Flags().StringP("hook", "H", "", "Filter to a specific hook")
	hooksListCmd.Flags().Bool("json", false, "Output in JSON format")
	hooksListCmd.Flags().BoolP("quiet", "q", false, "Show only hook names")
}
