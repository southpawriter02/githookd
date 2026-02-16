package cmd

import (
	"fmt"
	"githookd/internal/git"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove githookd hooks from the current repository",
	Long: `Remove all githookd-managed hook symlinks from the current Git repository.
Hooks not managed by ghm are left untouched. Backed-up hooks are restored.
By default, .githooksrc.yml and .githooks/ are preserved.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		removeConfig, _ := cmd.Flags().GetBool("remove-config")

		// Verify we're in a git repo
		if _, err := git.GetRepoRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

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

		if dryRun {
			fmt.Println("Dry run mode: no changes will be made.")
		}

		removed, restored, skipped, err := doUninstall(hooksDir, ghmPath, dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if removed == 0 && restored == 0 {
			fmt.Println("Nothing to uninstall.")
		} else {
			verb := ""
			if dryRun {
				verb = "would be "
			}
			fmt.Printf("\nUninstall complete: %d hooks %sremoved, %d %srestored from backup, %d %sskipped.\n",
				removed, verb, restored, verb, skipped, verb)
		}

		// Handle config removal
		if removeConfig {
			if !dryRun {
				if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
					fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", configFile, err)
				} else if err == nil {
					fmt.Printf("Removed: %s\n", configFile)
				}
				if err := os.RemoveAll(githooksDir); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", githooksDir, err)
				} else {
					fmt.Printf("Removed: %s/\n", githooksDir)
				}
			} else {
				fmt.Printf("Would remove: %s\n", configFile)
				fmt.Printf("Would remove: %s/\n", githooksDir)
			}
		} else {
			if _, err := os.Stat(configFile); err == nil {
				fmt.Printf("Preserved: %s (use --remove-config to remove)\n", configFile)
			}
			if _, err := os.Stat(githooksDir); err == nil {
				fmt.Printf("Preserved: %s/ (use --remove-config to remove)\n", githooksDir)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().Bool("dry-run", false, "Preview actions without making changes")
	uninstallCmd.Flags().Bool("remove-config", false, "Also remove .githooksrc.yml and .githooks/")
}

// doUninstall removes ghm-managed hook symlinks and restores backups.
func doUninstall(hooksDir, ghmPath string, dryRun bool) (removed, restored, skipped int, err error) {
	canonicalGhm, resolveErr := filepath.EvalSymlinks(ghmPath)
	if resolveErr != nil {
		canonicalGhm = ghmPath
	}

	for _, hookName := range standardHooks {
		hookPath := filepath.Join(hooksDir, hookName)

		info, lstatErr := os.Lstat(hookPath)
		if lstatErr != nil {
			// Not present
			continue
		}

		// Check if it's a symlink
		if info.Mode()&os.ModeSymlink == 0 {
			// Regular file, not managed by ghm
			fmt.Printf("  Skipped: %s (regular file, not managed by ghm)\n", hookName)
			skipped++
			continue
		}

		// It's a symlink — check target
		canonicalTarget, evalErr := filepath.EvalSymlinks(hookPath)
		if evalErr != nil || canonicalTarget != canonicalGhm {
			target, _ := os.Readlink(hookPath)
			fmt.Printf("  Skipped: %s (symlink to %s, not managed by ghm)\n", hookName, target)
			skipped++
			continue
		}

		// It's a ghm symlink — remove it
		if !dryRun {
			if removeErr := os.Remove(hookPath); removeErr != nil {
				return removed, restored, skipped, fmt.Errorf("failed to remove %s: %w", hookName, removeErr)
			}
		}
		fmt.Printf("  Removed: %s\n", hookName)
		removed++

		// Check for backup to restore
		backupPath := hookPath + ".bak"
		if _, bakErr := os.Stat(backupPath); bakErr == nil {
			if !dryRun {
				if renameErr := os.Rename(backupPath, hookPath); renameErr != nil {
					fmt.Fprintf(os.Stderr, "  Warning: failed to restore backup for %s: %v\n", hookName, renameErr)
					continue
				}
			}
			fmt.Printf("  Restored: %s (from backup)\n", hookName)
			restored++
		}
	}

	return removed, restored, skipped, nil
}
