---
freshness:
  last_reviewed: "2026-01-31"
  strategy: code_changes
  watch:
    - "internal/freshness/checker.go"
---

# About docrot

docrot is a documentation freshness tracking tool that helps you keep your documentation up-to-date by detecting when docs become stale.

## Overview

Documentation tends to rot over time as code changes but docs don't get updated. docrot solves this by:

1. Adding freshness metadata to your markdown files (via YAML frontmatter)
2. Checking documentation freshness based on configurable strategies
3. Reporting which docs need review
4. Integrating with CI to enforce documentation freshness

## Freshness Strategies

docrot supports three strategies for determining when documentation becomes stale:

### 1. Interval Strategy

Documents expire after a specified time period since the last review. This is useful for general documentation that should be reviewed periodically.

```yaml
---
freshness:
  last_reviewed: "2026-01-31"
  strategy: interval
  interval: 90d  # Supports: d (days), w (weeks), m (months), y (years)
---
```

The checker calculates an expiration date by adding the interval to the `last_reviewed` date. If the current date is after the expiration date, the document is marked as stale.

### 2. Until Date Strategy

Documents expire on a specific calendar date. This is useful for time-sensitive documentation like feature announcements or temporary guides.

```yaml
---
freshness:
  last_reviewed: "2024-01-15"
  strategy: until_date
  expires: "2024-06-01"
---
```

The document becomes stale once the current date passes the `expires` date.

### 3. Code Changes Strategy

Documents expire when related source code files change. This is the most powerful strategy for keeping technical documentation synchronized with the codebase.

```yaml
---
freshness:
  last_reviewed: "2026-01-31"
  strategy: code_changes
  watch:
    - "internal/**/*.go"
    - "pkg/**/*.go"
---
```

The checker uses git history to detect when any files matching the watch patterns have been modified since the `last_reviewed` date. If changes are detected, the document is marked as stale.

**Default Watch Patterns:** If no watch patterns are specified, docrot uses these defaults:
- `**/*.rb` - Ruby files
- `**/*.go` - Go files  
- `**/*.ts` - TypeScript files
- `**/*.tsx` - TypeScript JSX files

## How the Checker Works

The freshness checker (`internal/freshness/checker.go`) is the core component that evaluates document staleness:

1. **Parse Document**: Extracts frontmatter from markdown files
2. **Validate Metadata**: Checks for required fields like `last_reviewed` and `strategy`
3. **Apply Strategy Logic**: 
   - For `interval`: Calculates expiration date and compares with current date
   - For `until_date`: Directly compares expiration date with current date
   - For `code_changes`: Queries git history for file changes matching watch patterns
4. **Return Result**: Provides status (fresh/stale/missing_frontmatter) and reason

### Performance Optimization

For code_changes strategy with many documents, the checker supports using a precomputed `FileChangeIndex` to avoid repeated git calls. This significantly improves performance when checking multiple documents that watch similar code patterns.

## Commands

- `docrot check [paths...]` - Check documentation freshness (exits 1 if stale docs found)
- `docrot list [paths...]` - List all docs and their freshness status
- `docrot review <file>` - Update the last_reviewed date to today
- `docrot add-frontmatter [paths...]` - Add freshness frontmatter to docs missing it
- `docrot init` - Create a `.docrot.yml` config file

## Exit Codes

- `0` - All docs are fresh (or no docs found)
- `1` - One or more docs are stale
- `2` - Configuration or usage error

## Configuration

Create a `.docrot.yml` file to configure default behavior:

```yaml
# Glob patterns for finding documentation
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"

# Patterns to exclude
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"

# What to do when a doc has no frontmatter: warn, fail, skip
on_missing_frontmatter: warn

# Default freshness settings
defaults:
  strategy: interval
  interval: 180d
  watch:
    - "**/*.go"
```

## Use Cases

1. **Keep API Documentation Current**: Use code_changes strategy to track when API implementations change
2. **Scheduled Documentation Reviews**: Use interval strategy for general docs that should be reviewed quarterly
3. **Deprecation Notices**: Use until_date strategy for temporary migration guides
4. **CI Integration**: Run `docrot check` in CI to block merges when docs are stale

## Architecture

- `internal/document` - Parses markdown files and extracts frontmatter
- `internal/freshness` - Contains the checker logic and strategy implementations  
- `internal/git` - Git operations for tracking file changes
- `internal/scanner` - Finds documentation files matching glob patterns
- `internal/config` - Configuration file parsing
- `cmd` - CLI commands using Cobra framework
