# docrot

Detect stale documentation. Keep your docs fresh.

## Installation

```bash
go install github.com/betterment/docrot@latest
```

Or build from source:

```bash
git clone https://github.com/betterment/docrot
cd docrot
go build -o bin/docrot
```

## Usage

### Check documentation freshness

```bash
docrot check [paths...]
```

Exits with code 1 if any docs are stale. Use in CI to enforce documentation freshness.

Options:
- `--config, -c` - Path to config file (default: `.docrot.yml`)
- `--format, -f` - Output format: `text` (default), `json`
- `--quiet, -q` - Only output stale docs
- `--workers, -w` - Number of parallel workers (default: CPU count)

### List all docs and their status

```bash
docrot list [paths...]
```

### Update last_reviewed date

```bash
docrot review <file> [files...]
```

Options:
- `--date, -d` - Use specific date instead of today (YYYY-MM-DD)

### Add frontmatter to docs missing it

```bash
docrot init [paths...]
```

Options:
- `--strategy, -s` - Default strategy: `interval` (default), `until_date`, `code_changes`
- `--interval, -i` - Default interval (default: `180d`)
- `--dry-run, -n` - Show what would be changed without modifying files

## Configuration

Create a `.docrot.yml` in your repository root:

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

# Number of parallel workers (0 or omit = use CPU count)
workers: 8

# Default freshness settings
defaults:
  strategy: interval
  interval: 180d
  watch:
    - "**/*.rb"
    - "**/*.go"
```

## Frontmatter Schema

Add freshness configuration to your markdown files:

### Interval Strategy

Doc expires after a specified duration since last review:

```yaml
---
freshness:
  last_reviewed: "2024-01-15"
  strategy: interval
  interval: 90d  # Supports: 30d, 12w, 3m, 1y
---
```

### Until Date Strategy

Doc expires on a specific date:

```yaml
---
freshness:
  last_reviewed: "2024-01-15"
  strategy: until_date
  expires: "2024-06-01"
---
```

### Code Changes Strategy

Doc expires when related code files change:

```yaml
---
freshness:
  last_reviewed: "2024-01-15"
  strategy: code_changes
  watch:
    - "../**/*.rb"
    - "../lib/**/*.ts"
---
```

## CI Integration

### GitHub Actions

```yaml
- name: Check documentation freshness
  run: docrot check
```

### Pre-commit Hook

```bash
#!/bin/sh
docrot check --quiet
```

## JSON Output

Use `--format json` for machine-readable output:

```json
{
  "summary": {
    "total": 3,
    "fresh": 2,
    "stale": 1,
    "missing_frontmatter": 0
  },
  "docs": [
    {
      "path": "doc/api.md",
      "status": "fresh",
      "strategy": "interval",
      "last_reviewed": "2024-01-15",
      "expires": "2024-04-14"
    },
    {
      "path": "doc/setup.md",
      "status": "stale",
      "strategy": "interval",
      "last_reviewed": "2023-06-01",
      "stale_since": "2023-11-28",
      "reason": "Interval of 180d exceeded"
    }
  ]
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All docs fresh (or no docs found) |
| 1 | One or more docs are stale |
| 2 | Configuration or usage error |

## License

MIT
