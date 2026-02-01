---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/document/**/*.go"
---

# Document Parsing

Docrot parses Markdown documents and extracts YAML frontmatter to determine freshness settings.

## Frontmatter Format

Frontmatter is YAML enclosed between `---` delimiters at the start of the file.

### Basic Structure

```markdown
---
docrot:
  strategy: interval
  last_reviewed: "2026-02-01"
  interval: "90d"
---

# Your Document Title

Document content here...
```

## Frontmatter Fields

All docrot configuration is nested under the `docrot` key in the frontmatter.

### Required Fields

**last_reviewed** (required for all strategies)
- Format: `"YYYY-MM-DD"` (ISO 8601 date)
- When the document was last reviewed
- Used by all strategies

**strategy** (required)
- One of: `"interval"`, `"until_date"`, `"code_changes"`
- Determines how freshness is checked
- Can be omitted if defaults exist in `.docrot.yml`

### Strategy-Specific Fields

**interval strategy:**
```yaml
docrot:
  strategy: interval
  last_reviewed: "2026-02-01"
  interval: "90d"  # Required
```
- `interval`: Duration format `<number><unit>` (d/w/m/y)

**until_date strategy:**
```yaml
docrot:
  strategy: until_date
  last_reviewed: "2026-02-01"
  expires: "2024-12-31"  # Required
```
- `expires`: Date when document becomes stale (format: `"YYYY-MM-DD"`)

**code_changes strategy:**
```yaml
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:  # Optional
    - "internal/api/**/*.go"
    - "cmd/server/**/*.go"
```
- `watch`: Array of glob patterns (optional, uses defaults if omitted)

## Parsing Process

### Step 1: Detect Frontmatter

1. Open file
2. Look for opening `---` on first line
3. Read lines until closing `---`
4. If no frontmatter found → `StatusMissingFrontmatter`

### Step 2: Parse YAML

1. Extract content between `---` delimiters
2. Parse as YAML
3. Extract nested `docrot` key
4. If parse error → mark as error

### Step 3: Validate Fields

1. Check for required fields (`last_reviewed`, `strategy`)
2. Validate date formats
3. Validate strategy-specific fields
4. If validation fails → mark as stale with error reason

## Parser Implementation

The parser uses Go's `gopkg.in/yaml.v3` package:

```go
type Frontmatter struct {
    Docrot *Freshness `yaml:"docrot"`
}

type Freshness struct {
    Strategy     string   `yaml:"strategy"`
    LastReviewed string   `yaml:"last_reviewed"`
    Interval     string   `yaml:"interval"`
    Expires      string   `yaml:"expires"`
    Watch        []string `yaml:"watch"`
}
```

## Document Structure

```go
type Document struct {
    Path      string
    Freshness *Freshness
}
```

## Examples

### Minimal (with defaults)

If `.docrot.yml` has defaults:
```yaml
defaults:
  strategy: interval
  interval: "180d"
```

Then document frontmatter can be minimal:
```markdown
---
docrot:
  last_reviewed: "2026-02-01"
---
```

### Complete

```markdown
---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/api/**/*.go"
    - "internal/handlers/**/*.go"
    - "cmd/**/*.go"
---

# API Documentation

This document describes our API architecture.
```

### With Other Frontmatter

Docrot frontmatter can coexist with other metadata:

```markdown
---
title: API Guide
author: Engineering Team
date: 2024-01-15
docrot:
  strategy: interval
  last_reviewed: "2026-02-01"
  interval: "90d"
---

# API Guide
```

Only the `docrot` key is parsed by docrot; other keys are ignored.

## Error Handling

### Missing Frontmatter

If no frontmatter found, behavior depends on config:

```yaml
on_missing_frontmatter: "warn"  # Show as missing (default)
on_missing_frontmatter: "skip"  # Ignore file
on_missing_frontmatter: "fail"  # Treat as stale
```

### Invalid Dates

```markdown
---
docrot:
  last_reviewed: "invalid"
---
```

Result: Marked as stale with reason `"invalid date format"`

### Missing Required Fields

**Missing strategy (without defaults):**
```markdown
---
docrot:
  last_reviewed: "2026-02-01"
---
```
Result: Error if no defaults in config

**Missing interval for interval strategy:**
```markdown
---
docrot:
  strategy: interval
  last_reviewed: "2026-02-01"
---
```
Result: Error if no default interval in config

### Malformed YAML

```markdown
---
docrot:
  strategy: interval
  last_reviewed: "2026-02-01"  # Missing quotes
  interval: 90d              # Missing quotes
```

Result: May parse correctly (YAML is flexible) or error depending on content

## Best Practices

### Always Quote Dates and Intervals

```yaml
# Good
docrot:
  last_reviewed: "2026-02-01"
  interval: "90d"

# Avoid (may work but inconsistent)
docrot:
  last_reviewed: "2026-02-01"
  interval: 90d
```

### Keep Frontmatter at Top

Frontmatter must be at the very beginning of the file:

```markdown
---
docrot:
  last_reviewed: "2026-02-01"
---

# Title starts here
```

Not:
```markdown
# Title

---
docrot:  # This won't be detected
  last_reviewed: "2026-02-01"
---
```

### Use Defaults

Define common settings in `.docrot.yml` to keep frontmatter minimal:

```yaml
# .docrot.yml
defaults:
  strategy: interval
  interval: "180d"
  watch:
    - "**/*.go"
```

```markdown
---
docrot:
  last_reviewed: "2026-02-01"  # Strategy and interval from defaults
---
```

## File Encoding

- UTF-8 encoding (standard for Markdown)
- Unix or Windows line endings supported
- BOM (byte order mark) not recommended but handled

## Performance

- **Parsing**: Fast (YAML parsing is efficient)
- **Caching**: None (files read each time)
- **Parallelization**: Multiple documents parsed concurrently (see [Parallel Checking](parallel-checking.md))

## Validation

The parser validates:
- Date format: `YYYY-MM-DD`
- Interval format: `<number><unit>` where unit is d/w/m/y
- Strategy values: Must be `interval`, `until_date`, or `code_changes`
- Required field presence

Invalid values result in documents marked as stale with descriptive error messages.
