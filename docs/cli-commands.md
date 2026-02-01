---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "cmd/**/*.go"
---

# CLI Commands

Docrot provides several commands for managing documentation freshness.

## Available Commands

### check

Check documentation files for staleness. This is the primary command for CI/CD integration.

```bash
docrot check [paths...]
```

**Flags:**
- `--quiet, -q`: Only output stale docs (hides fresh docs from output)

**Exit codes:**
- `0`: All docs are fresh
- `1`: Stale documentation found

**Example:**
```bash
# Check all docs in current directory
docrot check

# Check specific directory
docrot check ./docs

# Quiet mode for CI (only shows stale docs)
docrot check --quiet
```

### list

List all documentation files with their freshness status.

```bash
docrot list [paths...]
```

Shows all docs regardless of freshness state. Useful for reviewing all documentation in your project.

### review

Update the `last_reviewed` date in documentation frontmatter.

```bash
docrot review <file> [files...]
```

**Flags:**
- `--date, -d`: Use specific date instead of today (format: YYYY-MM-DD)

**Example:**
```bash
# Update single file to today's date
docrot review docs/api.md

# Update multiple files
docrot review docs/*.md

# Use specific date
docrot review docs/api.md --date 2024-01-15
```

### init

Create a default `.docrot.yml` configuration file in the current directory.

```bash
docrot init
```

### version

Print version information.

```bash
docrot version
```

## Global Flags

These flags work with all commands:

- `--config, -c`: Path to config file (default: `.docrot.yml`)
- `--format, -f`: Output format: `text` or `json` (default: `text`)
- `--workers, -w`: Number of parallel workers; 0 uses CPU count (default: `0`)

**Example:**
```bash
# Use custom config file
docrot check --config custom-config.yml

# Output as JSON for automation
docrot check --format json

# Use specific number of workers
docrot check --workers 4
```
