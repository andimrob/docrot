---
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/freshness/**/*.go"
---

# Freshness Strategies

docrot supports three strategies for determining when documentation becomes stale.

## interval

Marks documentation as stale after a specified duration since the last review.

```yaml
docrot:
  strategy: interval
  last_reviewed: "2026-01-31"
  interval: 90d
```

Supported interval units:
- `d` - days (e.g., `30d`)
- `w` - weeks (e.g., `12w`)
- `m` - months, approximated as 30 days (e.g., `6m`)
- `y` - years, approximated as 365 days (e.g., `1y`)

## until_date

Marks documentation as stale after a specific date. Useful for time-sensitive docs like release notes or deprecation notices.

```yaml
docrot:
  strategy: until_date
  expires: "2024-06-30"
```

## code_changes

Marks documentation as stale when related source code changes. This is the most powerful strategy for keeping docs in sync with code.

```yaml
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/api/**/*.go"
    - "pkg/client/*.go"
  ignore:
    - "**/*_test.go"
```

### Watch and Ignore Patterns

Patterns use doublestar glob syntax:
- `**` matches any number of directories
- `*` matches any characters within a path segment
- `?` matches a single character

### Smart Defaults

When no `watch` or `ignore` patterns are specified, docrot computes smart defaults based on the document's location:

- Documents in a `docs/` directory watch their parent subsystem
- Documents ignore their own docs directory to avoid self-triggering

For example, `subsystem/docs/readme.md` will automatically watch `subsystem/**/*` and ignore `subsystem/docs/**`.

### Pattern Priority

1. Frontmatter patterns (explicit `watch`/`ignore` in the document)
2. Config defaults (from `.docrot.yml` `defaults` section)
3. Smart defaults (computed from document location)
