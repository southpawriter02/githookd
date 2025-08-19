package cmd

import (
	"fmt"
	"githookd/internal/config"
	"githookd/internal/git"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [hook]",
	Short: "Run the specified hook",
	Long: `This command runs the specified hook. This is useful for testing your hooks
or for running them in a CI/CD environment.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hookName := args[0]
		hookArgs := args[1:]

		cfg, err := config.Load(configFile)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		commands, ok := cfg.Hooks[hookName]
		if !ok {
			// No commands for this hook, so we can exit successfully.
			os.Exit(0)
		}

		repoRoot, err := git.GetRepoRoot()
		if err != nil {
			fmt.Printf("Error getting repo root: %v\n", err)
			os.Exit(1)
		}

		for _, command := range commands {
			fmt.Printf("Running command: %s\n", command.Run)
			// The script and its arguments are passed to sh.
			script := command.Run + " " + strings.Join(hookArgs, " ")
			cmd := exec.Command("sh", "-c", script)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = append(os.Environ(),
				"GHM_HOOK_NAME="+hookName,
				"GHM_ROOT="+repoRoot,
			)

			if err := cmd.Run(); err != nil {
				fmt.Printf("Command failed: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
