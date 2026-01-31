---
freshness:
  last_reviewed: "2026-01-31"
  strategy: code_changes
  watch:
    - "../cmd/*.go"
---

# CLI Commands Reference

This document provides a comprehensive reference for all `docrot` CLI commands.

## Global Flags

These flags are available for all commands:

- `--config, -c` - Path to config file (default: `.docrot.yml`)
- `--format, -f` - Output format: `text` (default), `json`
- `--workers, -w` - Number of parallel workers (default: CPU count)

## Commands

### `docrot check`

Check documentation files for staleness. Exits with code 1 if any docs are stale.

**Usage:**
```bash
docrot check [paths...]
```

**Flags:**
- `--quiet, -q` - Only output stale docs

**Examples:**
```bash
# Check all docs in current directory
docrot check

# Check specific directory
docrot check ./docs

# Quiet mode - only show stale docs
docrot check --quiet

# JSON output for CI/CD integration
docrot check --format json
```

**Exit Codes:**
- `0` - All docs are fresh (or no docs found)
- `1` - One or more docs are stale
- `2` - Configuration or usage error

### `docrot list`

List all documentation files and their current freshness status.

**Usage:**
```bash
docrot list [paths...]
```

**Examples:**
```bash
# List all docs
docrot list

# List docs in specific directory
docrot list ./docs

# JSON output
docrot list --format json
```

### `docrot review`

Update the `last_reviewed` date in the frontmatter of one or more documentation files.

**Usage:**
```bash
docrot review <file> [files...]
```

**Flags:**
- `--date, -d` - Use specific date instead of today (YYYY-MM-DD)

**Examples:**
```bash
# Update single file to today
docrot review docs/api.md

# Update multiple files
docrot review docs/api.md docs/setup.md

# Use specific date
docrot review --date 2024-01-15 docs/api.md
```

### `docrot add-frontmatter`

Add freshness frontmatter to documentation files that don't have it.

**Usage:**
```bash
docrot add-frontmatter [paths...]
```

**Flags:**
- `--strategy, -s` - Default strategy: `interval`, `until_date`, `code_changes` (default: `interval`)
- `--interval, -i` - Default interval for interval strategy (default: `180d`)
- `--dry-run, -n` - Show what would be changed without modifying files

**Examples:**
```bash
# Add frontmatter to all docs (dry-run)
docrot add-frontmatter --dry-run

# Add frontmatter with default settings
docrot add-frontmatter

# Add with specific strategy and interval
docrot add-frontmatter --strategy interval --interval 90d

# Add frontmatter to specific directory
docrot add-frontmatter ./docs
```

### `docrot init`

Create a default `.docrot.yml` configuration file in the current directory.

**Usage:**
```bash
docrot init
```

**Examples:**
```bash
# Create config file
docrot init
```

This creates a `.docrot.yml` with sensible defaults including:
- Document patterns for finding markdown files
- Exclusion patterns for common directories
- Default freshness settings
- Configuration for handling missing frontmatter

### `docrot version`

Print the version number of docrot.

**Usage:**
```bash
docrot version
```

## Common Workflows

### Initial Setup

```bash
# 1. Create config file
docrot init

# 2. Add frontmatter to all docs
docrot add-frontmatter --dry-run  # Preview changes
docrot add-frontmatter            # Apply changes

# 3. Check status
docrot list
```

### Regular Maintenance

```bash
# Check for stale docs
docrot check

# Review and update specific doc
docrot review docs/api.md

# List all docs and their status
docrot list
```

### CI/CD Integration

```bash
# In your CI pipeline
docrot check --format json
```

Exit code 1 indicates stale documentation, causing the CI build to fail.

## See Also

- [Configuration Guide](../README.md#configuration) - Details on `.docrot.yml` configuration
- [Frontmatter Schema](../README.md#frontmatter-schema) - Freshness frontmatter format
- [CI Integration](../README.md#ci-integration) - Examples for GitHub Actions and other CI systems
