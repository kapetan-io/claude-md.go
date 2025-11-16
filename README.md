# claude-md

A CLI tool for managing CLAUDE.md files across git repositories using centralized storage with symlinks.

## Overview

`claude-md` helps you manage CLAUDE.md files in your git repositories by storing them in a centralized location and creating symlinks in your repositories. This ensures that your CLAUDE.md files persist across repository cleanup operations (like `git clean -fdx`) and can be easily restored.

## Features

- **Centralized Storage**: All CLAUDE.md files are stored in `~/.claude/claude-md/<user>/<repo>/`
- **Symlink Management**: Converts files to symlinks pointing to storage
- **Case-Insensitive Matching**: Finds CLAUDE.md files regardless of case (CLAUDE.md, claude.md, Claude.MD, etc.)
- **Safe Operations**: Skips conflicts with warnings, never overwrites without permission
- **Git Integration**: Automatically detects repository and user from git configuration

## Installation

### From Source

```bash
git clone https://github.com/kapetan-io/claude-md.go
cd claude-md.go
go build -o claude-md ./cmd/claude-md
# Move to a location in your PATH
mv claude-md /usr/local/bin/
```

## Requirements

- Must be run from within a git repository
- Repository must have an `origin` remote configured
- Git user.email must be configured

## Usage

### Initialize Storage

Create the storage directory structure:

```bash
claude-md init
```

This creates `~/.claude/claude-md/<user>/<repo>/` where:
- `<user>` is extracted from `git config user.email` (part before @)
- `<repo>` is extracted from the origin remote URL

### Save CLAUDE.md Files

Find all CLAUDE.md files in your repository and convert them to symlinks:

```bash
claude-md save
```

This will:
1. Find all CLAUDE.md files (case-insensitive) in the repository
2. Copy each file to storage
3. Replace the original with a symlink to the stored copy

Files already converted to symlinks are skipped.

### Restore CLAUDE.md Files

Restore CLAUDE.md files as symlinks from storage:

```bash
claude-md restore
```

This is useful after:
- Cloning a repository
- Running `git clean -fdx`
- Switching branches that don't have CLAUDE.md files

### Clear Symlinks

Remove all CLAUDE.md symlinks from the repository:

```bash
claude-md clear
```

**Note**: This only removes symlinks. Files remain in storage and can be restored later.

## Storage Structure

Files are stored using the following structure:

```
~/.claude/claude-md/
└── <user>/
    └── <repo>/
        ├── CLAUDE.md                    # Root-level file
        └── source~go~api~CLAUDE.md     # Nested file from source/go/api/
```

Path components are joined with `~` for nested files:
- `source/go/api/CLAUDE.md` → `source~go~api~CLAUDE.md`
- `docs/CLAUDE.md` → `docs~CLAUDE.md`

## Limitations

- Directory names containing `~` are not supported (will error)
- Repository must have an `origin` remote
- Only works within git repositories

## Example Workflow

```bash
# Initial setup in a repository
cd my-project
claude-md init
claude-md save

# After git clean or clone
cd my-project
claude-md restore

# Temporarily remove symlinks (files stay in storage)
claude-md clear

# Restore them later
claude-md restore
```

## Help

For detailed help on any command:

```bash
claude-md --help
claude-md init --help
claude-md save --help
claude-md restore --help
claude-md clear --help
```

## License

Copyright © 2024 Kapetan IO

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
