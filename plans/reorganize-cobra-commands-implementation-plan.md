# Reorganize Cobra Commands Implementation Plan

## Overview

Reorganize the project structure to follow standard Go layout conventions by moving cobra command definitions from `cmd/` to `internal/cli/`, allowing `cmd/` to be used for the standard `cmd/claude-md/main.go` entry point pattern.

## Current State Analysis

**Current Structure:**
- `cmd/` package contains all cobra command definitions:
  - `root.go` - Root cobra command definition
  - `run.go` - Run() function (entry point with RunOptions for testing)
  - `save.go`, `restore.go`, `init.go`, `clear.go` - Individual cobra commands
  - Test files: `*_test.go` using external test package `cmd_test`
  - `testutil_test.go` - Shared test utilities

- `main.go` in project root:
  - Imports `github.com/kapetan-io/claude-md.go/cmd`
  - Calls `cmd.Run()` with args and output streams

- Build process:
  - Command: `go build -o claude-md .`
  - Builds from root `main.go`
  - Produces `claude-md` binary (currently named `claude-md.go` in directory)

### Key Discoveries:
- Each command registers itself via `init()` calling `rootCmd.AddCommand()` - see cmd/save.go:28-30
- Commands access output via package-level `currentOutput` variable - see cmd/run.go:11
- Tests use external package `cmd_test` following guidelines - see cmd/save_test.go:1
- Module name: `github.com/kapetan-io/claude-md.go`
- No build automation (Makefile/justfile) currently exists

## Desired End State

After this refactoring, the project will have:

**Directory Structure:**
```
cmd/
└── claude-md/
    └── main.go          # Entry point

internal/
├── cli/                 # Moved from cmd/
│   ├── root.go
│   ├── run.go
│   ├── save.go
│   ├── restore.go
│   ├── init.go
│   ├── clear.go
│   ├── run_test.go
│   ├── save_test.go
│   ├── restore_test.go
│   ├── init_test.go
│   ├── clear_test.go
│   └── testutil_test.go
├── files/
├── git/
├── operations/
├── output/
└── storage/
```

**Verification:**
- Build with: `go build -o claude-md ./cmd/claude-md`
- Run tests with: `go test ./...`
- All tests pass (same coverage as before)
- Binary functions identically to current version
- README reflects new build instructions

## What We're NOT Doing

- NOT changing any command functionality or behavior
- NOT modifying the cobra command logic itself
- NOT changing test implementation (only moving files and updating imports)
- NOT adding new features or commands
- NOT changing the public API of the cli package (Run() function signature stays the same)

## Implementation Approach

This is a pure refactoring task focused on moving files and updating imports. The approach is:

1. Create new directory structure
2. Move files with minimal changes (only package names and imports)
3. Update imports across the codebase
4. Update documentation and build instructions
5. Verify all tests pass and binary works

Each phase is designed to keep the project in a working state.

## Phase 1: Create New Structure and Move CLI Code

### Overview
Create `internal/cli/` directory, move all cobra command files from `cmd/` to `internal/cli/`, and create the new `cmd/claude-md/main.go` entry point.

### Changes Required:

#### 0. Capture Baseline Metrics
**Action**: Establish baseline before making changes

```bash
# Capture current test coverage
go test ./... -cover | tee /tmp/coverage-baseline.txt

# Verify current build works
go build -o claude-md .
./claude-md --help
```

**Expected Results:**
- All tests pass
- Binary builds and runs successfully
- Coverage output saved for comparison

#### 1. Create Directory Structure
**Action**: Create new directories

```bash
mkdir -p cmd/claude-md
mkdir -p internal/cli
```

#### 2. Move Command Files
**Files to move**: `cmd/*.go` → `internal/cli/*.go`

Files to move:
- `cmd/root.go` → `internal/cli/root.go`
- `cmd/run.go` → `internal/cli/run.go`
- `cmd/save.go` → `internal/cli/save.go`
- `cmd/restore.go` → `internal/cli/restore.go`
- `cmd/init.go` → `internal/cli/init.go`
- `cmd/clear.go` → `internal/cli/clear.go`

**Changes for each file**:
- Update package declaration from `package cmd` to `package cli`
- Keep all imports unchanged initially (will be updated if needed)
- Keep all code logic identical

#### 3. Move Test Files
**Files to move**: `cmd/*_test.go` → `internal/cli/*_test.go`

Files to move:
- `cmd/run_test.go` → `internal/cli/run_test.go`
- `cmd/save_test.go` → `internal/cli/save_test.go`
- `cmd/restore_test.go` → `internal/cli/restore_test.go`
- `cmd/init_test.go` → `internal/cli/init_test.go`
- `cmd/clear_test.go` → `internal/cli/clear_test.go`
- `cmd/testutil_test.go` → `internal/cli/testutil_test.go`

**Changes for external test files** (run_test.go, save_test.go, restore_test.go, init_test.go, clear_test.go):
- Update package declaration from `package cmd_test` to `package cli_test`
- Update imports: `github.com/kapetan-io/claude-md.go/cmd` → `github.com/kapetan-io/claude-md.go/internal/cli`

**Changes for testutil_test.go specifically**:
- Update package declaration from `package cmd` to `package cli` (NOTE: not `cli_test`)
- Keep all code identical
- **Important**: This file uses the internal package name (`cli`) not the external test package (`cli_test`) because it exports test utilities (`SetupTestGitRepo`) that are accessed from external test files

**Testing Requirements:**

No new tests needed - this phase only moves existing tests.

**Existing tests that will be moved**:
```go
func TestRun(t *testing.T)                    // From cmd/run_test.go
func TestSave(t *testing.T)                   // From cmd/save_test.go
func TestRestore(t *testing.T)                // From cmd/restore_test.go
func TestInit(t *testing.T)                   // From cmd/init_test.go
func TestClear(t *testing.T)                  // From cmd/clear_test.go
```

**Test Objectives:**
- All moved tests must pass in new location
- Test package changes from `cmd_test` to `cli_test`
- Imports updated to reference `internal/cli`

**Context for implementation:**
- Follow external test package pattern: package name should be `cli_test` not `cli`
- Reference CLAUDE.md testing guidelines about external test packages
- Tests should continue to test via public API (Run() function)

#### 4. Create New Main Entry Point
**File**: `cmd/claude-md/main.go`
**Changes**: New file

```go
package main

import (
	"os"

	"github.com/kapetan-io/claude-md.go/internal/cli"
)

func main() {
	args := os.Args[1:]
	exitCode := cli.Run(args, cli.RunOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	os.Exit(exitCode)
}
```

**Function Responsibilities:**
- Entry point for the claude-md binary
- Calls `cli.Run()` with command-line args and standard I/O streams
- Exits with the return code from cli.Run()
- Pattern reference: Same pattern as current main.go:9-16, just different import path

**Testing Requirements:**

No direct tests for main.go (follows standard practice for main packages).

**Test Objectives:**
- Integration tests in `test/integration_test.go` will validate the binary works end-to-end
- Existing CLI tests in `internal/cli/*_test.go` validate the Run() function

**Context for implementation:**
- This file is nearly identical to current root main.go
- Only difference is import path: `internal/cli` instead of `cmd`
- main() functions are typically not unit tested, tested via integration/e2e tests

#### 5. Delete Old cmd/ Directory
**Action**: Remove old cmd/ directory after confirming all files moved

```bash
rm -rf cmd/
# Then recreate with just the new structure
mkdir -p cmd/claude-md
```

Note: This step should only happen after step 4 is complete and cmd/claude-md/main.go exists.

#### 6. Delete Root main.go
**File**: `main.go`
**Action**: Delete file

```bash
rm main.go
```

This file is replaced by `cmd/claude-md/main.go`.

#### 7. Update Integration Test Imports
**File**: `test/integration_test.go`
**Changes**: Update import statement

Change line 10:
```go
// Old
import "github.com/kapetan-io/claude-md.go/cmd"

// New
import "github.com/kapetan-io/claude-md.go/internal/cli"
```

All references to `cmd.Run` and `cmd.RunOptions` in the file will automatically work with the new import path.

**Context for implementation:**
- This integration test uses `cmd.Run()` to test the full CLI workflow - see test/integration_test.go:65, 76, 104, 124
- Only the import path changes; all test code remains identical
- The test uses `cmd.RunOptions` which will become `cli.RunOptions` via the import alias

### Validation
- [x] Run: `go build -o claude-md ./cmd/claude-md`
- [x] Verify: Build succeeds and creates `claude-md` binary
- [x] Run: `./claude-md --help`
- [x] Verify: Help text displays correctly
- [x] Run: `go test ./...`
- [x] Verify: All tests pass with same coverage as before

**File Count Verification:**
```bash
# Should show 6 command files (not counting tests)
ls internal/cli/*.go | grep -v "_test.go" | wc -l

# Should show 6 test files
ls internal/cli/*_test.go | wc -l

# Should show 1 main.go
ls cmd/claude-md/*.go | wc -l

# Should show 0 .go files in cmd/ root
ls cmd/*.go 2>/dev/null | wc -l || echo "0"
```

Expected:
- 6 command files in internal/cli/
- 6 test files in internal/cli/
- 1 main.go in cmd/claude-md/
- 0 .go files in cmd/ root

### Context for implementation:
- **File Movement Strategy**:
  - Option 1 (Preferred): Use `git mv` to preserve git history: `git mv cmd/root.go internal/cli/root.go`
  - Option 2: Simple move if git history not critical: `mv cmd/*.go internal/cli/`
  - **Important**: After moving files, you still need to edit each file to update package declarations
- Verify no files are left in old `cmd/` before deleting
- The old binary `claude-md.go` in root should be deleted or added to .gitignore (it's a compiled binary, not source code)

## Phase 2: Update Documentation and Configuration

### Overview
Update build instructions, gitignore, and any documentation referencing the old structure.

### Changes Required:

#### 1. Update .gitignore
**File**: `.gitignore`
**Changes**: Add binary names to ignore

```gitignore
# Add to existing .gitignore
claude-md
claude-md.go
```

**Context for implementation:**
- Binary `claude-md.go` in root is currently not ignored (shouldn't be committed)
- Standard practice to ignore binaries matching the project name
- This prevents accidentally committing built binaries

#### 2. Update README.md
**File**: `README.md`
**Changes**: Update build instructions in the Installation section

Current (lines 21-27):
```bash
git clone https://github.com/kapetan-io/claude-md.go
cd claude-md.go
go build -o claude-md .
# Move to a location in your PATH
mv claude-md /usr/local/bin/
```

New:
```bash
git clone https://github.com/kapetan-io/claude-md.go
cd claude-md.go
go build -o claude-md ./cmd/claude-md
# Move to a location in your PATH
mv claude-md /usr/local/bin/
```

**Context for implementation:**
- Only change is the build command: from `go build -o claude-md .` to `go build -o claude-md ./cmd/claude-md`
- All other installation instructions remain the same
- No functional changes to the tool itself

### Validation
- [x] Run: `go build -o claude-md ./cmd/claude-md`
- [x] Verify: Binary is NOT tracked by git (`git status` shows it as ignored)
- [x] Run: `./claude-md --help`
- [x] Verify: Tool functions correctly
- [x] Verify: README instructions are accurate

### Context for implementation:
- Follow the exact build command format in the README
- Ensure .gitignore patterns work for both `claude-md` and `claude-md.go` (current misnamed binary)

## Phase 3: Final Verification and Cleanup

### Overview
Comprehensive testing to ensure the refactoring is complete and nothing is broken.

### Changes Required:

#### 1. Run Full Test Suite
**Action**: Execute all tests with coverage

```bash
go test ./... -cover
```

**Expected Results:**
- All tests pass
- Coverage should be identical to pre-refactoring
- No import errors
- No missing packages

#### 2. Build and Test Binary
**Action**: Build and manually test all commands

```bash
# Build
go build -o claude-md ./cmd/claude-md

# Test each command
./claude-md --help
./claude-md init --help
./claude-md save --help
./claude-md restore --help
./claude-md clear --help
```

**Expected Results:**
- All commands display correct help text
- Command structure is identical to before
- No errors or warnings

#### 3. Verify Directory Structure
**Action**: Confirm old cmd/ is gone and new structure is correct

```bash
# Should only show cmd/claude-md/
ls -la cmd/

# Should show cli/ among other directories
ls -la internal/

# Should NOT exist
ls main.go
```

**Expected Results:**
- `cmd/` only contains `claude-md/main.go`
- `internal/cli/` contains all command files and tests
- Root `main.go` is deleted

#### 4. Check for Stale Imports
**Action**: Search codebase for any references to old `cmd` package

```bash
# Should return no results from non-cli code
grep -r "github.com/kapetan-io/claude-md.go/cmd" . --include="*.go" --exclude-dir=".git"
```

**Expected Results:**
- Only `cmd/claude-md/main.go` imports from `internal/cli`
- No other files import the old `cmd` package path

### Testing Requirements:

No new tests needed - this phase verifies existing tests work.

**Test Objectives:**
- Confirm all moved tests pass in their new location
- Verify test coverage is maintained
- Ensure integration tests (if any) still work

### Validation
- [x] Run: `go test ./... -cover | tee /tmp/coverage-final.txt`
- [x] Verify: All tests pass, coverage matches baseline from Phase 1 Step 0
- [x] Run: `diff /tmp/coverage-baseline.txt /tmp/coverage-final.txt` to compare coverage
- [x] Run: `go build -o claude-md ./cmd/claude-md && ./claude-md init --help`
- [x] Verify: Binary works and commands function correctly
- [x] Run: `ls cmd/ main.go 2>&1`
- [x] Verify: cmd/ only has claude-md/, main.go doesn't exist (expect error on ls main.go)
- [x] Run: `grep -r "github.com/kapetan-io/claude-md.go/cmd\"" . --include="*.go" --exclude-dir=".git"`
- [x] Verify: Only found in cmd/claude-md/main.go (and possibly test/integration_test.go if not using import alias)

### Context for implementation:
- This is the final validation before considering the refactoring complete
- If any issues are found, trace back to see which phase introduced the problem
- The grep command should only find the import in cmd/claude-md/main.go
- **IDE/Editor Considerations**:
  - IntelliJ/GoLand: May need "Invalidate Caches and Restart"
  - VS Code: May need to reload window or restart Go language server
  - Run `go mod tidy` to ensure module cache is up to date

## Implementation Notes

### Import Path Changes
All import statements referencing the old `cmd` package must be updated:
- Old: `github.com/kapetan-io/claude-md.go/cmd`
- New: `github.com/kapetan-io/claude-md.go/internal/cli`

### Package Declaration Changes
- Command files: `package cmd` → `package cli`
- Test files: `package cmd_test` → `package cli_test`

### File Movement Strategy
Recommended approach:
1. Use `git mv` to preserve file history
2. Move files first, then update content
3. Run tests after each logical group of changes

### Rollback Strategy
If issues arise:
1. The changes are isolated to file locations and imports
2. Git history preserves the old structure
3. Can revert to previous commit if needed

### Testing Strategy
Use TDD approach:
1. Run tests before changes to establish baseline
2. After each phase, run `go test ./...` to catch regressions immediately
3. Manual testing of binary after final phase
