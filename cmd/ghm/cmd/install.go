package cmd

import (
	"fmt"
	"githookd/internal/config"
	"githookd/internal/git"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

const (
	githooksDir       = ".githooks"
	configFile        = ".githooksrc.yml"
	defaultConfigFile = `# githookd configuration
# For more information, see https://github.com/githookd/githookd
`
)

var standardHooks = config.StandardHooks

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install githookd in the current repository",
	Long: `This command installs githookd in the current Git repository.
It creates the .githooks directory and the .githooksrc.yml configuration file,
and sets up the Git hooks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")

		// Verify we're in a git repo
		repoRoot, err := git.GetRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_ = repoRoot

		if dryRun {
			fmt.Println("Dry run mode: no changes will be made.")
		}

		// Create .githooks directory
		if !dryRun {
			if err := os.MkdirAll(githooksDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating %s directory: %v\n", githooksDir, err)
				os.Exit(1)
			}
		}
		fmt.Printf("Created %s directory.\n", githooksDir)

		// Create .githooksrc.yml file
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			if !dryRun {
				if err := os.WriteFile(configFile, []byte(defaultConfigFile), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating %s file: %v\n", configFile, err)
					os.Exit(1)
				}
			}
			fmt.Printf("Created %s file.\n", configFile)
		} else {
			fmt.Printf("%s file already exists.\n", configFile)
		}

		// Install hooks
		hooksDir, err := git.GetHooksDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		ghmPath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		installed, skipped, backedUp, err := doInstall(hooksDir, ghmPath, dryRun, force)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error installing hooks: %v\n", err)
			os.Exit(1)
		}

		verb := ""
		if dryRun {
			verb = "would be "
		}
		fmt.Printf("\nInstallation complete: %d hooks %sinstalled, %d %sskipped, %d %sbacked up.\n",
			installed, verb, skipped, verb, backedUp, verb)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().Bool("dry-run", false, "Preview actions without making changes")
	installCmd.Flags().Bool("force", false, "Reinstall all hooks even if already managed by ghm")
}

// doInstall handles the actual hook installation logic.
func doInstall(hooksDir, ghmPath string, dryRun, force bool) (installed, skipped, backedUp int, err error) {
	// Ensure hooks directory exists
	if !dryRun {
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			return 0, 0, 0, fmt.Errorf("failed to create hooks directory: %w", err)
		}
	}

	// Resolve canonical path of ghm binary
	canonicalGhm, resolveErr := filepath.EvalSymlinks(ghmPath)
	if resolveErr != nil {
		canonicalGhm = ghmPath
	}

	for _, hookName := range standardHooks {
		hookPath := filepath.Join(hooksDir, hookName)

		info, lstatErr := os.Lstat(hookPath)
		if lstatErr != nil {
			// File doesn't exist — create symlink
			if !dryRun {
				if symlinkErr := os.Symlink(ghmPath, hookPath); symlinkErr != nil {
					return installed, skipped, backedUp, fmt.Errorf("failed to create symlink for %s: %w", hookName, symlinkErr)
				}
			}
			fmt.Printf("  Installed: %s\n", hookName)
			installed++
			continue
		}

		// File exists
		if info.Mode()&os.ModeSymlink != 0 {
			// It's a symlink — check if it points to ghm
			target, readlinkErr := os.Readlink(hookPath)
			if readlinkErr == nil {
				canonicalTarget, evalErr := filepath.EvalSymlinks(hookPath)
				if evalErr == nil && canonicalTarget == canonicalGhm {
					// Symlink already points to ghm
					if force {
						if !dryRun {
							os.Remove(hookPath)
							if symlinkErr := os.Symlink(ghmPath, hookPath); symlinkErr != nil {
								return installed, skipped, backedUp, fmt.Errorf("failed to recreate symlink for %s: %w", hookName, symlinkErr)
							}
						}
						fmt.Printf("  Reinstalled: %s (forced)\n", hookName)
						installed++
						continue
					}
					fmt.Printf("  Skipped: %s (already managed by ghm)\n", hookName)
					skipped++
					continue
				}
			}
			// Foreign symlink — back up and replace
			_ = target // suppress unused warning
		}

		// Regular file or foreign symlink — back up
		backupPath := hookPath + ".bak"
		if _, bakErr := os.Stat(backupPath); bakErr == nil {
			// .bak exists, use timestamped name
			backupPath = hookPath + ".bak." + time.Now().Format("20060102150405")
		}

		if !dryRun {
			if renameErr := os.Rename(hookPath, backupPath); renameErr != nil {
				return installed, skipped, backedUp, fmt.Errorf("failed to back up %s: %w", hookName, renameErr)
			}
			if symlinkErr := os.Symlink(ghmPath, hookPath); symlinkErr != nil {
				return installed, skipped, backedUp, fmt.Errorf("failed to create symlink for %s: %w", hookName, symlinkErr)
			}
		}
		fmt.Printf("  Backed up + Installed: %s\n", hookName)
		installed++
		backedUp++
	}

	return installed, skipped, backedUp, nil
}
