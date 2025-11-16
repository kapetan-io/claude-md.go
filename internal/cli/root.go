package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claude-md",
	Short: "Manage CLAUDE.md files across git repositories",
	Long: `claude-md is a CLI tool for managing CLAUDE.md files using centralized storage with symlinks.
It enables saving CLAUDE.md files from anywhere in a repository to a structured storage location
and restoring them as symlinks.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands here as they are created
}
