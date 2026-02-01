---
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/scanner/**/*.go"
---

# File Discovery

docrot uses an optimized file scanner to discover documentation files.

## Pattern Configuration

Documentation files are found using glob patterns in `.docrot.yml`:

```yaml
patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"

exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
```

## Scanning Behavior

The scanner walks the directory tree with several optimizations:

### Early Directory Pruning

Certain directories are skipped entirely for performance:
- Hidden directories (starting with `.`)
- `node_modules`, `vendor`, `.git`, `.svn`
- `__pycache__`, `.cache`, `dist`, `build`
- `tmp`, `log`, `logs`

Additionally, any directory matching an `exclude` pattern is pruned.

### Target Directory Extraction

The scanner extracts literal directory names from patterns (e.g., `doc`, `docs` from `**/docs/**/*.md`) and only considers markdown files within those directories. This avoids checking every `.md` file in the repository.

### Pattern Matching

Files must:
1. Have a `.md` extension
2. Be inside a target directory (extracted from patterns)
3. Match at least one pattern
4. Not match any exclude pattern

## Glob Syntax

Patterns use doublestar syntax:
- `**` matches any number of directories
- `*` matches any characters in a single path segment
- `?` matches a single character

Examples:
- `**/docs/**/*.md` - Any `.md` file in any `docs` directory at any depth
- `docs/*.md` - Only `.md` files directly in `docs/` (not subdirectories)
