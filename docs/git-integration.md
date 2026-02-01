---
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/git/**/*.go"
---

# Git Integration

docrot uses git history to detect code changes for the `code_changes` freshness strategy.

## Requirements

- Git must be installed and available in PATH
- The repository must be a valid git repository (has a `.git` directory)

## How It Works

docrot queries git log to find files that changed after a document's `last_reviewed` date:

```bash
git log --since=<date> --name-only --format=%aI --diff-filter=ACMR
```

The `--diff-filter=ACMR` flag includes only:
- **A**dded files
- **C**opied files
- **M**odified files
- **R**enamed files

Deleted files are excluded since they don't require documentation updates.

## FileChangeIndex

For performance, docrot builds a `FileChangeIndex` once per check run rather than making individual git calls for each document. This index:

1. Contains all file changes since the oldest `last_reviewed` date
2. Maps file paths to their most recent change date
3. Supports efficient pattern matching against watch/ignore patterns

## Pattern Matching

Watch and ignore patterns are matched using doublestar glob syntax. The matching is performed against paths relative to the repository root.

Examples:
- `**/*.go` - All Go files in any directory
- `internal/api/**` - All files under internal/api
- `**/test/**` - All files in any test directory
