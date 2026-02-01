---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/output/**/*.go"
---

# Output Formatting

Docrot supports two output formats: human-readable text and machine-readable JSON.

## Format Selection

Specify format with the `--format` flag:

```bash
# Text format (default)
docrot check

# JSON format
docrot check --format json
```

## Text Format

The default text format is designed for human readability and CI/CD integration.

### Standard Output

```
✓ docs/api.md (fresh)
✗ docs/guide.md (stale: code changed in internal/api/handler.go on 2024-02-01)
? docs/setup.md (missing frontmatter)
✓ docs/config.md (fresh)

Summary: 2 fresh, 1 stale, 1 missing frontmatter
```

### Status Icons

- `✓`: Fresh (document is up to date)
- `✗`: Stale (document needs review)
- `?`: Missing frontmatter (no docrot configuration)

### Stale Reasons

When a document is stale, the text format shows why:

**Interval strategy:**
```
✗ docs/api.md (stale: last reviewed 2023-06-01, expired 2024-01-01)
```

**Until date strategy:**
```
✗ docs/guide.md (stale: expired 2024-01-15)
```

**Code changes strategy:**
```
✗ docs/architecture.md (stale: code changed in src/api.go on 2024-02-01)
```

### Quiet Mode

Use `--quiet` flag to only show stale documents:

```bash
docrot check --quiet
```

Output:
```
✗ docs/guide.md (stale: code changed in internal/api/handler.go on 2024-02-01)

Summary: 2 fresh, 1 stale, 1 missing frontmatter
```

Fresh documents are hidden, but the summary still shows total counts.

### Summary Line

Always displayed at the end:
```
Summary: <fresh_count> fresh, <stale_count> stale, <missing_count> missing frontmatter
```

## JSON Format

Machine-readable format for automation and CI/CD integration.

### Structure

```json
{
  "summary": {
    "total": 4,
    "fresh": 2,
    "stale": 1,
    "missing_frontmatter": 1
  },
  "docs": [
    {
      "path": "docs/api.md",
      "status": "fresh",
      "strategy": "interval",
      "last_reviewed": "2024-01-15",
      "reason": ""
    },
    {
      "path": "docs/guide.md",
      "status": "stale",
      "strategy": "code_changes",
      "last_reviewed": "2023-12-01",
      "reason": "code changed in internal/api/handler.go on 2024-02-01"
    },
    {
      "path": "docs/setup.md",
      "status": "missing_frontmatter",
      "strategy": "",
      "last_reviewed": "",
      "reason": "no docrot frontmatter found"
    }
  ]
}
```

### Field Descriptions

**Summary object:**
- `total`: Total number of documents checked
- `fresh`: Number of fresh documents
- `stale`: Number of stale documents
- `missing_frontmatter`: Number of documents without frontmatter

**Document object:**
- `path`: File path relative to root
- `status`: One of: `"fresh"`, `"stale"`, `"missing_frontmatter"`
- `strategy`: Freshness strategy: `"interval"`, `"until_date"`, `"code_changes"`, or empty string
- `last_reviewed`: ISO 8601 date (e.g., `"2024-01-15"`) or empty string
- `reason`: Explanation for stale status or error message

### Status Values

- `"fresh"`: Document is up to date
- `"stale"`: Document needs review
- `"missing_frontmatter"`: No docrot frontmatter found

### Example Usage

Parse JSON output in scripts:

```bash
# Get stale document count
docrot check --format json | jq '.summary.stale'

# List all stale documents
docrot check --format json | jq -r '.docs[] | select(.status=="stale") | .path'

# Get stale docs with reasons
docrot check --format json | jq -r '.docs[] | select(.status=="stale") | "\(.path): \(.reason)"'
```

### CI/CD Integration

JSON format is ideal for CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Check documentation
  run: |
    OUTPUT=$(docrot check --format json)
    STALE_COUNT=$(echo "$OUTPUT" | jq '.summary.stale')
    if [ "$STALE_COUNT" -gt 0 ]; then
      echo "Found $STALE_COUNT stale documents"
      echo "$OUTPUT" | jq -r '.docs[] | select(.status=="stale")'
      exit 1
    fi
```

## Exit Codes

Both formats use the same exit codes:

- `0`: All documents are fresh
- `1`: Stale documentation found or error occurred

Check the exit code to determine if the check passed:

```bash
if docrot check; then
    echo "All docs are fresh!"
else
    echo "Stale docs found or error occurred"
fi
```

## Configuration

Output format is controlled by the `--format` flag (no config file option):

```bash
# Text format (default)
docrot check --format text

# JSON format
docrot check --format json
```

The `--quiet` flag only affects text format:

```bash
# Only stale docs in text format
docrot check --format text --quiet

# Quiet has no effect on JSON format
docrot check --format json --quiet  # Still outputs full JSON
```

## list vs check Output

The `list` command uses the same formatters but:
- Shows all documents (no filter)
- Always exits with code 0 (informational)
- Doesn't have `--quiet` flag

```bash
# List all docs
docrot list

# List in JSON format
docrot list --format json
```
