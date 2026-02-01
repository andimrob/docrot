---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/config/**/*.go"
---

# Configuration

This document describes the configuration file format for docrot.

## Configuration File

Docrot looks for a configuration file at `.docrot.yml` in the repository root (or use `--config` to specify a different path).

## Configuration Options

### Full Example

```yaml
# Glob patterns for finding documentation files
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"

# Patterns to exclude from discovery
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"

# What to do when a document has no frontmatter
# Options: warn (default), fail, skip
on_missing_frontmatter: warn

# Number of parallel workers for processing documents
# 0 or omitting this field uses the CPU count
workers: 8

# Default freshness settings applied to documents
defaults:
  # Default freshness strategy
  # Options: interval, until_date, code_changes
  strategy: interval
  
  # Default interval for the interval strategy
  # Supports: days (d), weeks (w), months (m), years (y)
  # Examples: 30d, 12w, 3m, 1y
  interval: 180d
  
  # Default watch patterns for the code_changes strategy
  # Glob patterns for files to monitor
  watch:
    - "**/*.rb"
    - "**/*.go"
    - "**/*.ts"
    - "**/*.tsx"
```

## Field Reference

### `patterns`

**Type:** `[]string`  
**Default:** `["**/doc/**/*.md", "**/docs/**/*.md"]`

Glob patterns used to discover documentation files. Docrot will scan for markdown files matching these patterns.

**Examples:**
```yaml
patterns:
  - "docs/**/*.md"
  - "README.md"
  - "*/GUIDE.md"
```

### `exclude`

**Type:** `[]string`  
**Default:** `["**/node_modules/**", "**/vendor/**"]`

Glob patterns for paths to exclude from documentation discovery. Use this to skip directories that may contain markdown files but aren't documentation (e.g., dependencies, build artifacts).

**Examples:**
```yaml
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
  - "**/dist/**"
  - "**/.git/**"
```

### `on_missing_frontmatter`

**Type:** `string`  
**Default:** `"warn"`  
**Options:** `warn`, `fail`, `skip`

Controls behavior when a documentation file is missing docrot frontmatter:

- `warn` - Print a warning message but continue processing
- `fail` - Exit with an error code
- `skip` - Silently skip the file

**Example:**
```yaml
on_missing_frontmatter: fail
```

### `workers`

**Type:** `int`  
**Default:** `0` (uses CPU count)

Number of parallel workers to use when processing documentation files. Set to a specific number to limit parallelism, or use `0` (or omit) to use the number of CPU cores available.

**Examples:**
```yaml
# Use 4 workers
workers: 4

# Use CPU count (omit or set to 0)
workers: 0
```

### `defaults`

**Type:** `object`

Default freshness settings applied to documents that don't specify their own configuration. Individual documents can override these defaults in their frontmatter.

#### `defaults.strategy`

**Type:** `string`  
**Default:** `"interval"`  
**Options:** `interval`, `until_date`, `code_changes`

The default freshness strategy for documents:

- `interval` - Document expires after a specified duration
- `until_date` - Document expires on a specific date
- `code_changes` - Document expires when watched files change

**Example:**
```yaml
defaults:
  strategy: code_changes
```

#### `defaults.interval`

**Type:** `string`  
**Default:** `"180d"`

Default interval for the `interval` strategy. Supports days (d), weeks (w), months (m), and years (y).

**Examples:**
```yaml
defaults:
  interval: 90d   # 90 days
  # interval: 12w  # 12 weeks
  # interval: 6m   # 6 months
  # interval: 1y   # 1 year
```

#### `defaults.watch`

**Type:** `[]string`  
**Default:** `["**/*.rb", "**/*.go", "**/*.ts", "**/*.tsx"]`

Default watch patterns for the `code_changes` strategy. These glob patterns specify which files to monitor for changes that would mark documentation as stale.

**Examples:**
```yaml
defaults:
  watch:
    - "src/**/*.py"
    - "lib/**/*.js"
    - "internal/**/*.go"
```

## Minimal Configuration

If you don't provide a configuration file, docrot uses sensible defaults:

```yaml
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"

exclude:
  - "**/node_modules/**"
  - "**/vendor/**"

on_missing_frontmatter: warn

defaults:
  strategy: interval
  interval: 180d
  watch:
    - "**/*.rb"
    - "**/*.go"
    - "**/*.ts"
    - "**/*.tsx"
```

## Configuration Merging

When a configuration file is present, docrot merges it with the defaults:

- If you specify `patterns`, it completely replaces the default patterns
- If you specify `exclude`, it completely replaces the default exclude patterns
- If you specify `defaults.watch`, it completely replaces the default watch patterns
- Other fields use the file value if provided, otherwise fall back to defaults

## See Also

- [Freshness Strategies](freshness-strategies.md) - Detailed information about each strategy
- [CLI Commands](cli-commands.md) - Command-line usage
- [Frontmatter Schema](../README.md#frontmatter-schema) - Document-level configuration
