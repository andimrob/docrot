---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/output/**/*.go"
---

# Output Formatting

docrot supports two output formats: text (human-readable) and JSON (machine-readable). You can specify the format using the `--format` or `-f` flag:

```bash
docrot check --format text   # Default
docrot check --format json
```

## Text Format

The text formatter provides human-readable output with status icons and a summary.

### Status Icons

- `✓` - Fresh: documentation is up-to-date
- `✗` - Stale: documentation needs review
- `?` - Missing frontmatter: document lacks docrot configuration

### Example Output

```
✓ docs/api.md [fresh]
✗ docs/setup.md [stale]
  └─ Interval of 180d exceeded
? docs/readme.md [missing_frontmatter]

Summary: 1 fresh, 1 stale, 1 missing frontmatter
```

### Quiet Mode

Use the `--quiet` or `-q` flag to only show stale and missing frontmatter documents:

```bash
docrot check --quiet
```

In quiet mode:
- Fresh documents are not displayed
- Stale and missing frontmatter documents are still shown
- The summary is not displayed

## JSON Format

The JSON formatter outputs machine-readable structured data suitable for CI/CD integration and programmatic processing.

### Structure

```json
{
  "summary": {
    "total": 3,
    "fresh": 1,
    "stale": 1,
    "missing_frontmatter": 1
  },
  "docs": [
    {
      "path": "docs/api.md",
      "status": "fresh",
      "strategy": "interval",
      "last_reviewed": "2024-01-15",
      "expires": "2024-07-15"
    },
    {
      "path": "docs/setup.md",
      "status": "stale",
      "strategy": "interval",
      "last_reviewed": "2023-06-01",
      "stale_since": "2023-11-28",
      "reason": "Interval of 180d exceeded"
    },
    {
      "path": "docs/readme.md",
      "status": "missing_frontmatter"
    }
  ]
}
```

### Summary Object

The summary provides aggregate statistics:

| Field | Type | Description |
|-------|------|-------------|
| `total` | int | Total number of documents checked |
| `fresh` | int | Number of fresh documents |
| `stale` | int | Number of stale documents |
| `missing_frontmatter` | int | Number of documents without docrot frontmatter |

### Document Object

Each document in the `docs` array contains:

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Path to the documentation file |
| `status` | string | One of: `fresh`, `stale`, `missing_frontmatter` |
| `strategy` | string | Freshness strategy (omitted for missing frontmatter) |
| `last_reviewed` | string | Last review date in YYYY-MM-DD format |
| `expires` | string | Expiration date (for `interval` and `until_date` strategies) |
| `stale_since` | string | Date when document became stale |
| `reason` | string | Human-readable explanation of staleness |
| `changed_files` | array | List of changed files (for `code_changes` strategy) |

Note: Optional fields are omitted when not applicable.
