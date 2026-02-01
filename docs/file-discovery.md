---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/scanner/**/*.go"
---

# File Discovery

Docrot uses an optimized file scanner to discover documentation files in your repository. The scanner leverages glob patterns and implements several performance optimizations to handle large codebases efficiently.

## Overview

The scanner is implemented in `internal/scanner/scanner.go` and uses the [doublestar](https://github.com/bmatcuk/doublestar) library for glob pattern matching, which supports `**` wildcards for recursive directory matching.

## How It Works

### Pattern Matching

The scanner accepts two types of patterns:

1. **Include Patterns**: Glob patterns that specify which files to scan (e.g., `**/doc/**/*.md`, `**/docs/**/*.md`)
2. **Exclude Patterns**: Glob patterns that specify which files or directories to ignore (e.g., `**/node_modules/**`, `**/vendor/**`)

### Scanning Process

The scanner performs the following steps:

1. **Extract Target Directories**: Parses include patterns to identify literal directory names (e.g., "doc", "docs" from patterns like `**/doc/**/*.md`)
2. **Walk Directory Tree**: Traverses the file system starting from the root directory
3. **Early Pruning**: Skips directories that match exclude patterns or common build artifacts (see Performance Optimizations)
4. **Filter by Extension**: Only processes `.md` files
5. **Target Directory Check**: Verifies files are in one of the target directories
6. **Pattern Matching**: Confirms files match at least one include pattern
7. **Final Exclusion Check**: Applies exclude patterns to the file path

### Performance Optimizations

The scanner includes several optimizations for large repositories:

#### 1. Early Directory Pruning

Directories are skipped entirely if they match exclude patterns or are in the hardcoded skip list:
- `node_modules` - Node.js dependencies
- `vendor` - Go/Ruby dependencies
- `.git`, `.svn` - Version control
- `__pycache__`, `.cache` - Python/general cache
- `dist`, `build` - Build output
- `tmp`, `log`, `logs` - Temporary/log files

This prevents scanning thousands of files in dependency directories.

#### 2. Target Directory Filtering

Files are only processed if they're in directories extracted from the include patterns. For example, with pattern `**/docs/**/*.md`, only files in directories named "docs" are considered.

#### 3. Hidden Directory Skipping

Directories starting with `.` (except `.git` which is explicitly handled) are automatically skipped unless they're the root directory.

## Usage Example

```go
scanner := scanner.New(
    "/path/to/repo",
    []string{"**/doc/**/*.md", "**/docs/**/*.md"},  // Include patterns
    []string{"**/node_modules/**", "**/vendor/**"}, // Exclude patterns
)

files, err := scanner.Scan()
if err != nil {
    // Handle error
}

// files contains absolute paths to all matching documentation files
```

## Configuration

In `.docrot.yml`, you can configure the patterns:

```yaml
# Glob patterns for finding documentation
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"

# Patterns to exclude
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
  - "**/tmp/**"
```

## Pattern Syntax

Patterns support the following wildcards:

- `*` - Matches any characters within a path segment
- `**` - Matches any characters across multiple path segments
- `?` - Matches a single character
- `{a,b}` - Matches either `a` or `b`

### Examples

- `**/docs/**/*.md` - All .md files in any "docs" directory at any depth
- `doc/*.md` - All .md files directly in a "doc" directory at root
- `**/*.md` - All .md files anywhere in the repository
- `docs/{api,guide}/*.md` - Files in docs/api or docs/guide subdirectories

## Implementation Details

### Key Methods

- `New(root, patterns, exclude)` - Creates a new scanner instance
- `Scan()` - Executes the scan and returns matching file paths
- `extractTargetDirs()` - Extracts literal directory names from patterns
- `isInTargetDir()` - Checks if a file is in a target directory
- `shouldSkipDir()` - Determines if a directory should be skipped entirely
- `matchesPattern()` - Checks if a file matches any include pattern
- `isExcluded()` - Checks if a file matches any exclude pattern

### Dependencies

- `github.com/bmatcuk/doublestar/v4` - Glob pattern matching with `**` support
- Standard library: `os`, `path/filepath`, `io/fs`, `strings`
