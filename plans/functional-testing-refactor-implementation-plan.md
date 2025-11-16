# Functional Testing Refactor Implementation Plan

## Overview

Refactor the claude-md CLI application to follow functional testing principles as defined in the functional-testing skill. The application currently violates functional testing principles by directly calling Cobra's Execute() method, hardcoding stdout/stderr, and requiring binary compilation for testing. This refactor will make the CLI testable by adding a Run() function with dependency injection while maintaining Cobra for command routing.

## Current State Analysis

### Violations of Functional Testing Principles

1. **No Run() function**: `main.go:10` directly calls `cmd.Execute()` - Cobra-specific, untestable
2. **No dependency injection**: Commands use `internal/output/formatter.go` with hardcoded `os.Stdout`/`os.Stderr` (lines 10, 15, 20)
3. **Tests build binaries**: `test/integration_test.go:55` compiles binary and uses `exec.Command()` instead of calling functions
4. **Can't capture output**: No way to inject test writers for stdout/stderr verification

### What Works Well

1. **Clean operations layer**: `internal/operations/save.go`, `internal/operations/restore.go`, `internal/operations/clear.go` return result structs with no I/O
2. **Internal packages use correct test patterns**: All use `XXX_test` packages (`internal/git/repo_test.go:1`, `internal/storage/paths_test.go:1`, `internal/operations/clear_test.go:1`)
3. **Consistent command structure**: All 4 commands follow same pattern (init, save, restore, clear)

### Key Discoveries

- **Command pattern** (`cmd/save.go:33-105`): All commands follow identical structure:
  1. Get git repo info
  2. Extract user/repo name
  3. Create PathConverter
  4. Call operations function
  5. Print results
- **Operations return results**: All operations functions already return structured results (e.g., `operations.SaveResult`, `operations.RestoreResult`, `operations.ClearResult`)
- **Output functions**: `internal/output/formatter.go` has 3 functions: `PrintInfo()`, `PrintSuccess()`, `PrintError()`

## Desired End State

A CLI application that follows functional testing principles where:

1. **main() calls Run()**: `main.go` calls `cmd.Run(os.Argv[1:], cmd.RunOptions{...})`
2. **Run() accepts dependencies**: `cmd.Run(args []string, opts RunOptions) int` where RunOptions includes `Stdout` and `Stderr`
3. **Tests call Run() directly**: Tests in `cmd_test` package call `cmd.Run()` with test args and buffers
4. **Output is injectable**: Commands write to injected `io.Writer` instead of global stdout/stderr
5. **No binary compilation in tests**: Tests import and call cmd package functions

### Verification

After implementation, verify by:
- Running `go test ./...` - all tests pass
- Running `go test -v ./cmd` - shows unit tests for all 4 commands
- Running `go test -v ./test` - integration test uses Run() not exec.Command
- Running `go build .` - binary still works identically
- Running each command manually - behavior unchanged

## What We're NOT Doing

- NOT removing Cobra (keeping it for command routing)
- NOT adding context.Context (this is not a server CLI)
- NOT preserving backward compatibility (project not in use)
- NOT changing the behavior of any commands
- NOT changing the operations layer (already well-structured)
- NOT testing internal/private functions (already correct)

## Implementation Approach

The refactor follows this strategy:

1. **Infrastructure first**: Create Run() function and RunOptions struct that wraps Cobra
2. **Make output injectable**: Refactor output package to accept io.Writer
3. **Update commands**: Modify all 4 commands to use injected output
4. **Add tests**: Create unit tests for each command
5. **Refactor integration test**: Update to call Run() instead of exec.Command

This approach ensures each phase builds on the previous one and maintains a working state throughout.

## Phase 1: Create Run() Function and RunOptions Infrastructure

### Overview
Add the core Run() function that tests will call, along with RunOptions for dependency injection. This wraps Cobra's Execute() to make it testable.

### Changes Required

#### 1. cmd/run.go (New File)
**File**: `cmd/run.go`
**Changes**: Create new file with Run() function

```go
// Run executes the CLI with given arguments and options
func Run(args []string, opts RunOptions) int

// RunOptions provides injectable dependencies for testing
type RunOptions struct {
	Stdout io.Writer
	Stderr io.Writer
}
```

**Function Responsibilities:**
- Set default writers to os.Stdout/os.Stderr if nil
- Create Output instance and store in package variable for commands to access
- Configure Cobra's output streams (rootCmd.SetOut/SetErr) for help text
- Set rootCmd args with `rootCmd.SetArgs(args)`
- Call `rootCmd.Execute()`
- Return exit code: 0 on success, 1 on error

**Complete Implementation Example:**
```go
// Package-level variable for commands to access output
var currentOutput *output.Output

func Run(args []string, opts RunOptions) int {
	// Set defaults
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	// Make output available to commands
	currentOutput = output.NewOutput(opts.Stdout, opts.Stderr)

	// Configure Cobra's output streams (for help text, errors)
	rootCmd.SetOut(opts.Stdout)
	rootCmd.SetErr(opts.Stderr)

	// Let Cobra parse args and route to commands
	rootCmd.SetArgs(args)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}
```

**Context for implementation:**
- Follow pattern from `functional-testing.md:16-33`
- Commands access output via package-level `currentOutput` variable
- main.go:10 currently calls `cmd.Execute()` which will be replaced
- Cobra's SetOut/SetErr ensures help text goes to injected writers

#### 2. main.go
**File**: `main.go`
**Changes**: Replace cmd.Execute() with cmd.Run()

```go
func main() {
	os.Exit(cmd.Run(os.Argv[1:], cmd.RunOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}))
}
```

**Context for implementation:**
- Current code at `main.go:9-13`
- Pattern from `functional-testing.md:27-33`

### Testing Requirements

```go
// New test file: cmd/run_test.go
func TestRunWithInvalidCommand(t *testing.T)
func TestRunWithHelpFlag(t *testing.T)
func TestRunWithNoArgs(t *testing.T)
```

**Test Objectives:**
- Verify Run() returns exit code 1 for invalid commands
- Verify Run() returns exit code 0 for help flag
- Verify output goes to provided writers not global stdout/stderr
- Verify default writers work when nil

**Context for implementation:**
- Tests in `package cmd_test` to enforce public interface testing
- Follow test pattern from `functional-testing.md:76-98`
- Use `bytes.Buffer` for capturing output
- These tests don't need working directory changes (only test Run() infrastructure)

**Important Note on Working Directory:**
The Run() function does NOT change the working directory. Commands call `git.FindRepository()` which searches from the current working directory. Tests that need to simulate being in a git repository (Phase 4) must handle working directory explicitly using `os.Chdir()`. Pattern:

```go
func TestCommandInGitRepo(t *testing.T) {
	repoDir := t.TempDir()
	// ... set up git repo in repoDir ...

	// Change to repo directory before calling Run()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(repoDir)

	var stdout bytes.Buffer
	exitCode := cmd.Run([]string{"init"}, cmd.RunOptions{
		Stdout: &stdout,
	})
	// ... assertions ...
}
```

**CRITICAL**: Tests using `os.Chdir()` MUST NOT use `t.Parallel()` as this causes race conditions with working directory changes.

### Validation
- [ ] Run: `go test ./cmd -v -run TestRun`
- [ ] Verify: Tests pass and use Run() function
- [ ] Run: `go build .`
- [ ] Verify: Binary builds and runs

## Phase 2: Refactor Output Package for Dependency Injection

### Overview
Modify the output package to accept io.Writer parameters instead of using global os.Stdout/os.Stderr. Create an Output type that holds the writers.

### Changes Required

#### 1. internal/output/formatter.go
**File**: `internal/output/formatter.go`
**Changes**: Replace global functions with Output type

```go
// Output handles formatted output to configurable writers
type Output struct {
	Stdout io.Writer
	Stderr io.Writer
}

// NewOutput creates an Output with default writers
func NewOutput(stdout, stderr io.Writer) *Output

// PrintInfo prints informational message to stdout
func (o *Output) PrintInfo(format string, args ...interface{})

// PrintSuccess prints success message to stdout
func (o *Output) PrintSuccess(format string, args ...interface{})

// PrintError prints error message to stderr
func (o *Output) PrintError(format string, args ...interface{})
```

**Function Responsibilities:**
- NewOutput sets defaults to os.Stdout/os.Stderr if nil
- PrintInfo writes to o.Stdout (current: `fmt.Printf`, `formatter.go:10`)
- PrintSuccess writes to o.Stdout (current: `fmt.Printf`, `formatter.go:15`)
- PrintError writes to o.Stderr (current: `fmt.Fprintf(os.Stderr, ...)`, `formatter.go:20`)
- All functions add newline automatically (preserve current behavior)

**Context for implementation:**
- Current implementation at `internal/output/formatter.go:8-21`
- Use `fmt.Fprintf(o.Stdout, format+"\n", args...)`
- Keep the automatic newline behavior

#### 2. cmd/run.go
**File**: `cmd/run.go`
**Changes**: Create Output instance and make it accessible to commands

```go
// Package-level variable for commands to access
var currentOutput *output.Output

// In Run() function, initialize currentOutput
func Run(args []string, opts RunOptions) int {
	currentOutput = output.NewOutput(opts.Stdout, opts.Stderr)
	// ... rest of Run() implementation
}
```

**Context for implementation:**
- Commands will access via `currentOutput` package variable
- Alternative: Use cobra command context, but package variable is simpler

### Testing Requirements

```go
// New test file: internal/output/formatter_test.go
func TestPrintInfo(t *testing.T)
func TestPrintSuccess(t *testing.T)
func TestPrintError(t *testing.T)
func TestNewOutputDefaults(t *testing.T)
```

**Test Objectives:**
- Verify PrintInfo writes to stdout with newline
- Verify PrintSuccess writes to stdout with newline
- Verify PrintError writes to stderr with newline
- Verify formatting with args works correctly
- Verify nil writers default to os.Stdout/os.Stderr

**Context for implementation:**
- Tests in `package output_test`
- Use `bytes.Buffer` to capture output
- Verify exact output strings

### Validation
- [ ] Run: `go test ./internal/output -v`
- [ ] Verify: All output tests pass
- [ ] Run: `go build .`
- [ ] Verify: No compilation errors

## Phase 3: Update All Commands to Use Injected Output

### Overview
Refactor all 4 commands (init, save, restore, clear) to use the injected Output instance instead of the global output package functions.

### Changes Required

#### 1. cmd/init.go
**File**: `cmd/init.go`
**Changes**: Replace all `output.PrintX()` calls with `currentOutput.PrintX()`

**Function Responsibilities:**
- Line 34: `currentOutput.PrintError("Error: %v", err)` (was `output.PrintError`)
- Line 40: `currentOutput.PrintError("Error: %v", err)`
- Line 46: `currentOutput.PrintError("Error: %v", err)`
- Line 52: `currentOutput.PrintError("Error: %v", err)`
- Line 58: `currentOutput.PrintError("Error: %v", err)`
- Line 64: `currentOutput.PrintError("Error: %v", err)`
- Line 71: `currentOutput.PrintInfo("Storage directory already exists: %s", storageDir)`
- Line 72: `currentOutput.PrintInfo("User: %s", user)`
- Line 77: `currentOutput.PrintError("Error: %v", err)`
- Line 81: `currentOutput.PrintSuccess("Created storage directory: %s", storageDir)`
- Line 82: `currentOutput.PrintInfo("User: %s", user)`

**Context for implementation:**
- Current code at `cmd/init.go:31-85`
- Pattern: Replace `output.` with `currentOutput.`
- No logic changes, only output calls

#### 2. cmd/save.go
**File**: `cmd/save.go`
**Changes**: Replace all `output.PrintX()` calls with `currentOutput.PrintX()`

**Function Responsibilities:**
- Line 36: `currentOutput.PrintError("Error: %v", err)`
- Line 42: `currentOutput.PrintError("Error: %v", err)`
- Line 48: `currentOutput.PrintError("Error: %v", err)`
- Line 54: `currentOutput.PrintError("Error: %v", err)`
- Line 60: `currentOutput.PrintError("Error: %v", err)`
- Line 66: `currentOutput.PrintError("Error: %v", err)`
- Line 72: `currentOutput.PrintError("Error finding CLAUDE.md files: %v", err)`
- Line 78: `currentOutput.PrintInfo("No CLAUDE.md files found in repository")`
- Line 90: `currentOutput.PrintSuccess("Saved: %s", result.RepoRelativePath)`
- Line 94: `currentOutput.PrintInfo("Warning: %s", result.Warning)`
- Line 98: `currentOutput.PrintError("Error: %s", result.Warning)`
- Line 102: `currentOutput.PrintInfo("\nSummary: %d saved, %d skipped, %d errors", saved, skipped, errors)`

**Context for implementation:**
- Current code at `cmd/save.go:33-105`
- Same pattern as init command

#### 3. cmd/restore.go
**File**: `cmd/restore.go`
**Changes**: Replace all `output.PrintX()` calls with `currentOutput.PrintX()`

**Function Responsibilities:**
- Line 35: `currentOutput.PrintError("Error: %v", err)`
- Line 41: `currentOutput.PrintError("Error: %v", err)`
- Line 47: `currentOutput.PrintError("Error: %v", err)`
- Line 53: `currentOutput.PrintError("Error: %v", err)`
- Line 59: `currentOutput.PrintError("Error: %v", err)`
- Line 65: `currentOutput.PrintError("Error: %v", err)`
- Line 72: `currentOutput.PrintError("Error finding stored files: %v", err)`
- Line 77: `currentOutput.PrintInfo("No stored CLAUDE.md files found for this repository")`
- Line 89: `currentOutput.PrintSuccess("Restored: %s", result.RepoRelativePath)`
- Line 94: `currentOutput.PrintInfo("Warning: %s", result.Warning)`
- Line 99: `currentOutput.PrintInfo("\nSummary: %d restored, %d skipped (%d warnings)", restored, skipped, warnings)`

**Context for implementation:**
- Current code at `cmd/restore.go:32-102`
- Same pattern as previous commands

#### 4. cmd/clear.go
**File**: `cmd/clear.go`
**Changes**: Replace all `output.PrintX()` calls with `currentOutput.PrintX()`

**Function Responsibilities:**
- Line 34: `currentOutput.PrintError("Error: %v", err)`
- Line 40: `currentOutput.PrintError("Error: %v", err)`
- Line 46: `currentOutput.PrintError("Error: %v", err)`
- Line 52: `currentOutput.PrintError("Error: %v", err)`
- Line 58: `currentOutput.PrintError("Error: %v", err)`
- Line 64: `currentOutput.PrintError("Error: %v", err)`
- Line 77: `currentOutput.PrintSuccess("Removed: %s", result.RepoRelativePath)`
- Line 80: `currentOutput.PrintInfo("Skipped %s: %s", result.RepoRelativePath, result.SkipReason)`
- Line 83: `currentOutput.PrintError("Error removing %s: %v", result.RepoRelativePath, result.Error)`
- Line 88: `currentOutput.PrintInfo("No CLAUDE.md symlinks found in repository")`
- Line 90: `currentOutput.PrintInfo("\nSummary: %d removed, %d skipped, %d errors", removed, skipped, errors)`
- Line 91: `currentOutput.PrintInfo("Note: Stored files remain in storage")`

**Context for implementation:**
- Current code at `cmd/clear.go:31-95`
- Same pattern as previous commands

### Testing Requirements

**NOTE**: This phase only refactors existing commands. No new functionality is added, so no new test signatures are needed. The existing integration test at `test/integration_test.go` will verify commands still work. Phase 4 will add new unit tests.

### Validation
- [ ] Run: `go build .`
- [ ] Verify: Binary compiles successfully
- [ ] Run: `./claude-md --help`
- [ ] Verify: Help text displays correctly
- [ ] Run: `go test ./test -v`
- [ ] Verify: Existing integration test still passes (tests real file operations)

## Phase 4: Add Unit Tests for Each Command

### Overview
Create unit tests for all 4 commands that call Run() directly with test arguments and buffers. These tests verify command output and behavior without file system operations (use temporary directories).

### Changes Required

#### 1. cmd/testutil_test.go (New File - Test Helper)
**File**: `cmd/testutil_test.go`
**Changes**: Create shared test helper for git repository setup

```go
package cmd

// setupTestGitRepo creates a temporary git repository configured for testing
// Returns the absolute path to the repository directory
func setupTestGitRepo(t *testing.T) string
```

**Function Responsibilities:**
- Use `t.TempDir()` to create temporary directory
- Create subdirectory for repository
- Initialize git repository with `exec.Command("git", "init")`
- Configure git user.email with test value
- Configure git user.name with test value
- Add origin remote with test URL
- Return absolute path to repository directory
- Use `t.Helper()` to mark as test helper

**Complete Implementation:**
```go
package cmd

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupTestGitRepo creates a temporary git repository configured for testing
func setupTestGitRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "test-repo")
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	// Initialize git
	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	// Add origin remote
	cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	return repoDir
}
```

**Context for implementation:**
- This helper avoids duplicating 30+ lines of setup code in each test file
- Follow pattern from `test/integration_test.go:13-47`
- All command tests will use this helper
- Defined in package `cmd` (not `cmd_test`) so it's only compiled during tests

#### 2. cmd/init_test.go (New File)
**File**: `cmd/init_test.go`
**Changes**: Create new test file

```go
// Test signatures for new tests
func TestInitCommand(t *testing.T)
func TestInitCommandAlreadyExists(t *testing.T)
func TestInitCommandErrorCases(t *testing.T)
```

**Test Objectives:**
- Verify init command creates storage directory
- Verify init command prints success message with user and path
- Verify init command handles already-existing directory
- Verify init command handles errors (no git repo, no remote, etc.)
- Verify output goes to stdout (success messages) and stderr (errors)

**Context for implementation:**
- Use `setupTestGitRepo(t)` helper to create test repository
- Change to repo directory with `os.Chdir(repoDir)` before calling Run()
- Remember to defer `os.Chdir(oldDir)` to restore working directory
- DO NOT use `t.Parallel()` - working directory changes are not safe with parallel tests
- Call `cmd.Run([]string{"init"}, cmd.RunOptions{Stdout: &stdout, Stderr: &stderr})`
- Verify exit code 0 for success
- Verify stdout contains "Created storage directory" and "User: test"
- Verify stderr is empty on success
- Follow pattern from Phase 1 working directory example
- Follow pattern from `functional-testing.md:76-98` for Run() calling

#### 2. cmd/save_test.go (New File)
**File**: `cmd/save_test.go`
**Changes**: Create new test file

```go
// Test signatures for new tests
func TestSaveCommand(t *testing.T)
func TestSaveCommandNoFiles(t *testing.T)
func TestSaveCommandAlreadySymlink(t *testing.T)
func TestSaveCommandErrorCases(t *testing.T)
```

**Test Objectives:**
- Verify save command converts CLAUDE.md files to symlinks
- Verify save command prints success for each saved file
- Verify save command handles no files found
- Verify save command skips already-converted symlinks
- Verify save command handles errors
- Verify summary line is printed

**Context for implementation:**
- Use `setupTestGitRepo(t)` helper to create test repository
- Change to repo directory with `os.Chdir(repoDir)` and defer restore
- Create test CLAUDE.md files with `os.WriteFile(filepath.Join(repoDir, "CLAUDE.md"), ...)`
- Run init command first: `cmd.Run([]string{"init"}, ...)`
- Call `cmd.Run([]string{"save"}, cmd.RunOptions{Stdout: &stdout, Stderr: &stderr})`
- Verify files are symlinks with `os.Lstat()` and `info.Mode()&os.ModeSymlink`
- Verify stdout contains "Saved:" messages and summary
- DO NOT use `t.Parallel()` - working directory changes are not safe

#### 3. cmd/restore_test.go (New File)
**File**: `cmd/restore_test.go`
**Changes**: Create new test file

```go
// Test signatures for new tests
func TestRestoreCommand(t *testing.T)
func TestRestoreCommandNoStoredFiles(t *testing.T)
func TestRestoreCommandAlreadyExists(t *testing.T)
func TestRestoreCommandErrorCases(t *testing.T)
```

**Test Objectives:**
- Verify restore command creates symlinks from storage
- Verify restore command prints success for each restored file
- Verify restore command handles no stored files
- Verify restore command skips already-existing files
- Verify restore command handles errors
- Verify summary line is printed

**Context for implementation:**
- Use `setupTestGitRepo(t)` helper to create test repository
- Change to repo directory with `os.Chdir(repoDir)` and defer restore
- Create CLAUDE.md files, run init, then run save to populate storage
- Delete symlinks with `os.Remove()` to simulate `git clean`
- Call `cmd.Run([]string{"restore"}, cmd.RunOptions{Stdout: &stdout, Stderr: &stderr})`
- Verify symlinks are recreated with `os.Lstat()`
- Verify stdout contains "Restored:" messages and summary
- DO NOT use `t.Parallel()` - working directory changes are not safe

#### 4. cmd/clear_test.go (New File)
**File**: `cmd/clear_test.go`
**Changes**: Create new test file

```go
// Test signatures for new tests
func TestClearCommand(t *testing.T)
func TestClearCommandNoSymlinks(t *testing.T)
func TestClearCommandSkipsRegularFiles(t *testing.T)
func TestClearCommandErrorCases(t *testing.T)
```

**Test Objectives:**
- Verify clear command removes symlinks
- Verify clear command prints success for each removed file
- Verify clear command handles no symlinks found
- Verify clear command skips regular files
- Verify clear command handles errors
- Verify storage files remain after clear
- Verify summary line is printed

**Context for implementation:**
- Use `setupTestGitRepo(t)` helper to create test repository
- Change to repo directory with `os.Chdir(repoDir)` and defer restore
- Create CLAUDE.md files, run init, then run save to create symlinks
- Call `cmd.Run([]string{"clear"}, cmd.RunOptions{Stdout: &stdout, Stderr: &stderr})`
- Verify symlinks are removed with `os.IsNotExist(err)` from `os.Lstat()`
- Verify storage files still exist in ~/.claude/claude-md/test/repo.git/
- Verify stdout contains "Removed:" messages and summary
- DO NOT use `t.Parallel()` - working directory changes are not safe

### Validation
- [ ] Run: `go test ./cmd -v`
- [ ] Verify: All command tests pass
- [ ] Run: `go test ./cmd -cover`
- [ ] Verify: Good coverage of command code paths
- [ ] Run: `go build .`
- [ ] Verify: Binary still works

## Phase 5: Refactor Integration Test to Use Run()

### Overview
Update the existing integration test to call `cmd.Run()` directly instead of building a binary and using `exec.Command()`. This completes the functional testing refactor.

### Changes Required

#### 1. test/integration_test.go
**File**: `test/integration_test.go`
**Changes**: Replace binary execution with Run() calls

**Existing tests that may require updates:**
```go
func TestFullWorkflow(t *testing.T)  // Update: Use cmd.Run() instead of exec.Command
```

**Requirements for Refactoring:**
- Remove binary build code (lines 48-58)
- Remove `exec.Command(claudeMdBinary, ...)` calls
- Replace with `cmd.Run([]string{...}, cmd.RunOptions{...})`
- Keep as `package test` (already correct - no change needed)
- Import `github.com/kapetan-io/claude-md.go/cmd` package
- Use `bytes.Buffer` for stdout/stderr instead of `CombinedOutput()`
- Use `os.Chdir(repoDir)` before each Run() call (defer restore after setup)
- Verify stdout contains expected messages (e.g., "User: test")
- Verify exit codes: 0 for success
- Keep all file system verification (symlink checks, content checks)
- Preserve sub-test structure (t.Run for Init, Save, Clear, Restore)

**Context for implementation:**
- Current test at `test/integration_test.go:13-148`
- Lines 48-58: Remove binary build
- Lines 62-66: Replace `exec.Command(claudeMdBinary, "init")` with `cmd.Run([]string{"init"}, ...)`
- Lines 72-75: Replace `exec.Command(claudeMdBinary, "save")` with `cmd.Run([]string{"save"}, ...)`
- Lines 98-101: Replace `exec.Command(claudeMdBinary, "clear")` with `cmd.Run([]string{"clear"}, ...)`
- Lines 119-122: Replace `exec.Command(claudeMdBinary, "restore")` with `cmd.Run([]string{"restore"}, ...)`
- Pattern from `functional-testing.md:76-98`

**Test Objectives:**
- Verify full workflow: init → save → clear → restore
- Verify each command returns exit code 0
- Verify files are converted to symlinks (save)
- Verify symlinks are removed (clear)
- Verify symlinks are restored (restore)
- Verify file contents are preserved throughout
- Verify stdout contains success messages
- Verify stderr is empty on success

**Context for implementation:**
- Keep the git setup code (lines 13-47) - that's correct
- Keep file system verification - that's testing the actual behavior
- Only change how commands are executed (Run() vs exec.Command)
- Add `os.Chdir(repoDir)` after git setup, before first Run() call
- Add `defer os.Chdir(oldDir)` to restore working directory after test
- Each sub-test (Init, Save, Clear, Restore) calls Run() while in repoDir
- Pattern: `exitCode := cmd.Run([]string{"init"}, cmd.RunOptions{Stdout: &stdout, Stderr: &stderr})`

### Validation
- [ ] Run: `go test ./test -v`
- [ ] Verify: Integration test passes
- [ ] Verify: No binary compilation happens
- [ ] Run: `go test ./...`
- [ ] Verify: All tests in project pass

## Final Validation

After completing all phases:

- [ ] Run: `go test ./... -v`
- [ ] Verify: All tests pass (cmd unit tests, operations tests, integration test)
- [ ] Run: `go test ./... -cover`
- [ ] Verify: Good test coverage across packages
- [ ] Run: `go build .`
- [ ] Verify: Binary compiles successfully
- [ ] Run: `./claude-md init` in a test git repo
- [ ] Verify: Command works identically to before refactor
- [ ] Run: `./claude-md save` after creating CLAUDE.md
- [ ] Verify: Files converted to symlinks
- [ ] Run: `./claude-md clear`
- [ ] Verify: Symlinks removed
- [ ] Run: `./claude-md restore`
- [ ] Verify: Symlinks restored
- [ ] Verify: No `exec.Command("go", "build", ...)` in any test file
- [ ] Verify: All tests in `cmd/*_test.go` call `cmd.Run()` directly

## Success Criteria

The refactor is complete when:

1. **Run() function exists**: Tests can call `cmd.Run(args, opts)` with injectable stdout/stderr
2. **No binary compilation in tests**: Zero `exec.Command("go", "build", ...)` calls
3. **All tests pass**: Unit tests for each command + integration test
4. **Commands use injected output**: All 4 commands write to `currentOutput` not global `output` package
5. **Binary behavior unchanged**: Manual testing shows identical behavior
6. **Functional testing principles followed**: Tests call public interface with real execution, no mocked main logic
