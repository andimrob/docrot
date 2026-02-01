---
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "cmd/**/*.go"
---

# CLI Commands

## Global Flags

These flags apply to all commands:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `.docrot.yml` | Path to config file |
| `--format` | `-f` | `text` | Output format: `text` or `json` |
| `--workers` | `-w` | `0` | Number of parallel workers (0 = CPU count) |

## Commands

### `docrot init`

Create a `.docrot.yml` config file with sensible defaults.

```bash
docrot init
```

Fails if a config file already exists.

### `docrot check [paths...]`

Check documentation freshness. Exits with code 1 if any docs are stale.

```bash
docrot check                    # Check all docs
docrot check docs/              # Check specific directory
docrot check docs/api.md        # Check specific file
docrot check --quiet            # Only show stale docs
docrot check --format json      # JSON output
```

| Flag | Short | Description |
|------|-------|-------------|
| `--quiet` | `-q` | Only output stale docs |

### `docrot list [paths...]`

List all documentation files and their freshness status. Unlike `check`, does not exit with error code for stale docs.

```bash
docrot list
docrot list --format json
```

### `docrot review <file> [files...]`

Update the `last_reviewed` date to today (or a specified date).

```bash
docrot review docs/api.md                    # Set to today
docrot review docs/api.md --date 2024-01-15  # Set specific date
docrot review docs/*.md                      # Multiple files
```

| Flag | Short | Description |
|------|-------|-------------|
| `--date` | `-d` | Use specific date (YYYY-MM-DD) instead of today |

### `docrot add-frontmatter [paths...]`

Add freshness frontmatter to documents that don't have it.

```bash
docrot add-frontmatter                          # Add to all docs
docrot add-frontmatter --strategy code_changes  # Use code_changes strategy
docrot add-frontmatter --dry-run                # Preview changes
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--strategy` | `-s` | `interval` | Default strategy |
| `--interval` | `-i` | `180d` | Default interval |
| `--dry-run` | `-n` | `false` | Show what would change |

### `docrot files <doc-path> [doc-path...]`

List all files in a document's domain (matching its watch patterns).

```bash
docrot files docs/api.md
docrot files docs/api.md --format json
```

Useful for understanding what code a document covers or for feeding to an LLM.

### `docrot version`

Print the version number.

```bash
docrot version
```
