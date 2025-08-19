package cmd

import (
	"fmt"
	"githookd/internal/git"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	githooksDir      = ".githooks"
	configFile       = ".githooksrc.yml"
	defaultConfigFile = `# githookd configuration
# For more information, see https://github.com/githookd/githookd
`
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install githookd in the current repository",
	Long: `This command installs githookd in the current Git repository.
It creates the .githooks directory and the .githooksrc.yml configuration file,
and sets up the Git hooks.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running install command...")

		// Create .githooks directory
		if err := os.MkdirAll(githooksDir, 0755); err != nil {
			fmt.Printf("Error creating %s directory: %v\n", githooksDir, err)
			os.Exit(1)
		}
		fmt.Printf("Created %s directory.\n", githooksDir)

		// Create .githooksrc.yml file
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			if err := os.WriteFile(configFile, []byte(defaultConfigFile), 0644); err != nil {
				fmt.Printf("Error creating %s file: %v\n", configFile, err)
				os.Exit(1)
			}
			fmt.Printf("Created %s file.\n", configFile)
		} else {
			fmt.Printf("%s file already exists.\n", configFile)
		}

		// Install hooks
		if err := installHooks(); err != nil {
			fmt.Printf("Error installing hooks: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

var standardHooks = []string{
	"applypatch-msg",
	"pre-applypatch",
	"post-applypatch",
	"pre-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-push",
	"pre-receive",
	"update",
	"post-receive",
	"post-update",
	"push-to-checkout",
	"pre-auto-gc",
	"post-rewrite",
	"sendemail-validate",
}

func installHooks() error {
	hooksDir, err := git.GetHooksDir()
	if err != nil {
		return err
	}

	ghmPath, err := os.Executable()
	if err != nil {
		return err
	}

	for _, hookName := range standardHooks {
		hookPath := filepath.Join(hooksDir, hookName)

		if _, err := os.Lstat(hookPath); err == nil {
			// File exists, back it up
			backupPath := hookPath + ".bak"
			fmt.Printf("Backing up existing hook: %s -> %s\n", hookPath, backupPath)
			if err := os.Rename(hookPath, backupPath); err != nil {
				return fmt.Errorf("failed to back up existing hook %s: %w", hookName, err)
			}
		}

		fmt.Printf("Installing hook: %s\n", hookPath)
		if err := os.Symlink(ghmPath, hookPath); err != nil {
			return fmt.Errorf("failed to create symlink for hook %s: %w", hookName, err)
		}
	}

	return nil
}
