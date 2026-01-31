---
freshness:
  last_reviewed: "2026-01-31"
  strategy: code_changes
  watch:
    - "../internal/config/config.go"
---

# Configuration

This document describes the configuration options available for docrot.

## Configuration File

docrot uses a YAML configuration file (default: `.docrot.yml`) to customize its behavior. The configuration file should be placed in the root of your repository.

## Configuration Structure

The configuration file supports the following top-level options:

### `patterns`

**Type:** `[]string` (array of strings)  
**Default:**
```yaml
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"
```

Glob patterns for finding documentation files. docrot will scan all files matching these patterns to check their freshness status.

**Example:**
```yaml
patterns:
  - "docs/**/*.md"
  - "*.md"
  - "guides/**/*.markdown"
```

### `exclude`

**Type:** `[]string` (array of strings)  
**Default:**
```yaml
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
```

Glob patterns for files and directories to exclude from scanning. Use this to skip dependency directories or other files that shouldn't be checked.

**Example:**
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
**Valid values:** `warn`, `fail`, `skip`

Determines how docrot should handle documentation files that don't have freshness frontmatter:

- `warn`: Report files without frontmatter as warnings but continue (default)
- `fail`: Treat missing frontmatter as an error and exit with non-zero status
- `skip`: Silently ignore files without frontmatter

**Example:**
```yaml
on_missing_frontmatter: fail
```

### `workers`

**Type:** `int`  
**Default:** `0` (uses CPU count)

Number of parallel workers to use when scanning documentation files. Set to `0` or omit to automatically use the number of available CPU cores.

**Example:**
```yaml
workers: 4
```

### `defaults`

**Type:** `object`

Default freshness settings applied to documentation files. These can be overridden on a per-file basis using frontmatter.

#### `defaults.strategy`

**Type:** `string`  
**Default:** `"interval"`  
**Valid values:** `interval`, `until_date`, `code_changes`

Default freshness strategy to use:

- `interval`: Documentation expires after a specified duration since last review
- `until_date`: Documentation expires on a specific date
- `code_changes`: Documentation expires when related code files change

**Example:**
```yaml
defaults:
  strategy: code_changes
```

#### `defaults.interval`

**Type:** `string`  
**Default:** `"180d"`  
**Format:** `<number><unit>` where unit is `d` (days), `w` (weeks), `m` (months), or `y` (years)

Default interval for the `interval` strategy. Specifies how long documentation remains fresh after the last review.

**Example:**
```yaml
defaults:
  interval: 90d  # 90 days
```

Common intervals:
- `30d` - 30 days (1 month)
- `90d` - 90 days (3 months)
- `12w` - 12 weeks
- `6m` - 6 months
- `1y` - 1 year

#### `defaults.watch`

**Type:** `[]string` (array of strings)  
**Default:**
```yaml
defaults:
  watch:
    - "**/*.rb"
    - "**/*.go"
    - "**/*.ts"
    - "**/*.tsx"
```

Default glob patterns for files to watch when using the `code_changes` strategy. Documentation is marked as stale when any of these files are modified after the last review date.

**Example:**
```yaml
defaults:
  watch:
    - "src/**/*.py"
    - "lib/**/*.js"
```

## Complete Example

Here's a complete example configuration file:

```yaml
# Glob patterns for finding documentation
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"
  - "README.md"

# Patterns to exclude
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
  - "**/dist/**"

# What to do when a doc has no frontmatter
on_missing_frontmatter: warn

# Number of parallel workers (0 = use CPU count)
workers: 8

# Default freshness settings
defaults:
  strategy: interval
  interval: 180d
  watch:
    - "**/*.rb"
    - "**/*.go"
    - "**/*.ts"
    - "**/*.tsx"
```

## Configuration Loading

The configuration file is loaded using the following logic:

1. If the config file doesn't exist, default values are used
2. If the config file exists but is empty, default values are used
3. If the config file contains values, they are merged with defaults:
   - Specified values override defaults
   - Omitted values use the default settings

This means you only need to specify the options you want to change from the defaults.

## See Also

- [Frontmatter Schema](../README.md#frontmatter-schema) - Per-file freshness configuration
- [CLI Usage](../README.md#usage) - Command-line interface reference
