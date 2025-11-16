# CLAUDE.md Management CLI Tool Implementation Plan

## Overview

This plan describes the implementation of `claude-md`, a CLI tool for managing CLAUDE.md files across git repositories using a centralized storage system with symlinks. The tool enables saving CLAUDE.md files from anywhere in a repository to a structured storage location (`~/.claude/claude-md/<user>/<repo>.git/`) and restoring them as symlinks, ensuring changes persist across repository cleanup operations.

## Current State Analysis

This is a greenfield project starting at `/Users/thrawn/Development/claude-md` with:
- Go module: `github.com/kapetan-io/claude-md.go`
- Go version: 1.24.4
- No existing code structure

### Key Requirements Discovered:
- Must work from anywhere within a git repository
- Git repository must have an `origin` remote configured
- User identification from `git config user.email` (trimmed at `@`)
- Repository name extracted from `origin` remote URL (used as-is from URL, e.g., `repo.git` if URL ends with `repo.git`, or `repo` if URL ends with `repo`)
- Path conversion uses `~` as separator for ALL path components (e.g., `source/go/api/CLAUDE.md` → `source~go~api~CLAUDE.md`)
- Directory names containing `~` are not supported (must error)
- Storage structure: `~/.claude/claude-md/<user>/<repo-name>/<path-with-tildes>~CLAUDE.md`
- Root-level CLAUDE.md stored as: `~/.claude/claude-md/<user>/<repo-name>/CLAUDE.md`
- File matching is case-insensitive (matches `CLAUDE.md`, `claude.md`, `Claude.MD`, etc.)
- Symlinks use absolute paths to storage files for reliability
- Conflicts during save/restore: skip with warning, showing both file paths and reason

## Desired End State

A working CLI tool with the following commands:

1. **`claude-md init`** - Creates storage directory structure, reports user name
2. **`claude-md save`** - Finds all CLAUDE.md files in repo, converts to symlinks, reports results
3. **`claude-md restore`** - Restores all saved CLAUDE.md files as symlinks, reports results
4. **`claude-md clear`** - Removes all CLAUDE.md symlinks from repo, reports results

### Verification:
- All commands return exit code 0 on success, non-zero on errors
- Run `go build` successfully
- Run `go test ./...` with all tests passing
- Manual testing of each command in a test git repository

## What We're NOT Doing

- Auto-launching Claude CLI after restore (removed from original scope)
- Supporting non-git directories
- Supporting repositories without `origin` remote
- Supporting directory names containing `~` character
- Deleting stored files when running `clear` (only removes symlinks)
- Supporting multiple git hosting providers' authentication (just parse URL)
- Interactive prompts or confirmations (all commands execute directly)

## Key Design Decisions

These decisions were clarified during requirements gathering and must be followed:

1. **Storage Filename Format**: Use `~` as separator for ALL path components, including the final CLAUDE.md
   - Example: `source/go/api/CLAUDE.md` → `source~go~api~CLAUDE.md`
   - Root level: `CLAUDE.md` → `CLAUDE.md` (no conversion needed)

2. **Repository Name Handling**: Extract repo name from origin URL and use exactly as provided
   - If URL is `git@github.com:user/repo.git` → use `repo.git`
   - If URL is `https://github.com/user/repo` → use `repo`
   - Do NOT add or remove `.git` extension

3. **Conflict Resolution**: Skip with warning, showing both file paths
   - **Save**: If storage file already exists, skip and warn user
   - **Restore**: If regular file exists at target location, skip and warn user
   - Both show repo path and storage path in warning message

4. **Case Sensitivity**: File matching is case-insensitive
   - Match `CLAUDE.md`, `claude.md`, `Claude.MD`, `ClaUde.Md`, etc.
   - Use `strings.EqualFold` for comparison

5. **Symlink Paths**: Always use absolute paths
   - Symlinks point to absolute path in storage: `/Users/user/.claude/claude-md/.../file`
   - More reliable than relative paths, works from any directory

6. **Git Operations**: Use `os/exec` to run git commands directly
   - No external git libraries needed
   - Execute `git config --get remote.origin.url` and `git config --get user.email`
   - Simpler, more reliable, no dependency management

7. **User Extraction**: Parse email address before `@` symbol
   - `john.doe@example.com` → `john.doe`
   - Error if email is empty or doesn't contain `@`

## Implementation Approach

Build the tool in phases, starting with core infrastructure and git operations, then implementing each command incrementally. Use Cobra for CLI framework, standard library for file operations, and `os/exec` package to run git commands directly (no external git libraries needed). Follow TDD approach with tests written before implementation.

## Phase 1: Project Setup & Core Infrastructure

### Overview
Set up the Cobra CLI structure and implement core utilities for git repository detection, config reading, and path manipulation. This phase provides the foundation for all commands.

### Changes Required:

#### 1. Project Structure & Dependencies
**Files**: Create initial project structure

```
claude-md/
├── cmd/
│   └── root.go          # Root command setup
├── internal/
│   ├── git/
│   │   └── repo.go      # Git operations
│   └── storage/
│       └── paths.go     # Path manipulation utilities
├── main.go              # Entry point
└── go.mod               # Dependencies
```

**Dependencies to add**:
```bash
go get github.com/spf13/cobra@latest
```

**Note**: Git operations will use `os/exec` package to run git commands directly - no additional git library dependencies needed.

#### 2. Git Repository Operations
**File**: `internal/git/repo.go`
**Changes**: Implement git repository detection and config reading

```go
package git

// Repository represents a git repository and its configuration
type Repository struct {
    RootPath   string
    RepoName   string
    UserEmail  string
}

// FindRepository detects the git repository from the current or parent directories
func FindRepository() (*Repository, error)

// GetOriginURL retrieves the origin remote URL
func (r *Repository) GetOriginURL() (string, error)

// GetUserEmail retrieves user.email from git config
func (r *Repository) GetUserEmail() (string, error)

// ExtractRepoName extracts repository name from git remote URL
// Handles both SSH (git@github.com:user/repo.git) and HTTPS (https://github.com/user/repo.git)
// Returns repo name as-is from URL (e.g., "kapetan.git" if URL ends with "kapetan.git", "kapetan" if URL ends with "kapetan")
func ExtractRepoName(remoteURL string) (string, error)

// ExtractUserFromEmail extracts username from email by trimming @ and everything after
// Example: "john.doe@example.com" → "john.doe"
// Returns error if email is empty or doesn't contain @
func ExtractUserFromEmail(email string) (string, error)
```

**Function Responsibilities:**
- `FindRepository`: Walk up from current directory to find `.git` directory, return error if not found
- `GetOriginURL`: Execute `git config --get remote.origin.url` using `os/exec.Command`, return error if origin remote doesn't exist
- `GetUserEmail`: Execute `git config --get user.email` using `os/exec.Command`, return error if not configured
- `ExtractRepoName`: Parse URL using string operations to extract final path component after last `/` or `:`, preserve extension as-is
- `ExtractUserFromEmail`: Split on `@` and return first part, return error if email is empty or doesn't contain `@`
- All git operations error if not in git repo, no origin remote, or config values missing/invalid

#### 3. Storage Path Utilities
**File**: `internal/storage/paths.go`
**Changes**: Implement path conversion and storage location logic

```go
package storage

// PathConverter handles conversion between repository paths and storage paths
type PathConverter struct {
    StorageRoot string  // ~/.claude/claude-md/<user>
    RepoName    string  // Repository name as extracted from origin URL (e.g., "kapetan.git" or "kapetan")
}

// NewPathConverter creates a new path converter
// Returns error if home directory cannot be determined
func NewPathConverter(user, repoName string) (*PathConverter, error)

// ConvertToStorageName converts a repository relative path to storage filename
// Example: "source/go/api/CLAUDE.md" → "source~go~api~CLAUDE.md"
// Example: "CLAUDE.md" → "CLAUDE.md"
func (pc *PathConverter) ConvertToStorageName(repoRelativePath string) (string, error)

// ConvertToRepoPath converts a storage filename back to repository relative path
// Example: "source~go~api~CLAUDE.md" → "source/go/api/CLAUDE.md"
// Example: "CLAUDE.md" → "CLAUDE.md"
func (pc *PathConverter) ConvertToRepoPath(storageFilename string) string

// GetStoragePath returns full path to stored file
func (pc *PathConverter) GetStoragePath(repoRelativePath string) (string, error)

// GetRepoStorageDir returns the directory where this repo's files are stored
// Returns: ~/.claude/claude-md/<user>/<repo-name>/
func (pc *PathConverter) GetRepoStorageDir() string

// ValidatePath checks if path contains invalid characters (like ~)
func ValidatePath(path string) error

// EnsureStorageDir creates storage directory if it doesn't exist
func (pc *PathConverter) EnsureStorageDir() error
```

**Function Responsibilities:**
- `NewPathConverter`: Expand home directory using `os.UserHomeDir()`, construct storage root, return error if home dir unavailable
- `ConvertToStorageName`: Split path on `/`, validate no component contains `~`, join with `~` for all components (e.g., `source/go/CLAUDE.md` → `source~go~CLAUDE.md`), handle root case (CLAUDE.md stays as CLAUDE.md)
- `ConvertToRepoPath`: Split storage filename on `~`, join with `/` to reconstruct original path
- `GetStoragePath`: Combine storage root, repo name, and converted filename
- `GetRepoStorageDir`: Return `~/.claude/claude-md/<user>/<repo-name>/`
- `ValidatePath`: Error if any directory component contains `~` character
- `EnsureStorageDir`: Use `os.MkdirAll` with `0755` permissions

#### 4. Root Command Setup
**File**: `cmd/root.go`
**Changes**: Create Cobra root command structure

```go
package cmd

// Execute runs the root command
func Execute() error
```

**Function Responsibilities:**
- Initialize Cobra root command
- Set up global flags if needed
- Configure command hierarchy
- Handle command execution errors

**File**: `main.go`
**Changes**: Entry point

```go
package main

func main()
```

**Function Responsibilities:**
- Call `cmd.Execute()`
- Exit with appropriate exit code

### Testing Requirements:

```go
// internal/git/repo_test.go
func TestFindRepository(t *testing.T)
func TestExtractRepoName(t *testing.T)
func TestExtractUserFromEmail(t *testing.T)

// internal/storage/paths_test.go
func TestConvertToStorageName(t *testing.T)
func TestConvertToRepoPath(t *testing.T)
func TestValidatePath(t *testing.T)
func TestGetStoragePath(t *testing.T)
```

**Test Objectives:**
- `TestFindRepository`: Verify detection from subdirectories, error when not in repo
- `TestExtractRepoName`: Test SSH and HTTPS URL formats, with and without .git extension
- `TestExtractUserFromEmail`: Test various email formats
- `TestConvertToStorageName`: Test path conversion, root case, error on `~` in path
- `TestConvertToRepoPath`: Test reverse conversion, root case
- `TestValidatePath`: Test rejection of paths containing `~`
- `TestGetStoragePath`: Test full path generation

**Context for Implementation:**
- Use `filepath.Walk` pattern for finding git root (walk up parent directories)
- Use `os/exec` package to run git commands or `tcnksm/go-gitconfig` library
- Use `filepath.Join` for path operations
- Use `os.UserHomeDir()` for getting home directory
- Use table-driven tests for URL parsing and path conversion

### Validation
- [ ] Run: `go build -o claude-md .`
- [ ] Run: `go test ./...`
- [ ] Verify: All tests pass, binary builds successfully

## Phase 2: Init Command

### Overview
Implement the `init` command that creates the storage directory structure and reports the user name extracted from git config.

### Changes Required:

#### 1. Init Command
**File**: `cmd/init.go`
**Changes**: Create init command implementation

```go
package cmd

// initCmd represents the init command
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize claude-md storage directory",
    Long:  "Creates the ~/.claude/claude-md/<user> directory structure for storing CLAUDE.md files",
    RunE:  runInit,
}

func init()

func runInit(cmd *cobra.Command, args []string) error
```

**Function Responsibilities:**
- Call git operations to get user email
- Extract user from email
- Create storage directory structure `~/.claude/claude-md/<user>/`
- Report user name used
- If directory exists, report it already exists and skip creation
- Return appropriate error codes

#### 2. Output Formatting
**File**: `internal/output/formatter.go` (new)
**Changes**: Create simple output formatting utilities

```go
package output

// PrintInfo prints informational message
func PrintInfo(format string, args ...interface{})

// PrintSuccess prints success message
func PrintSuccess(format string, args ...interface{})

// PrintError prints error message
func PrintError(format string, args ...interface{})
```

**Function Responsibilities:**
- Simple wrappers around `fmt.Printf` or `fmt.Fprintf(os.Stderr, ...)`
- Consistent formatting across commands

### Testing Requirements:

```go
// cmd/init_test.go
func TestInitCommand(t *testing.T)
```

**Test Objectives:**
- Verify directory creation with correct path
- Verify output message includes user name
- Verify idempotent behavior (running twice doesn't error)
- Test in non-git directory (should error)

**Context for Implementation:**
- Use `os.MkdirAll` for directory creation
- Use `os.Stat` to check if directory exists
- Print to stdout, errors to stderr
- Return non-zero exit code on errors

### Validation
- [ ] Run: `go build -o claude-md .`
- [ ] Run: `go test ./...`
- [ ] Run: `./claude-md init` in a git repo
- [ ] Verify: Directory created at `~/.claude/claude-md/<user>/`
- [ ] Verify: Running twice reports directory exists

## Phase 3: Save Command

### Overview
Implement the `save` command that finds all CLAUDE.md files in the repository, validates paths, converts them to symlinks pointing to storage, and reports results.

### Changes Required:

#### 1. File Discovery
**File**: `internal/files/discovery.go` (new)
**Changes**: Implement CLAUDE.md file discovery

```go
package files

// ClaudeFile represents a found CLAUDE.md file
type ClaudeFile struct {
    AbsolutePath     string  // Full path to file
    RepoRelativePath string  // Path relative to repo root
    IsSymlink        bool    // Whether it's already a symlink
}

// FindClaudeFiles finds all CLAUDE.md files in the repository
func FindClaudeFiles(repoRoot string) ([]ClaudeFile, error)

// IsSymlink checks if a path is a symbolic link
func IsSymlink(path string) (bool, error)
```

**Function Responsibilities:**
- `FindClaudeFiles`: Use `filepath.Walk` from repo root, find all files matching `CLAUDE.md` (case-insensitive using `strings.EqualFold`)
- Skip `.git` directory during walk
- Calculate relative path from repo root using `filepath.Rel`
- Check each file with `os.Lstat` to determine if it's a symlink
- Return list of discovered files
- `IsSymlink`: Use `os.Lstat` and check `fi.Mode() & os.ModeSymlink != 0`

#### 2. Save Operation
**File**: `internal/operations/save.go` (new)
**Changes**: Implement save logic

```go
package operations

// SaveResult represents the result of saving a file
type SaveResult struct {
    RepoRelativePath string
    StoragePath      string  // Path to storage file (for warnings/info)
    Success          bool
    Skipped          bool
    SkipReason       string  // "already symlink", "invalid path", "storage file exists", etc.
    Warning          string  // Warning message for user
    Error            error
}

// SaveOptions contains options for save operation
type SaveOptions struct {
    RepoRoot     string
    PathConverter *storage.PathConverter
}

// SaveFiles converts CLAUDE.md files to symlinks
func SaveFiles(files []files.ClaudeFile, opts SaveOptions) []SaveResult
```

**Function Responsibilities:**
- For each file, validate path (no `~` in components), skip with error if invalid
- Skip if already a symlink (with skip reason)
- Check if storage file already exists at target location
- If storage file exists: skip with warning showing both repo path and storage path, explain conflict
- If storage file doesn't exist: create storage directory if needed, copy file content to storage location, remove original file, create symlink with absolute path to storage
- Use absolute path when creating symlink: `os.Symlink(absoluteStoragePath, repoFilePath)`
- Collect and return results for all files
- Handle errors gracefully (continue processing other files, collect errors in results)

#### 3. Save Command
**File**: `cmd/save.go`
**Changes**: Create save command

```go
package cmd

var saveCmd = &cobra.Command{
    Use:   "save",
    Short: "Save CLAUDE.md files to storage",
    Long:  "Finds all CLAUDE.md files in the repository and converts them to symlinks pointing to centralized storage",
    RunE:  runSave,
}

func init()

func runSave(cmd *cobra.Command, args []string) error
```

**Function Responsibilities:**
- Detect git repository
- Find all CLAUDE.md files
- Execute save operation
- Report results: number of files saved, skipped, errors
- List each file that was converted to symlink
- Return error if no files found

### Testing Requirements:

```go
// internal/files/discovery_test.go
func TestFindClaudeFiles(t *testing.T)
func TestIsSymlink(t *testing.T)

// internal/operations/save_test.go
func TestSaveFiles(t *testing.T)

// cmd/save_test.go
func TestSaveCommand(t *testing.T)
```

**Test Objectives:**
- `TestFindClaudeFiles`: Verify finding files in nested directories, detecting symlinks
- `TestIsSymlink`: Test symlink detection vs regular files
- `TestSaveFiles`: Test file conversion, path validation, skipping symlinks, error handling
- `TestSaveCommand`: Integration test with temporary git repo and files

**Context for Implementation:**
- Use `filepath.Walk` from standard library, skip `.git` directories
- Use `strings.EqualFold(filepath.Base(path), "CLAUDE.md")` for case-insensitive matching
- Use `os.Lstat` to avoid following symlinks during discovery
- Use `filepath.Abs` to get absolute path to storage file before creating symlink
- Use `os.ReadFile`/`os.WriteFile` for file content copying
- Use `os.Remove` to delete original file before creating symlink
- Use `os.Symlink(absoluteStoragePath, repoFilePath)` to create symlink
- Check storage file existence with `os.Stat` before copying
- Create temp directories in tests for isolated testing

### Validation
- [ ] Run: `go build -o claude-md .`
- [ ] Run: `go test ./...`
- [ ] Create test git repo with multiple CLAUDE.md files at different levels
- [ ] Run: `./claude-md save`
- [ ] Verify: Files converted to symlinks
- [ ] Verify: Stored files exist in `~/.claude/claude-md/<user>/<repo>.git/`
- [ ] Verify: Running save again skips already-symlinked files

## Phase 4: Restore Command

### Overview
Implement the `restore` command that finds all stored CLAUDE.md files for the current repository and creates symlinks in the appropriate locations.

### Changes Required:

#### 1. Storage Discovery
**File**: `internal/files/storage.go` (new)
**Changes**: Implement storage file discovery

```go
package files

// StoredFile represents a file in storage
type StoredFile struct {
    StorageFilename  string  // Filename in storage (e.g., "source~go~api~CLAUDE.md")
    RepoRelativePath string  // Path to restore in repo (e.g., "source/go/api/CLAUDE.md")
    StoragePath      string  // Full path to stored file
}

// FindStoredFiles finds all stored CLAUDE.md files for a repository
func FindStoredFiles(repoStorageDir string, converter *storage.PathConverter) ([]StoredFile, error)
```

**Function Responsibilities:**
- Read directory listing from repo storage directory
- Convert each filename back to repository path
- Return list of files to restore
- Handle empty directory (no stored files)

#### 2. Restore Operation
**File**: `internal/operations/restore.go` (new)
**Changes**: Implement restore logic

```go
package operations

// RestoreResult represents the result of restoring a file
type RestoreResult struct {
    RepoRelativePath string
    Success          bool
    Skipped          bool
    SkipReason       string  // "already exists", "parent dir missing", etc.
    Warning          string  // For parent directory missing case
    StoragePath      string  // Path to stored file (for warnings)
    Error            error
}

// RestoreOptions contains options for restore operation
type RestoreOptions struct {
    RepoRoot string
}

// RestoreFiles creates symlinks for stored files
func RestoreFiles(files []files.StoredFile, opts RestoreOptions) []RestoreResult
```

**Function Responsibilities:**
- For each stored file, construct target path in repo using `ConvertToRepoPath`
- Check if parent directory exists using `filepath.Dir` and `os.Stat`
- If parent directory missing: skip with warning showing both repo path and storage path, explain parent directory doesn't exist
- Check if file already exists at target location (use `os.Lstat` to not follow symlinks)
- If regular file exists: skip with warning showing both paths, explain file already exists
- If symlink exists: check target with `os.Readlink`, skip if points to correct storage path
- If symlink points to wrong location: skip with warning showing both paths and current target
- If nothing exists: create symlink using absolute path `os.Symlink(absoluteStoragePath, repoFilePath)`
- Collect and return results for all files
- Handle errors gracefully (continue processing other files)

#### 3. Restore Command
**File**: `cmd/restore.go`
**Changes**: Create restore command

```go
package cmd

var restoreCmd = &cobra.Command{
    Use:   "restore",
    Short: "Restore CLAUDE.md files from storage",
    Long:  "Creates symlinks for all stored CLAUDE.md files in the repository",
    RunE:  runRestore,
}

func init()

func runRestore(cmd *cobra.Command, args []string) error
```

**Function Responsibilities:**
- Detect git repository
- Find all stored files for this repo
- Execute restore operation
- Report results: number of files restored, skipped, warnings
- Display warnings for missing parent directories with storage paths
- Return error if no stored files found

### Testing Requirements:

```go
// internal/files/storage_test.go
func TestFindStoredFiles(t *testing.T)

// internal/operations/restore_test.go
func TestRestoreFiles(t *testing.T)

// cmd/restore_test.go
func TestRestoreCommand(t *testing.T)
```

**Test Objectives:**
- `TestFindStoredFiles`: Verify finding stored files, path conversion
- `TestRestoreFiles`: Test symlink creation, skipping existing files, parent directory checks
- `TestRestoreCommand`: Integration test with temporary storage and repo

**Context for Implementation:**
- Use `os.ReadDir` to list files in storage directory
- Filter to only files matching case-insensitive `CLAUDE.md` using `strings.EqualFold`
- Use `filepath.Dir` to get parent directory of target path
- Use `os.Stat` to check directory existence (returns error if doesn't exist)
- Use `os.Lstat` to check file existence without following symlinks
- Use `filepath.Abs` to get absolute path to storage file
- Use `os.Symlink(absoluteStoragePath, repoFilePath)` to create symlink
- Use `os.Readlink` to read existing symlink target for comparison

### Validation
- [ ] Run: `go build -o claude-md .`
- [ ] Run: `go test ./...`
- [ ] Clear test repo: `rm -rf test-repo/source/go/api/CLAUDE.md` (if it's a symlink)
- [ ] Run: `./claude-md restore`
- [ ] Verify: Symlinks created at correct locations
- [ ] Verify: Running restore again skips already-correct symlinks

## Phase 5: Clear Command

### Overview
Implement the `clear` command that finds and removes all CLAUDE.md symlinks from the repository (but does not delete stored files).

### Changes Required:

#### 1. Clear Operation
**File**: `internal/operations/clear.go` (new)
**Changes**: Implement clear logic

```go
package operations

// ClearResult represents the result of clearing symlinks
type ClearResult struct {
    RepoRelativePath string
    Success          bool
    Error            error
}

// ClearOptions contains options for clear operation
type ClearOptions struct {
    RepoRoot string
}

// ClearSymlinks removes all CLAUDE.md symlinks from repository
func ClearSymlinks(opts ClearOptions) []ClearResult
```

**Function Responsibilities:**
- Find all CLAUDE.md files in repository (case-insensitive, skip `.git` directory)
- Filter to only symlinks using `IsSymlink`
- For each symlink: remove using `os.Remove`
- Collect results showing which symlinks were removed
- Handle errors gracefully (continue processing other symlinks if one fails)
- Note: does NOT delete files in storage, only removes symlinks from repo

#### 2. Clear Command
**File**: `cmd/clear.go`
**Changes**: Create clear command

```go
package cmd

var clearCmd = &cobra.Command{
    Use:   "clear",
    Short: "Clear CLAUDE.md symlinks from repository",
    Long:  "Removes all CLAUDE.md symbolic links from the repository (stored files remain in storage)",
    RunE:  runClear,
}

func init()

func runClear(cmd *cobra.Command, args []string) error
```

**Function Responsibilities:**
- Detect git repository
- Execute clear operation
- Report results: number of symlinks removed, errors
- List each symlink that was removed
- Note that stored files remain in storage

### Testing Requirements:

```go
// internal/operations/clear_test.go
func TestClearSymlinks(t *testing.T)

// cmd/clear_test.go
func TestClearCommand(t *testing.T)
```

**Test Objectives:**
- `TestClearSymlinks`: Verify finding and removing only symlinks, preserving regular files
- `TestClearCommand`: Integration test verifying symlinks removed but stored files intact

**Context for Implementation:**
- Reuse `files.FindClaudeFiles` from save command
- Filter to only symlinks using `IsSymlink`
- Use `os.Remove` to delete symlinks
- Verify stored files still exist after clearing

### Validation
- [ ] Run: `go build -o claude-md .`
- [ ] Run: `go test ./...`
- [ ] Run: `./claude-md restore` (create some symlinks)
- [ ] Run: `./claude-md clear`
- [ ] Verify: Symlinks removed from repository
- [ ] Verify: Stored files still exist in `~/.claude/claude-md/<user>/<repo>.git/`
- [ ] Run: `./claude-md restore` again
- [ ] Verify: Symlinks can be restored again

## Phase 6: Polish & Documentation

### Overview
Add final polish including better error messages, usage examples, and README documentation.

### Changes Required:

#### 1. Error Messages & Help Text
**Files**: All `cmd/*.go` files
**Changes**: Improve command descriptions and examples

**Function Responsibilities:**
- Add usage examples to each command's `Long` description
- Improve error messages to be actionable
- Add `Example` field to commands showing typical usage

#### 2. README Documentation
**File**: `README.md` (new)
**Changes**: Create comprehensive README

**Content to include:**
- Project overview and purpose
- Installation instructions
- Usage examples for each command
- Storage structure explanation
- Limitations (no `~` in directory names)
- Requirements (must be in git repo with origin)

#### 3. Integration Testing
**File**: `test/integration_test.go` (new)
**Changes**: Create end-to-end integration tests

```go
package test

func TestFullWorkflow(t *testing.T)
```

**Test Objectives:**
- Test complete workflow: init → save → clear → restore
- Verify symlinks point to correct locations
- Verify file contents preserved through save/restore cycle
- Test error cases: no git repo, no origin, invalid paths

**Context for Implementation:**
- Create temporary git repository for testing
- Initialize git config in test repo
- Add origin remote for testing
- Clean up after tests

### Validation
- [ ] Run: `go build -o claude-md .`
- [ ] Run: `go test ./...`
- [ ] Run: `./claude-md --help`
- [ ] Verify: All commands documented with examples
- [ ] Test full workflow manually in real git repository
- [ ] Verify: README is clear and complete

## Cross-Phase Considerations

### Error Handling Strategy
- All functions return errors following Go conventions
- Commands return non-zero exit codes on errors
- Error messages include context (file paths, operation attempted)
- Distinguish between fatal errors (stop execution) and warnings (continue)

### Testing Strategy
- Follow TDD approach: write tests before implementation
- Use table-driven tests for validation and parsing logic
- Create temporary directories/files for file operation tests
- Use test fixtures for git repository testing
- Achieve high test coverage (aim for >80%)

### Git Operations
- Always validate repository has origin remote (error if not)
- Use absolute paths when creating symlinks
- Handle worktrees correctly by using git commands to find root

### File Operations
- Always use `os.Lstat` to avoid following symlinks during discovery
- Create directories with `0755` permissions
- Create files with `0644` permissions
- Handle permission errors gracefully

### Path Operations
- Always use `filepath` package functions for cross-platform compatibility
- Clean paths using `filepath.Clean`
- Convert to absolute paths when needed
- Use `filepath.Rel` for relative path calculation

## Dependencies Summary

Required third-party packages:
- `github.com/spf13/cobra` - CLI framework

Standard library packages heavily used:
- `os` - File operations, symlinks, reading directories
- `os/exec` - Running git commands
- `path/filepath` - Path manipulation, walking directories
- `fmt` - Formatting output
- `strings` - String parsing, case-insensitive matching
