---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/scanner/**/*.go"
---

# File Discovery

Docrot uses glob patterns to discover documentation files in your project. The scanner is optimized for performance with large codebases.

## How It Works

1. **Pattern Matching**: Uses [doublestar](https://github.com/bmatcuk/doublestar) library for glob pattern matching (supports `**` for recursive matching)
2. **Directory Pruning**: Extracts target directory names from patterns (e.g., "docs" from `**/docs/**/*.md`) and skips unrelated directories
3. **Exclusion**: Applies exclude patterns to skip files and directories
4. **Performance**: Only processes `.md` files in relevant directories

## Pattern Syntax

Glob patterns support:
- `*`: Matches any characters within a path segment
- `**`: Matches any characters across multiple path segments (recursive)
- `?`: Matches a single character
- `{a,b}`: Matches either `a` or `b`

### Examples

```yaml
# Match all markdown files in docs directories
patterns:
  - "**/docs/**/*.md"

# Match multiple specific directories
patterns:
  - "docs/**/*.md"
  - "api/**/*.md"
  - "guides/**/*.md"

# Match all markdown files anywhere
patterns:
  - "**/*.md"
```

## Exclude Patterns

Exclude patterns use the same glob syntax to skip files or directories.

### Default Exclusions

Docrot automatically excludes:
```yaml
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
```

Plus these hard-coded directories:
- `.git`, `.svn` (version control)
- `__pycache__`, `.cache` (caches)
- `dist`, `build` (build artifacts)
- `tmp`, `log`, `logs` (temporary files)
- Hidden directories (starting with `.`)

### Custom Exclusions

```yaml
exclude:
  - "**/node_modules/**"
  - "**/target/**"          # Rust build directory
  - "**/build/**"           # Build artifacts
  - "**/*.generated.md"     # Generated documentation
```

## Scanner Optimization

The scanner is optimized for performance with large repositories:

1. **Early Directory Pruning**: Skips excluded directories immediately using `filepath.SkipDir`, avoiding unnecessary file system traversal
2. **Target Directory Extraction**: Identifies target directories from patterns (e.g., "docs", "doc") and filters early
3. **File Extension Check**: Only processes `.md` files
4. **Hidden Directory Skip**: Automatically skips hidden directories (except root)

### Example

For pattern `**/docs/**/*.md`:
1. Scanner extracts "docs" as a target directory
2. During traversal, skips any directory not containing "docs" in its path
3. Only checks `.md` files within paths containing "docs"
4. Applies exclude patterns to final matches

## Configuration

File discovery is configured in `.docrot.yml`:

```yaml
patterns:
  - "**/docs/**/*.md"
  - "**/doc/**/*.md"

exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
  - "**/build/**"
```

## Usage

The scanner runs automatically with `check` and `list` commands:

```bash
# Scan current directory
docrot check

# Scan specific directory
docrot check ./docs

# Uses patterns and exclude from .docrot.yml
```
