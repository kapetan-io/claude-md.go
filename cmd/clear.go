package cmd

import (
	"github.com/kapetan-io/claude-md.go/internal/git"
	"github.com/kapetan-io/claude-md.go/internal/operations"
	"github.com/kapetan-io/claude-md.go/internal/storage"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear CLAUDE.md symlinks from repository",
	Long: `Removes all CLAUDE.md symbolic links from the repository.

This command will:
1. Find all CLAUDE.md symlinks in the repository
2. Remove each symlink

Note: This only removes the symlinks from the repository. The actual files
remain in storage and can be restored later using 'claude-md restore'.`,
	Example: `  # Clear all CLAUDE.md symlinks from current repository
  claude-md clear`,
	RunE: runClear,
}

func init() {
	rootCmd.AddCommand(clearCmd)
}

func runClear(cmd *cobra.Command, args []string) error {
	repo, err := git.FindRepository()
	if err != nil {
		currentOutput.PrintError("Error: %v", err)
		return err
	}

	email, err := repo.GetUserEmail()
	if err != nil {
		currentOutput.PrintError("Error: %v", err)
		return err
	}

	user, err := git.ExtractUserFromEmail(email)
	if err != nil {
		currentOutput.PrintError("Error: %v", err)
		return err
	}

	originURL, err := repo.GetOriginURL()
	if err != nil {
		currentOutput.PrintError("Error: %v", err)
		return err
	}

	repoName, err := git.ExtractRepoName(originURL)
	if err != nil {
		currentOutput.PrintError("Error: %v", err)
		return err
	}

	converter, err := storage.NewPathConverter(user, repoName)
	if err != nil {
		currentOutput.PrintError("Error: %v", err)
		return err
	}

	results := operations.ClearSymlinks(operations.ClearOptions{
		RepoRoot:      repo.RootPath,
		PathConverter: converter,
	})

	var removed, skipped, errors int
	for _, result := range results {
		if result.Success {
			removed++
			currentOutput.PrintSuccess("Removed: %s", result.RepoRelativePath)
		} else if result.Skipped {
			skipped++
			currentOutput.PrintInfo("Skipped %s: %s", result.RepoRelativePath, result.SkipReason)
		} else if result.Error != nil {
			errors++
			currentOutput.PrintError("Error removing %s: %v", result.RepoRelativePath, result.Error)
		}
	}

	if removed == 0 && errors == 0 && skipped == 0 {
		currentOutput.PrintInfo("No CLAUDE.md symlinks found in repository")
	} else {
		currentOutput.PrintInfo("\nSummary: %d removed, %d skipped, %d errors", removed, skipped, errors)
		currentOutput.PrintInfo("Note: Stored files remain in storage")
	}

	return nil
}
