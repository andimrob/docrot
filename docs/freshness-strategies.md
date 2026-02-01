---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/freshness/**/*.go"
---

# Freshness Strategies

Docrot supports three strategies for determining if documentation is stale. Each strategy has different use cases and configuration requirements.

## interval

Documentation expires after a specified duration since the last review.

### Configuration

```yaml
---
docrot:
  strategy: interval
  last_reviewed: "2026-02-01"
  interval: "90d"  # Required for interval strategy
---
```

### Format

Interval format: `<number><unit>`

**Units:**
- `d`: days
- `w`: weeks  
- `m`: months (30 days each)
- `y`: years (365 days each)

**Examples:**
- `"30d"`: 30 days
- `"12w"`: 12 weeks (84 days)
- `"3m"`: 3 months (90 days)
- `"1y"`: 1 year (365 days)

### How It Works

1. Parse `last_reviewed` date
2. Parse `interval` duration
3. Calculate expiration: `last_reviewed + interval = expires_at`
4. Compare: if `today > expires_at`, mark as stale

### Use Cases

- Regular review cycles (e.g., quarterly, annually)
- Process documentation that should be reviewed periodically
- API documentation that should be audited regularly

### Example

```yaml
---
docrot:
  strategy: interval
  last_reviewed: "2026-02-01"
  interval: "180d"  # Review every 6 months
---

# API Authentication Guide
```

Status:
- Fresh: if today is before 2026-08-01
- Stale: if today is on or after 2026-08-01

## until_date

Documentation expires on a specific calendar date.

### Configuration

```yaml
---
docrot:
  strategy: until_date
  last_reviewed: "2026-02-01"
  expires: "2026-12-31"  # Required for until_date strategy
---
```

### Format

Date format: `YYYY-MM-DD` (ISO 8601)

### How It Works

1. Parse `expires` date
2. Compare: if `today > expires`, mark as stale

### Use Cases

- Time-sensitive documentation (e.g., holiday schedules, temporary features)
- Documentation tied to specific releases or deadlines
- Deprecation notices with known end dates

### Example

```yaml
---
docrot:
  strategy: until_date
  last_reviewed: "2026-02-01"
  expires: "2026-12-31"  # Valid through end of year
---

# 2026 Holiday Support Schedule
```

Status:
- Fresh: if today is before or on 2026-12-31
- Stale: if today is after 2026-12-31

## code_changes

Documentation expires when related code files change after the last review.

### Configuration

```yaml
---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:  # Optional, overrides defaults
    - "src/api/**/*.ts"
    - "src/handlers/**/*.ts"
---
```

### Watch Patterns

Uses glob patterns to specify which code files to watch. If not specified in frontmatter, uses `defaults.watch` from `.docrot.yml`.

**Default watch patterns:**
```yaml
watch:
  - "**/*.rb"
  - "**/*.go"
  - "**/*.ts"
  - "**/*.tsx"
```

### How It Works

1. Parse `last_reviewed` date
2. Query git history: `git log --since=<last_reviewed>` for files matching watch patterns
3. If any watched files changed after `last_reviewed`, mark as stale

### Git Integration

Uses optimized batch querying:
1. Builds a `FileChangeIndex` with a single git call
2. Precomputes file→last_modified_date mapping
3. Reuses index for all documents (avoids N git calls)

See [Git Integration](git-integration.md) for details.

### Use Cases

- Technical documentation tied to specific code modules
- API documentation that should stay in sync with implementation
- Architecture docs that describe specific components

### Example

```yaml
---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/api/**/*.go"
    - "cmd/server/**/*.go"
---

# API Server Architecture
```

Status:
- Fresh: if no files matching `internal/api/**/*.go` or `cmd/server/**/*.go` changed after 2026-02-01
- Stale: if any watched files changed after 2026-02-01

## Strategy Selection Guide

| Strategy | Best For | Review Trigger |
|----------|----------|----------------|
| `interval` | Regular periodic reviews | Time-based (every N days) |
| `until_date` | Time-sensitive content | Specific calendar date |
| `code_changes` | Code-coupled documentation | Code modifications |

## Combining Strategies

Each document uses one strategy. For complex needs:

1. Use `code_changes` for implementation docs
2. Use `interval` for process docs
3. Use `until_date` for time-sensitive notices

## Overriding Defaults

Documents can override the default strategy from `.docrot.yml`:

```yaml
# .docrot.yml
defaults:
  strategy: interval
  interval: "180d"
```

```yaml
# docs/api.md - overrides default
---
docrot:
  strategy: code_changes  # Override default
  last_reviewed: "2026-02-01"
  watch:
    - "src/api/**/*.ts"
---
```
