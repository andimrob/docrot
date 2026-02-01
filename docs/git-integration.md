---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/git/**/*.go"
---

# Git Integration

Docrot uses git to track when files change, enabling the `code_changes` freshness strategy. This document explains how docrot integrates with git repositories.

## Overview

The git integration provides functionality to:
- Verify a directory is a git repository
- Find the root of a git repository
- Get the last commit date for a specific file
- Find files matching patterns that changed after a given date
- Build an efficient index of all file changes for batch queries

## Git Client

### Creating a Client

The `git.Client` requires a repository root directory:

```go
client, err := git.New("/path/to/repo")
if err != nil {
    // Not a git repository
}
```

The `New` function verifies that the path is a valid git repository by running `git rev-parse --git-dir`.

### Finding Repository Root

To find the git repository root from any path within the repository:

```go
repoRoot, err := git.FindRepoRoot("/path/to/repo/subdir")
if err != nil {
    // Not inside a git repository
}
// repoRoot is now the absolute path to the repository root
```

This uses `git rev-parse --show-toplevel` to locate the repository root.

## Querying File Changes

### Last Commit Date

Get the date of the last commit that touched a specific file:

```go
date, err := client.LastCommitDate("docs/readme.md")
if err != nil {
    // File has no git history or doesn't exist
}
```

This executes `git log -1 --format=%aI -- <path>` and parses the ISO 8601 timestamp.

### Files Changed Since Date

Find files matching patterns that changed after a given date:

```go
patterns := []string{"**/*.go", "internal/**/*.rb"}
since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

changes, err := client.FilesChangedSince(since, patterns, repoRoot)
if err != nil {
    // Error querying git
}

for _, change := range changes {
    fmt.Printf("File: %s, Changed: %s\n", change.Path, change.Date)
}
```

This uses `git log --since=<date> --name-only --format=%aI --diff-filter=ACMR` and filters results using [doublestar](https://github.com/bmatcuk/doublestar) glob patterns.

The `--diff-filter=ACMR` flag includes:
- **A**: Added files
- **C**: Copied files
- **M**: Modified files
- **R**: Renamed files

Deleted files are excluded.

## Performance Optimization: FileChangeIndex

When checking multiple documents, making individual git calls for each one is inefficient. The `FileChangeIndex` solves this by making a single git query and caching results.

### Building an Index

```go
// Get all files ever committed
index, err := client.BuildFileChangeIndex(time.Time{})

// Or get files changed since a specific date
since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
index, err := client.BuildFileChangeIndex(since)
```

The index contains a map of file paths to their most recent change dates.

### Checking if Files Changed

Check if any files matching patterns changed after a date:

```go
patterns := []string{"src/**/*.go"}
lastReviewed := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

if index.HasChangedSince(lastReviewed, patterns) {
    fmt.Println("Source code has changed since last review")
}
```

### Getting Changed Files

Retrieve all files matching patterns that changed after a date:

```go
patterns := []string{"lib/**/*.rb", "app/**/*.rb"}
lastReviewed := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

changes := index.GetChangedFiles(lastReviewed, patterns)
for _, change := range changes {
    fmt.Printf("File: %s, Changed: %s\n", change.Path, change.Date)
}
```

### Index Statistics

Get the total number of files in the index:

```go
count := index.FileCount()
fmt.Printf("Index contains %d files\n", count)
```

## Pattern Matching

All pattern matching uses the [doublestar](https://github.com/bmatcuk/doublestar) library, which supports:

- `*` - matches any characters except `/`
- `**` - matches any characters including `/` (recursive)
- `?` - matches a single character
- `[abc]` - matches one character from the set
- `{a,b}` - matches either pattern

Examples:
- `**/*.go` - all Go files in any directory
- `internal/**/*.go` - Go files in internal/ and subdirectories
- `src/{app,lib}/**/*.rb` - Ruby files in src/app/ or src/lib/
- `docs/*.md` - Markdown files directly in docs/ (not subdirectories)

## Git Commands Used

Docrot executes these git commands:

| Purpose | Command |
|---------|---------|
| Verify repo | `git rev-parse --git-dir` |
| Find repo root | `git rev-parse --show-toplevel` |
| Last commit date | `git log -1 --format=%aI -- <path>` |
| Changed files | `git log --since=<date> --name-only --format=%aI --diff-filter=ACMR` |

All commands use the `--format=%aI` flag to output ISO 8601 author dates for consistent parsing.

## Error Handling

Common errors:

- **Not a git repository**: Returned when `New()` or `FindRepoRoot()` is called on a path outside a git repository
- **File has no git history**: Returned by `LastCommitDate()` when a file exists but has never been committed
- **Git command failed**: Returned when git commands exit with non-zero status (e.g., git not installed, repository corrupted)

## How Docrot Uses Git

When processing a document with `strategy: code_changes`:

1. Parse the document's frontmatter to get `last_reviewed` date and `watch` patterns
2. Build a `FileChangeIndex` for efficient batch queries
3. Use `HasChangedSince()` to check if any watched files changed after `last_reviewed`
4. If changes detected, mark the document as stale
5. Use `GetChangedFiles()` to report which specific files triggered the staleness

This approach minimizes git operations while checking many documents.
