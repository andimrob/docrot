---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/config/**/*.go"
---

# Configuration

Docrot is configured via a `.docrot.yml` file in your project root. Run `docrot init` to create a default configuration file.

## Configuration File Structure

```yaml
# Glob patterns to find documentation files
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"

# Glob patterns to exclude from scanning
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"

# How to handle docs without frontmatter
# Options: "warn" (default), "skip", "fail"
on_missing_frontmatter: "warn"

# Number of parallel workers (0 = use CPU count)
workers: 0

# Default settings for docs
defaults:
  # Default freshness strategy
  # Options: "interval", "until_date", "code_changes"
  strategy: "interval"
  
  # Default interval duration (for interval strategy)
  # Units: d (days), w (weeks), m (months), y (years)
  interval: "180d"
  
  # Default watch patterns (for code_changes strategy)
  watch:
    - "**/*.rb"
    - "**/*.go"
    - "**/*.ts"
    - "**/*.tsx"
```

## Configuration Options

### patterns

Array of glob patterns to discover documentation files. Uses [doublestar](https://github.com/bmatcuk/doublestar) pattern matching.

**Default:**
```yaml
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"
```

**Examples:**
```yaml
# All markdown files in project
patterns:
  - "**/*.md"

# Specific directories only
patterns:
  - "docs/**/*.md"
  - "api/**/*.md"
```

### exclude

Array of glob patterns to exclude from scanning. Useful for ignoring generated files, dependencies, or build artifacts.

**Default:**
```yaml
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
```

Docrot also automatically skips common directories: `.git`, `.svn`, `__pycache__`, `.cache`, `dist`, `build`, `tmp`, `log`, `logs`.

### on_missing_frontmatter

How to handle documentation files without docrot frontmatter.

**Options:**
- `"warn"` (default): Show as "missing frontmatter" in output but don't fail
- `"skip"`: Silently ignore files without frontmatter
- `"fail"`: Treat missing frontmatter as stale (causes check to fail)

### workers

Number of parallel workers for checking documentation. `0` automatically uses the number of CPU cores.

**Default:** `0`

### defaults

Default settings applied to all documentation files. Individual files can override these in their frontmatter.

#### defaults.strategy

Default freshness strategy to use.

**Options:**
- `"interval"`: Docs expire after a duration since last review
- `"until_date"`: Docs expire on a specific date
- `"code_changes"`: Docs expire when related code changes

**Default:** `"interval"`

#### defaults.interval

Default interval duration for the interval strategy. Format: `<number><unit>` where unit is:
- `d`: days
- `w`: weeks
- `m`: months
- `y`: years

**Default:** `"180d"` (6 months)

**Examples:**
- `"90d"`: 90 days (3 months)
- `"12w"`: 12 weeks
- `"6m"`: 6 months
- `"1y"`: 1 year

#### defaults.watch

Default glob patterns for watching code changes (used by `code_changes` strategy). Individual docs can override this in their frontmatter.

**Default:**
```yaml
watch:
  - "**/*.rb"
  - "**/*.go"
  - "**/*.ts"
  - "**/*.tsx"
```

## Loading Configuration

1. Docrot looks for `.docrot.yml` in the current directory
2. If not found, uses default configuration
3. CLI flag `--config` can specify a custom path
4. Configuration is merged with defaults (not replaced)
