package cli

import (
	"github.com/kapetan-io/claude-md.go/internal/files"
	"github.com/kapetan-io/claude-md.go/internal/git"
	"github.com/kapetan-io/claude-md.go/internal/operations"
	"github.com/kapetan-io/claude-md.go/internal/storage"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore CLAUDE.md files from storage",
	Long: `Creates symlinks for all stored CLAUDE.md files in the repository.

This command will:
1. Find all stored CLAUDE.md files for this repository
2. Create symlinks in the appropriate locations pointing to storage

Files that already exist (regular files or symlinks) are skipped with a warning.
If a parent directory doesn't exist, the file is skipped with a warning.`,
	Example: `  # Restore all CLAUDE.md files for current repository
  claude-md restore`,
	RunE: runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, args []string) error {
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

	repoStorageDir := converter.GetRepoStorageDir()
	storedFiles, err := files.FindStoredFiles(repoStorageDir, converter)
	if err != nil {
		currentOutput.PrintError("Error finding stored files: %v", err)
		return err
	}

	if len(storedFiles) == 0 {
		currentOutput.PrintInfo("No stored CLAUDE.md files found for this repository")
		return nil
	}

	results := operations.RestoreFiles(storedFiles, operations.RestoreOptions{
		RepoRoot: repo.RootPath,
	})

	var restored, skipped, warnings int
	for _, result := range results {
		if result.Success {
			restored++
			currentOutput.PrintSuccess("Restored: %s", result.RepoRelativePath)
		} else if result.Skipped {
			skipped++
			if result.Warning != "" {
				warnings++
				currentOutput.PrintInfo("Warning: %s", result.Warning)
			}
		}
	}

	currentOutput.PrintInfo("\nSummary: %d restored, %d skipped (%d warnings)", restored, skipped, warnings)

	return nil
}
