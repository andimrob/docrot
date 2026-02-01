---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "**/*.go"
---

# What is Docrot?

Docrot (short for "documentation rot") is a tool that helps keep your documentation fresh by tracking staleness based on configurable strategies.

## The Problem

Documentation becomes stale over time:
- Code changes, but docs don't get updated
- Time-sensitive content expires
- No systematic way to track what needs review

Traditional solutions:
- Manual periodic audits (time-consuming, inconsistent)
- Git blame (shows last edit, not last review)
- Comments in docs (easily forgotten)

## The Solution

Docrot provides automated documentation freshness tracking:

1. **Explicit Review Tracking**: Mark when docs were last reviewed with `last_reviewed` field
2. **Multiple Strategies**: Choose how staleness is determined (time-based, code-based, date-based)
3. **CI/CD Integration**: Fail builds when docs are stale
4. **Minimal Overhead**: Simple YAML frontmatter, no external dependencies

## How It Works

### 1. Add Frontmatter to Docs

```markdown
---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/api/**/*.go"
---

# API Documentation

Your documentation content...
```

### 2. Check Freshness

```bash
docrot check
```

Output:
```
✓ docs/api.md (fresh)
✗ docs/guide.md (stale: code changed in internal/api/handler.go on 2024-02-01)

Summary: 1 fresh, 1 stale, 0 missing frontmatter
```

### 3. Update Review Date

```bash
docrot review docs/guide.md
```

## Three Freshness Strategies

### interval

Documentation expires after a duration since last review.

```yaml
docrot:
  strategy: interval
  last_reviewed: "2026-02-01"
  interval: "90d"  # 90 days
```

**Use case:** Regular review cycles (e.g., quarterly API docs)

### until_date

Documentation expires on a specific date.

```yaml
docrot:
  strategy: until_date
  last_reviewed: "2026-02-01"
  expires: "2024-12-31"
```

**Use case:** Time-sensitive content (e.g., holiday schedules, feature deprecations)

### code_changes

Documentation expires when related code changes.

```yaml
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/api/**/*.go"
    - "cmd/server/**/*.go"
```

**Use case:** Technical docs tied to specific code modules

## Key Features

### Performance Optimized

- **Parallel Processing**: Check multiple docs concurrently
- **Optimized Git Queries**: Single git call for all code change checks
- **Smart Directory Pruning**: Skip irrelevant directories early

### Flexible Configuration

- **Global Defaults**: Set defaults in `.docrot.yml`
- **Per-Document Overrides**: Customize in frontmatter
- **Pattern Matching**: Use glob patterns for file discovery and code watching

### CI/CD Friendly

- **Exit Codes**: Returns 1 if stale docs found
- **JSON Output**: Machine-readable format for automation
- **Quiet Mode**: Only show stale docs

### Multiple Output Formats

- **Text**: Human-readable with status icons
- **JSON**: Machine-readable for scripts and automation

## Example Workflow

### Initial Setup

1. Install docrot (build from source or download binary)
2. Create configuration: `docrot init`
3. Add frontmatter to docs (manually or with `docrot add-frontmatter`)

### Regular Use

1. **During Development**: Update docs and code together
2. **Before Merging**: Run `docrot check` in CI
3. **After Review**: Run `docrot review <file>` to mark as reviewed
4. **Failed Check**: Review and update stale docs

### CI Integration

```yaml
# .github/workflows/docs.yml
name: Documentation Check
on: [pull_request]
jobs:
  check-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Check doc freshness
        run: |
          docrot check
```

## Configuration Example

```yaml
# .docrot.yml
patterns:
  - "**/docs/**/*.md"
  - "**/doc/**/*.md"

exclude:
  - "**/node_modules/**"
  - "**/vendor/**"

on_missing_frontmatter: warn  # or "skip" or "fail"

workers: 0  # 0 = use CPU count

defaults:
  strategy: interval
  interval: "180d"  # 6 months
  watch:
    - "**/*.go"
    - "**/*.ts"
    - "**/*.tsx"
```

## Command Reference

| Command | Purpose |
|---------|---------|
| `check` | Check doc freshness (exits 1 if stale) |
| `list` | List all docs and their status |
| `review <file>` | Update last_reviewed date |
| `init` | Create default config file |
| `add-frontmatter` | Add docrot frontmatter to docs missing it |
| `version` | Show version info |

## Benefits

### For Developers
- **Clear ownership**: Know when docs need review
- **Automated reminders**: CI fails when docs are stale
- **Flexible tracking**: Choose strategy per document

### For Teams
- **Consistent process**: Same approach across projects
- **Reduced manual work**: No more manual doc audits
- **Better quality**: Keep docs in sync with code

### For Organizations
- **Compliance**: Track documentation review cycles
- **Metrics**: Measure doc freshness over time (via JSON output)
- **Standards**: Enforce doc maintenance policies

## Philosophy

Docrot is based on these principles:

1. **Explicit over Implicit**: Explicitly mark review dates, don't infer from git
2. **Flexible Strategies**: Different docs have different needs
3. **Low Friction**: Minimal overhead, easy to adopt
4. **Automation Friendly**: Designed for CI/CD integration
5. **Composable**: Works with existing doc workflows

## Getting Started

1. **Install docrot** (see README for installation)
2. **Initialize config**: `docrot init`
3. **Add frontmatter** to a few docs
4. **Run check**: `docrot check`
5. **Integrate with CI** (optional)

See the other documentation files for detailed information on:
- [CLI Commands](cli-commands.md)
- [Configuration](configuration.md)
- [Freshness Strategies](freshness-strategies.md)
- [File Discovery](file-discovery.md)
- [Git Integration](git-integration.md)
- [Parallel Checking](parallel-checking.md)
- [Output Formatting](output-formatting.md)
- [Document Parsing](document-parsing.md)
