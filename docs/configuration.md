---
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/config/**/*.go"
---

# Configuration

docrot is configured via a `.docrot.yml` file in your repository root. Run `docrot init` to create one with sensible defaults.

## Configuration Options

```yaml
# Glob patterns to find documentation files
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"

# Glob patterns to exclude from scanning
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"

# What to do when a doc has no freshness frontmatter: warn, skip, or fail
on_missing_frontmatter: warn

# Number of parallel workers (0 = use CPU count)
workers: 0

# Default freshness settings for new documents
defaults:
  strategy: interval    # interval, until_date, or code_changes
  interval: 180d        # Default interval for interval strategy
  watch:                # Default watch patterns for code_changes
    - "**/*.go"
  ignore:               # Default ignore patterns for code_changes
    - "**/docs/**"
```

## Options

### `patterns`
Glob patterns that match your documentation files. Supports doublestar (`**`) for recursive matching.

### `exclude`
Glob patterns for files/directories to skip entirely. Common exclusions like `node_modules` and `vendor` are built into the scanner for performance.

### `on_missing_frontmatter`
Controls behavior when a document lacks docrot frontmatter:
- `warn` (default): Show a warning but continue
- `skip`: Silently ignore the document
- `fail`: Treat as stale (exit code 1)

### `workers`
Number of parallel workers for checking documents. Set to 0 (default) to use the number of CPU cores.

### `defaults`
Default freshness settings applied when running `add-frontmatter` or when documents don't specify their own patterns:
- `strategy`: Default freshness strategy
- `interval`: Default interval for the interval strategy
- `watch`: Default watch patterns for code_changes strategy
- `ignore`: Default ignore patterns for code_changes strategy
