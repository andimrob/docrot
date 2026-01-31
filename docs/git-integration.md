---
freshness:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/git/**/*.go"
---

# Git Integration

Docrot integrates with Git to track when code files change, enabling documentation freshness based on code modifications. This integration is a core feature for the `code_changes` freshness strategy.

## Overview

The Git integration is implemented in `internal/git/git.go` and provides the following capabilities:

- **Repository Detection**: Verifies that a directory is a valid Git repository
- **Last Commit Date**: Gets the date of the last commit that touched a specific file
- **Changed Files Tracking**: Finds files matching patterns that changed after a specific date
- **Efficient Batch Queries**: Pre-computes file change indices to avoid individual Git calls per document

## Core Components

### Git Client

The `git.Client` struct is the main interface for Git operations:

```go
type Client struct {
    repoRoot string
}
```

Create a new client with:
```go
gitClient, err := git.New(repoRoot)
```

The client verifies the repository is valid by running `git rev-parse --git-dir`.

### Key Functions

#### LastCommitDate

Returns the date of the last commit that modified a file:

```go
date, err := client.LastCommitDate(path)
```

Uses: `git log -1 --format=%aI -- <path>`

#### FilesChangedSince

Finds files matching glob patterns that changed after a given date:

```go
changed, err := client.FilesChangedSince(since, patterns, relativeTo)
```

Returns a slice of `ChangedFile` structs containing:
- `Path`: File path relative to repo root
- `Date`: Date of the last change

Uses: `git log --since=<date> --name-only --format=%aI --diff-filter=ACMR`

#### FindRepoRoot

Finds the root directory of a Git repository from any path within it:

```go
root, err := git.FindRepoRoot(path)
```

Uses: `git rev-parse --show-toplevel`

### File Change Index

For performance optimization, docrot can pre-compute a file change index:

```go
type FileChangeIndex struct {
    files map[string]time.Time
}
```

Building an index makes ONE Git call instead of multiple calls:

```go
// Get all files changed since a date (or all files if time.Time{} passed)
index, err := client.BuildFileChangeIndex(since)
```

The index provides efficient queries:

- `HasChangedSince(since, patterns) bool` - Quick check if any files changed
- `GetChangedFiles(since, patterns) []ChangedFile` - Get list of changed files
- `FileCount() int` - Number of files in the index

## Usage in Code Changes Strategy

When a document uses the `code_changes` strategy in its frontmatter:

```yaml
---
freshness:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/**/*.go"
    - "cmd/**/*.go"
---
```

The freshness checker:

1. Parses the `last_reviewed` date
2. Gets the watch patterns (or uses defaults: `**/*.rb`, `**/*.go`, `**/*.ts`, `**/*.tsx`)
3. Either:
   - Calls `FilesChangedSince()` directly for single document checks
   - Uses a pre-built `FileChangeIndex` for batch operations (more efficient)
4. Marks the doc as stale if any watched files changed after `last_reviewed`

## Git Command Details

All Git operations use `os/exec.Command` to run Git commands. The commands are:

| Operation | Git Command | Purpose |
|-----------|-------------|---------|
| Verify repo | `git rev-parse --git-dir` | Check if directory is a Git repo |
| Find root | `git rev-parse --show-toplevel` | Get repository root path |
| Last commit | `git log -1 --format=%aI -- <path>` | Get last commit date for a file |
| Changed files | `git log --since=<date> --name-only --format=%aI --diff-filter=ACMR` | List files changed since date |

The `--diff-filter=ACMR` flag includes only:
- **A**dded files
- **C**opied files  
- **M**odified files
- **R**enamed files

This excludes deleted files, which is appropriate since we're tracking files that need documentation updates.

## Pattern Matching

File patterns use the `doublestar` library for glob matching, supporting:

- `*` - Matches any characters within a path segment
- `**` - Matches zero or more directories
- `?` - Matches a single character
- `{a,b}` - Matches either pattern a or pattern b

Examples:
- `internal/**/*.go` - All Go files under internal/
- `**/*.{ts,tsx}` - All TypeScript files anywhere
- `cmd/*/main.go` - main.go in direct subdirectories of cmd/

## Performance Considerations

When checking multiple documents:

1. **Without Index** (slower): Each document triggers a separate `git log` call
2. **With Index** (faster): One `git log` call builds the index, then all documents query it

For large repositories or many documents, using the `FileChangeIndex` significantly reduces execution time.

## Error Handling

Git operations can fail for several reasons:

- Not a Git repository: Returns error from `New()` or operations
- File has no Git history: Returns "file has no git history" error
- Invalid dates: Parsing errors for date strings
- Git command failures: Returns stderr from Git command

The freshness checker handles these gracefully by marking documents as stale with an appropriate error message.
