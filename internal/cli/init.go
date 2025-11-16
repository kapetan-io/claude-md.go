package cli

import (
	"os"

	"github.com/kapetan-io/claude-md.go/internal/git"
	"github.com/kapetan-io/claude-md.go/internal/storage"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize claude-md storage directory",
	Long: `Creates the ~/.claude/claude-md/<user>/<repo> directory structure for storing CLAUDE.md files.

The storage location is determined by:
- User: extracted from git config user.email (part before @)
- Repository: extracted from git remote origin URL

This command must be run from within a git repository with an origin remote configured.`,
	Example: `  # Initialize storage for current repository
  claude-md init`,
	RunE:    runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
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

	storageDir := converter.GetRepoStorageDir()

	if info, err := os.Stat(storageDir); err == nil && info.IsDir() {
		currentOutput.PrintInfo("Storage directory already exists: %s", storageDir)
		currentOutput.PrintInfo("User: %s", user)
		return nil
	}

	if err := converter.EnsureStorageDir(); err != nil {
		currentOutput.PrintError("Error: %v", err)
		return err
	}

	currentOutput.PrintSuccess("Created storage directory: %s", storageDir)
	currentOutput.PrintInfo("User: %s", user)

	return nil
}
