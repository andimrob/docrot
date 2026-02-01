---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/checker/**/*.go"
---

# Parallel Checking

Docrot processes documents in parallel for maximum performance. This document explains how the parallel checking system works and how to configure it.

## Overview

The parallel checker in `internal/checker/checker.go` processes multiple documentation files concurrently, making it fast even with large repositories. By default, it uses all available CPU cores, but you can configure the number of workers.

## Architecture

The parallel processing happens in three phases:

### Phase 1: Document Parsing

All documents are parsed in parallel using a worker pool. Each worker:
- Reads a document file from the job queue
- Parses the frontmatter to extract docrot configuration
- Returns the parsed document or any parse errors

This phase is purely CPU-bound and benefits greatly from parallelization.

### Phase 2: FileChangeIndex Building (Optimization)

For documents using the `code_changes` strategy, docrot optimizes git operations by building a `FileChangeIndex`. This optimization:

- Identifies the oldest `last_reviewed` date among all code_changes documents
- Makes **one single git log call** starting from that date
- Builds an in-memory index of all file changes
- Eliminates the need for individual git calls per document

Without this optimization, checking 100 documents could mean 100 separate git calls. With the index, it's just one git call regardless of document count.

### Phase 3: Freshness Checking

All parsed documents are checked in parallel using the FileChangeIndex. Each worker:
- Takes a parsed document from the job queue
- Checks its freshness status using the appropriate strategy
- For `code_changes` strategy, queries the FileChangeIndex instead of calling git directly
- Returns the freshness result (fresh, stale, or error)

## Worker Configuration

Control the number of parallel workers with the `--workers` flag:

```bash
# Use 4 workers
docrot check --workers 4

# Use default (CPU count)
docrot check

# Use maximum parallelism (CPU count)
docrot check --workers 0
```

You can also set workers in `.docrot.yml`:

```yaml
workers: 8
```

## Performance Considerations

**When Parallelism Helps:**
- Large number of documents (10+)
- Documents using `code_changes` strategy with many watch patterns
- Mixed workload of parsing and git operations

**When to Use Fewer Workers:**
- Very small repositories (1-5 documents) where overhead exceeds benefit
- Systems with limited CPU or memory
- When running in CI with constrained resources

**Optimal Settings:**
- Default (CPU count) works well for most cases
- Increase workers for CPU-heavy parsing workloads
- Decrease workers if you encounter resource constraints

## Implementation Details

The checker uses Go's channels and goroutines for safe parallel execution:

```go
// Worker pool pattern
jobs := make(chan string, len(paths))
results := make(chan parsedDoc, len(paths))

var wg sync.WaitGroup
for i := 0; i < workers; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for path := range jobs {
            // Process document...
        }
    }()
}
```

All document checking is thread-safe. Multiple goroutines can:
- Parse different documents simultaneously
- Query the shared FileChangeIndex concurrently (read-only)
- Write results to separate channels without conflicts

## See Also

- [Git Integration](git-integration.md) - Details on the FileChangeIndex optimization
- [Freshness Strategies](freshness-strategies.md) - Different strategies and their performance characteristics
