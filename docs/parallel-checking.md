---
docrot:
  strategy: code_changes
  last_reviewed: "2026-01-31"
  watch:
    - "internal/checker/**/*.go"
---

# Parallel Checking

docrot processes documentation files in parallel for performance on large repositories.

## How It Works

The checker runs in three phases:

1. **Parse Phase**: All documents are parsed in parallel using a worker pool
2. **Index Phase**: A single git call builds a `FileChangeIndex` containing all file changes since the oldest `last_reviewed` date across all documents
3. **Check Phase**: All documents are checked in parallel against the precomputed index

## Worker Configuration

The number of parallel workers can be configured:

```yaml
# .docrot.yml
workers: 4  # Use 4 workers
```

Or via command line:

```bash
docrot check --workers 4
```

Setting workers to `0` (the default) uses the number of CPU cores.

## Performance Optimization

The `FileChangeIndex` is the key optimization. Instead of making one git call per document to check for code changes, docrot:

1. Finds the oldest `last_reviewed` date across all documents
2. Makes a single `git log` call to build an index of all file changes since that date
3. Queries the index in memory for each document

This reduces git operations from O(n) to O(1), making docrot fast even on repositories with hundreds of documentation files.
