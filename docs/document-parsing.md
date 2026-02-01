---
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/document/**/*.go"
---

# Document Parsing

docrot parses markdown files to extract freshness configuration from YAML frontmatter.

## Frontmatter Format

Freshness configuration is stored under a `docrot` key in the document's frontmatter:

```yaml
---
title: My Document
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "src/**/*.ts"
---

# My Document

Document content here...
```

## Supported Fields

| Field | Type | Description |
|-------|------|-------------|
| `strategy` | string | Freshness strategy: `interval`, `until_date`, or `code_changes` |
| `last_reviewed` | string | Date of last review in `YYYY-MM-DD` format |
| `interval` | string | Duration for interval strategy (e.g., `90d`, `6m`) |
| `expires` | string | Expiration date for until_date strategy |
| `watch` | []string | Glob patterns for files to watch (code_changes) |
| `ignore` | []string | Glob patterns for files to ignore (code_changes) |

## Parsing Behavior

- Frontmatter must be delimited by `---` at the start of the file
- The `docrot` key is optional; documents without it are reported as missing frontmatter
- All other frontmatter (like `title`, `author`, etc.) is preserved
- The document content (everything after the closing `---`) is stored separately
