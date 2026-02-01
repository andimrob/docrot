---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/git/**/*.go"
---

# Git Integration

Docrot integrates with Git to track when code files change, enabling the `code_changes` freshness strategy.

## Git Client

The git package provides a client for querying Git history without shelling out to the git command repeatedly.

### Initialization

```go
gitClient, err := git.New(rootDir)
```

The client operates on a Git repository rooted at `rootDir`. If the directory is not a Git repository, the client returns an error (handled gracefully - docs using `code_changes` strategy will be marked as stale).

## Core Methods

### LastCommitDate

Get the last commit date for a specific file:

```go
date, err := gitClient.LastCommitDate("docs/api.md")
```

Uses: `git log -1 --format=%aI -- <path>`

### FilesChangedSince

Get all files matching patterns that changed after a specific date:

```go
patterns := []string{"**/*.go", "**/*.ts"}
since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
files, err := gitClient.FilesChangedSince(since, patterns)
```

Uses: `git log --since=<date> --name-only --pretty=format:`

## Performance Optimization: FileChangeIndex

For the `code_changes` strategy, docrot needs to check if files matching watch patterns changed after each document's `last_reviewed` date. Naively, this would require N git calls (one per document).

### The Problem

Checking 100 documents with `code_changes` strategy:
- **Naive approach**: 100 git calls
- **Performance**: Slow, especially in large repos

### The Solution: BuildFileChangeIndex

Build a reusable index of all file changes with a single git call:

```go
index, err := gitClient.BuildFileChangeIndex(earliestDate)
```

1. Makes ONE `git log` call to get all changes since the earliest `last_reviewed` date
2. Parses output into a map: `filename → last_modified_date`
3. Returns `FileChangeIndex` that can be queried repeatedly

### Using the Index

```go
// Build index once
index, _ := gitClient.BuildFileChangeIndex(since)

// Query multiple times
patterns := []string{"**/*.go"}
files := index.GetChangedFiles(since, patterns)
```

The index provides O(1) lookup instead of shelling out to git for each query.

## Git Log Format

Docrot uses specific git log formats for parsing:

### Date Format
- `--format=%aI`: ISO 8601 timestamp (e.g., `2024-01-15T10:30:00-08:00`)
- Parsed with Go's `time.RFC3339` format

### File List
- `--name-only`: Shows only filenames (not diffs)
- `--pretty=format:`: Omits commit metadata
- Combined with date: `--pretty=format:%aI` to interleave dates

## Integration with code_changes Strategy

When checking documents with `code_changes` strategy:

1. **Collect all documents** using `code_changes`
2. **Find earliest** `last_reviewed` date among them
3. **Build index once**: `BuildFileChangeIndex(earliest)`
4. **For each document**: Use `index.GetChangedFiles(doc.LastReviewed, doc.Watch)`
5. **Mark as stale** if any files returned

### Example Flow

```go
// Pseudo-code for checker
var earliestDate time.Time
var codeChangeDocs []Document

// Phase 1: Identify code_changes docs
for _, doc := range docs {
    if doc.Strategy == "code_changes" {
        codeChangeDocs = append(codeChangeDocs, doc)
        if doc.LastReviewed.Before(earliestDate) {
            earliestDate = doc.LastReviewed
        }
    }
}

// Phase 2: Build index once
index := gitClient.BuildFileChangeIndex(earliestDate)

// Phase 3: Check all docs using the same index
for _, doc := range codeChangeDocs {
    changedFiles := index.GetChangedFiles(doc.LastReviewed, doc.Watch)
    if len(changedFiles) > 0 {
        doc.Status = StatusStale
    }
}
```

## Error Handling

- **Not a Git repo**: Gracefully handled; docs using `code_changes` marked as stale
- **Git command errors**: Treated as no changes (doc remains fresh)
- **Parse errors**: Logged and skipped; affected docs may be incorrectly marked fresh

## Performance Characteristics

| Approach | Git Calls | Suitable For |
|----------|-----------|--------------|
| Individual queries | N (one per doc) | Small number of docs |
| FileChangeIndex | 1 | Large number of docs |

In practice, docrot always uses `FileChangeIndex` when any document uses `code_changes` strategy, ensuring optimal performance regardless of repository size.

## Example Git Commands

Docrot internally runs commands like:

```bash
# Get last commit date for a file
git log -1 --format=%aI -- docs/api.md

# Get all changes since date
git log --since="2024-01-15T00:00:00Z" --name-only --pretty=format:%aI

# Example index build output:
# 2024-02-01T10:00:00Z
# src/api.go
# src/handler.go
# 2024-01-20T15:30:00Z
# README.md
# docs/guide.md
```

The parser converts this into:
```
src/api.go → 2024-02-01T10:00:00Z
src/handler.go → 2024-02-01T10:00:00Z
README.md → 2024-01-20T15:30:00Z
docs/guide.md → 2024-01-20T15:30:00Z
```

## Pattern Matching

Git returns full file paths. The index uses [doublestar](https://github.com/bmatcuk/doublestar) for glob pattern matching:

```go
patterns := []string{"internal/**/*.go", "cmd/**/*.go"}
// Matches: internal/api/handler.go
// Matches: cmd/server/main.go
// Doesn't match: pkg/util/helper.go
```
