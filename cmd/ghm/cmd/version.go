package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

// SetVersionInfo is called by main() to inject build-time version data.
func SetVersionInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of githookd",
	Long:  `Display the version, git commit, and build date of this ghm binary.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ghm %s\n", buildVersion)
		fmt.Printf("  commit: %s\n", buildCommit)
		fmt.Printf("  built:  %s\n", buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
