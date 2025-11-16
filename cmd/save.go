package cmd

import (
	"github.com/kapetan-io/claude-md.go/internal/files"
	"github.com/kapetan-io/claude-md.go/internal/git"
	"github.com/kapetan-io/claude-md.go/internal/operations"
	"github.com/kapetan-io/claude-md.go/internal/storage"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save CLAUDE.md files to storage",
	Long: `Finds all CLAUDE.md files in the repository and converts them to symlinks pointing to centralized storage.

This command will:
1. Find all CLAUDE.md files (case-insensitive) in the repository
2. Copy each file to ~/.claude/claude-md/<user>/<repo>/
3. Replace the original file with a symlink to the stored copy

Files already converted to symlinks are skipped. If a storage file already exists,
the operation is skipped with a warning.`,
	Example: `  # Save all CLAUDE.md files in current repository
  claude-md save`,
	RunE: runSave,
}

func init() {
	rootCmd.AddCommand(saveCmd)
}

func runSave(cmd *cobra.Command, args []string) error {
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

	claudeFiles, err := files.FindClaudeFiles(repo.RootPath)
	if err != nil {
		currentOutput.PrintError("Error finding CLAUDE.md files: %v", err)
		return err
	}

	if len(claudeFiles) == 0 {
		currentOutput.PrintInfo("No CLAUDE.md files found in repository")
		return nil
	}

	results := operations.SaveFiles(claudeFiles, operations.SaveOptions{
		RepoRoot:      repo.RootPath,
		PathConverter: converter,
	})

	var saved, skipped, errors int
	for _, result := range results {
		if result.Success {
			saved++
			currentOutput.PrintSuccess("Saved: %s", result.RepoRelativePath)
		} else if result.Skipped {
			skipped++
			if result.Warning != "" {
				currentOutput.PrintInfo("Warning: %s", result.Warning)
			}
		} else if result.Error != nil {
			errors++
			currentOutput.PrintError("Error: %s", result.Warning)
		}
	}

	currentOutput.PrintInfo("\nSummary: %d saved, %d skipped, %d errors", saved, skipped, errors)

	return nil
}
