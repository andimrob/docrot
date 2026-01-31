---
freshness:
  last_reviewed: "2026-01-31"
  strategy: code_changes
  watch:
    - "../internal/freshness/**/*.go"
---

# Freshness Strategies

This document describes the freshness strategies available in docrot for managing documentation staleness.

## Overview

Docrot supports three strategies for determining when documentation becomes stale:

1. **interval** - Time-based expiration
2. **until_date** - Explicit expiration date
3. **code_changes** - Code change detection

## Interval Strategy

The `interval` strategy marks documentation as stale after a specified duration has passed since the last review.

### Configuration

```yaml
---
freshness:
  last_reviewed: "2024-01-15"
  strategy: interval
  interval: 90d
---
```

### Supported Intervals

- `d` - Days (e.g., `30d`)
- `w` - Weeks (e.g., `4w`)
- `m` - Months (approximately 30 days, e.g., `3m`)
- `y` - Years (365 days, e.g., `1y`)

### Behavior

- Documentation is **fresh** until `last_reviewed + interval`
- Documentation becomes **stale** after that date
- The stale reason includes the expiration date
- The `expires` field shows when the document will expire (if still fresh)

### Example

If `last_reviewed` is `2024-01-15` and `interval` is `90d`, the document expires on `2024-04-14`.

## Until Date Strategy

The `until_date` strategy marks documentation as stale after a specific calendar date.

### Configuration

```yaml
---
freshness:
  last_reviewed: "2024-01-15"
  strategy: until_date
  expires: "2024-06-01"
---
```

### Behavior

- Documentation is **fresh** until the `expires` date
- Documentation becomes **stale** after that date
- Useful for time-sensitive documentation (e.g., migration guides, deprecation notices)
- The `expires` field shows the explicit expiration date
- The `stale_since` field equals the `expires` date when stale

### Use Cases

- Migration deadlines
- Deprecation timelines
- Event-specific documentation
- Temporary features or processes

## Code Changes Strategy

The `code_changes` strategy marks documentation as stale when related code files change after the last review date.

### Configuration

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

### Behavior

- Documentation is **fresh** if no watched files changed since `last_reviewed`
- Documentation becomes **stale** if any watched files changed
- The stale reason includes the first changed file and its change date
- The `stale_since` field shows when the first change occurred

### Watch Patterns

Watch patterns use glob syntax:
- `*` matches any characters within a path segment
- `**` matches any characters across multiple path segments
- Patterns are relative to the documentation file location

### Default Watch Patterns

If no `watch` patterns are specified, docrot uses these defaults:

```go
var DefaultWatchPatterns = []string{"**/*.rb", "**/*.go", "**/*.ts", "**/*.tsx"}
```

### Use Cases

- API documentation (watch API implementation files)
- Configuration guides (watch config parsing code)
- Architecture diagrams (watch relevant modules)
- Tutorial code (watch example code files)

## Performance Optimization

### File Change Index

For the `code_changes` strategy, docrot supports batch optimization through `CheckWithIndex`:

- When checking multiple documents, docrot can build a single git history index
- This avoids repeated git operations for each document
- The index contains all file changes with timestamps
- Documents query the index for their watch patterns

This optimization is transparent and automatically used by the `docrot check` command when processing multiple documents.

## Implementation Details

### Status Values

Documents can have three status values:

- `fresh` - Documentation is up to date
- `stale` - Documentation needs review
- `missing_frontmatter` - No freshness configuration found

### Date Format

All dates use ISO 8601 format: `YYYY-MM-DD` (e.g., `2024-01-15`)

### Error Handling

If a document has invalid configuration, it's marked as `stale` with a reason explaining the error:

- Invalid date format in `last_reviewed` or `expires`
- Invalid interval format
- Missing git client for `code_changes` strategy
- Git errors when checking file history

### Result Fields

Each freshness check produces a `Result` with:

- `path` - Document file path
- `status` - `fresh`, `stale`, or `missing_frontmatter`
- `strategy` - The strategy used (`interval`, `until_date`, `code_changes`)
- `last_reviewed` - The last review date
- `expires` - When the document expires (for `interval` and `until_date`)
- `stale_since` - When the document became stale
- `reason` - Human-readable explanation of status

## Best Practices

### Choosing a Strategy

**Use `interval` when:**
- Documentation needs regular review regardless of changes
- You want consistent review cycles
- The content is process or policy documentation

**Use `until_date` when:**
- Documentation has a known expiration date
- The content is time-sensitive
- You want to ensure review before a specific deadline

**Use `code_changes` when:**
- Documentation describes code behavior
- Code changes invalidate the documentation
- You want automatic staleness detection based on related changes

### Setting Intervals

Common interval values:
- `30d` - Fast-changing features
- `90d` - Standard documentation
- `180d` - Stable features
- `1y` - Rarely changing content

### Watch Pattern Tips

- Keep patterns specific to avoid false positives
- Watch implementation code, not tests (usually)
- Use relative paths from the documentation file
- Consider watching multiple file types if relevant

### Review Workflow

1. Run `docrot check` to find stale documentation
2. Review and update the documentation content
3. Run `docrot review <file>` to update `last_reviewed` to today
4. The document becomes fresh again

## Testing

The freshness checker includes comprehensive tests:

- `TestCheck_IntervalStrategy_Fresh` - Verifies fresh interval documents
- `TestCheck_IntervalStrategy_Stale` - Verifies stale interval documents
- `TestCheck_UntilDateStrategy_Fresh` - Verifies fresh until_date documents
- `TestCheck_UntilDateStrategy_Stale` - Verifies stale until_date documents
- `TestCheck_CodeChangesStrategy_Fresh` - Verifies fresh code_changes documents
- `TestCheck_CodeChangesStrategy_Stale` - Verifies stale code_changes documents
- `TestCheck_IntervalParsing` - Verifies interval format parsing
- `TestCheckWithIndex_*` - Verifies batch optimization with file change index

All tests use actual git repositories to ensure realistic behavior.
