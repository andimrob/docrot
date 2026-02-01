---
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/output/**/*.go"
---

# Output Formatting

docrot supports two output formats: text (default) and JSON.

## Text Format

The default human-readable format:

```
✓ docs/getting-started.md [fresh]
✗ docs/api-reference.md [stale]
  ├─ src/api/endpoints.go
  └─ src/api/handlers.go
? docs/changelog.md [missing_frontmatter]
  └─ No freshness frontmatter found

Summary: 1 fresh, 1 stale, 1 missing frontmatter
```

Status icons:
- `✓` - fresh
- `✗` - stale
- `?` - missing frontmatter

For stale documents using `code_changes` strategy, changed files are listed with tree connectors (`├─` and `└─`).

### Quiet Mode

Use `--quiet` or `-q` to only show stale documents:

```bash
docrot check --quiet
```

## JSON Format

Machine-readable JSON output for CI/CD integration:

```bash
docrot check --format json
```

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
      "path": "docs/getting-started.md",
      "status": "fresh",
      "strategy": "interval",
      "last_reviewed": "2024-01-15",
      "expires": "2024-04-15"
    },
    {
      "path": "docs/api-reference.md",
      "status": "stale",
      "strategy": "code_changes",
      "last_reviewed": "2024-01-01",
      "stale_since": "2024-01-20",
      "reason": "Code changed: src/api/endpoints.go (2024-01-20)",
      "changed_files": ["src/api/endpoints.go", "src/api/handlers.go"]
    }
  ]
}
```
