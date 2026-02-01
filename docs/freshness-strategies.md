---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/freshness/**/*.go"
---

# Freshness Strategies

Docrot supports three strategies for determining when documentation becomes stale. Each strategy is configured in the frontmatter of your markdown files.

## Strategy Types

### 1. Interval Strategy

The `interval` strategy marks documentation as stale after a specified duration since it was last reviewed.

**Frontmatter configuration:**

```yaml
---
docrot:
  last_reviewed: "2026-02-01"
  strategy: interval
  interval: 90d  # Supports: d (days), w (weeks), m (months), y (years)
---
```

**Supported interval formats:**
- `30d` - Days (e.g., 30 days)
- `12w` - Weeks (e.g., 12 weeks = 84 days)
- `3m` - Months (e.g., 3 months ≈ 90 days)
- `1y` - Years (e.g., 1 year = 365 days)

**When to use:** Best for documentation that needs periodic review regardless of code changes, such as:
- Process documentation
- Getting started guides
- General conceptual documentation

**Result fields:**
- `status`: "fresh" or "stale"
- `expires`: Date when the document will become stale (if fresh)
- `stale_since`: Date when the document became stale (if stale)
- `reason`: Explanation of why the document is stale

### 2. Until Date Strategy

The `until_date` strategy marks documentation as stale after a specific date.

**Frontmatter configuration:**

```yaml
---
docrot:
  last_reviewed: "2026-02-01"
  strategy: until_date
  expires: "2024-06-01"
---
```

**When to use:** Best for time-sensitive documentation with known expiration dates, such as:
- Event-specific documentation
- Temporary feature flags
- Beta program documentation
- Migration guides with deadlines

**Result fields:**
- `status`: "fresh" or "stale"
- `expires`: The specified expiration date
- `stale_since`: The expiration date (if stale)
- `reason`: Explanation including the expiration date

### 3. Code Changes Strategy

The `code_changes` strategy marks documentation as stale when related code files change after the last review date.

**Frontmatter configuration:**

```yaml
---
docrot:
  last_reviewed: "2026-02-01"
  strategy: code_changes
  watch:
    - "**/*.go"
    - "internal/api/**/*.ts"
    - "lib/handlers/**/*.rb"
---
```

**Default watch patterns:**
If no `watch` patterns are specified, the following defaults are used:
- `**/*.rb` - Ruby files
- `**/*.go` - Go files
- `**/*.ts` - TypeScript files
- `**/*.tsx` - TypeScript JSX files

**Pattern matching:**
- Patterns use glob syntax with `**` for recursive directory matching
- Patterns are matched against file paths relative to the repository root
- Multiple patterns can be specified to watch different areas of the codebase

**When to use:** Best for API documentation, code examples, and technical documentation tied to specific implementation details, such as:
- API endpoint documentation
- Code usage examples
- Architecture documentation
- Developer guides

**Result fields:**
- `status`: "fresh" or "stale"
- `reason`: Includes the first changed file and its change date
- `changed_files`: Array of all files that changed since last review
- `stale_since`: Date of the first code change

**Performance optimization:**
The checker supports an optional `CheckWithIndex` method that accepts a precomputed `FileChangeIndex`. This avoids individual git calls when checking multiple documents with the `code_changes` strategy, significantly improving performance for batch operations.

## Choosing a Strategy

Consider these factors when selecting a strategy:

1. **Use `interval`** when:
   - Documentation needs regular review regardless of code changes
   - You want to ensure periodic updates to evolving best practices
   - The documentation is process or policy-oriented

2. **Use `until_date`** when:
   - You know the exact date when documentation becomes outdated
   - The documentation relates to temporary features or events
   - There's a specific deadline for review or removal

3. **Use `code_changes`** when:
   - Documentation closely reflects implementation details
   - You want to be notified when related code changes
   - The documentation includes code examples or API references

## Status Values

All strategies can return these status values:

- `fresh`: Documentation is up to date
- `stale`: Documentation needs review and updating
- `missing_frontmatter`: No docrot configuration found in the frontmatter

## Error Handling

If a strategy encounters an error (invalid date format, missing git client for `code_changes`, unknown strategy), the status is set to `stale` with a descriptive reason explaining the error.
