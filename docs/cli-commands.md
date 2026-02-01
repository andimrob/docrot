---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "cmd/**/*.go"
---

# CLI Commands

This document describes all available CLI commands in docrot.

## Global Flags

All commands support the following global flags:

- `--config, -c <path>` - Path to config file (default: `.docrot.yml`)
- `--format, -f <format>` - Output format: `text`, `json` (default: `text`)
- `--workers, -w <count>` - Number of parallel workers (0 = use CPU count)

## Commands

### `docrot check [paths...]`

Check all documentation files for staleness. Exits with code 1 if any docs are stale.

This is the primary command for validating documentation freshness in CI/CD pipelines.

**Usage:**
```bash
docrot check [paths...]
```

**Flags:**
- `--quiet, -q` - Only output stale docs (suppresses fresh docs from output)

**Arguments:**
- `[paths...]` - Optional paths to check (default: current directory)

**Examples:**
```bash
# Check all docs in current directory
docrot check

# Check specific directory
docrot check ./docs

# Only show stale docs
docrot check --quiet

# Use JSON output for CI
docrot check --format json
```

**Exit Codes:**
- `0` - All docs are fresh (or no docs found)
- `1` - One or more docs are stale
- `2` - Configuration or usage error

---

### `docrot list [paths...]`

List all documentation files and their current freshness status.

Unlike `check`, this command always shows all docs regardless of their status and always exits with code 0.

**Usage:**
```bash
docrot list [paths...]
```

**Arguments:**
- `[paths...]` - Optional paths to list (default: current directory)

**Examples:**
```bash
# List all docs in current directory
docrot list

# List docs in specific directory
docrot list ./docs

# Get JSON output
docrot list --format json
```

---

### `docrot review <file> [files...]`

Update the `last_reviewed` date in the frontmatter of one or more documentation files.

Use this command after reviewing documentation to mark it as fresh.

**Usage:**
```bash
docrot review <file> [files...]
```

**Flags:**
- `--date, -d <date>` - Use specific date instead of today (format: YYYY-MM-DD)

**Arguments:**
- `<file>` - Required: First file to review
- `[files...]` - Optional: Additional files to review

**Examples:**
```bash
# Review a single file (sets last_reviewed to today)
docrot review docs/api.md

# Review multiple files at once
docrot review docs/api.md docs/setup.md docs/guide.md

# Review with specific date
docrot review docs/api.md --date 2024-01-15

# Review after updating content
docrot review docs/cli-commands.md
```

**Notes:**
- The file must already have docrot frontmatter with a `last_reviewed` field
- The date format must be YYYY-MM-DD
- Multiple files can be reviewed in a single command

---

### `docrot init`

Create a default `.docrot.yml` configuration file in the current directory.

This creates a starter configuration with sensible defaults.

**Usage:**
```bash
docrot init
```

**Examples:**
```bash
# Create config in current directory
docrot init
```

**Output:**
Creates a `.docrot.yml` file with:
- Default patterns for finding documentation (`**/doc/**/*.md`, `**/docs/**/*.md`)
- Default exclusions (`node_modules`, `vendor`)
- Default freshness settings (interval strategy with 180 day interval)

**Notes:**
- Will fail if `.docrot.yml` already exists
- The generated config should be reviewed and customized for your project

---

### `docrot add-frontmatter [paths...]`

Add freshness frontmatter to documentation files that don't have it.

This is useful when first adopting docrot in a project with existing documentation.

**Usage:**
```bash
docrot add-frontmatter [paths...]
```

**Flags:**
- `--strategy, -s <strategy>` - Default strategy to use (default: `interval`)
  - Options: `interval`, `until_date`, `code_changes`
- `--interval, -i <duration>` - Default interval for interval strategy (default: `180d`)
  - Format: `30d` (days), `12w` (weeks), `3m` (months), `1y` (years)
- `--dry-run, -n` - Show what would be changed without modifying files

**Arguments:**
- `[paths...]` - Optional paths to process (default: current directory)

**Examples:**
```bash
# Add frontmatter to all docs missing it
docrot add-frontmatter

# Preview changes without modifying files
docrot add-frontmatter --dry-run

# Use code_changes strategy by default
docrot add-frontmatter --strategy code_changes

# Set shorter interval
docrot add-frontmatter --interval 90d

# Process specific directory
docrot add-frontmatter ./docs
```

**Notes:**
- Only affects files that don't already have docrot frontmatter
- Sets `last_reviewed` to today's date
- Merges with existing YAML frontmatter if present

---

### `docrot version`

Print the version number of docrot.

**Usage:**
```bash
docrot version
```

**Examples:**
```bash
# Show version
docrot version
# Output: docrot 0.1.0
```

---

## Common Workflows

### Initial Setup

```bash
# Create config file
docrot init

# Add frontmatter to existing docs
docrot add-frontmatter --dry-run  # Preview
docrot add-frontmatter             # Apply

# Check everything is working
docrot check
```

### Regular Maintenance

```bash
# Check for stale docs
docrot check --quiet

# Review documentation after updates
docrot review docs/updated-file.md

# List all docs with their status
docrot list
```

### CI Integration

```bash
# In GitHub Actions or other CI
docrot check --format json
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success - all docs fresh (or no docs found) |
| `1` | Failure - one or more docs are stale |
| `2` | Error - configuration or usage error |

**Note:** The `list` and `review` commands always exit with code 0 on success.
