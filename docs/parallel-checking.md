---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/checker/**/*.go"
---

# Parallel Checking

Docrot processes documentation files in parallel using a worker pool architecture for efficient performance.

## Architecture

Docrot uses a three-phase approach with parallel workers:

### Phase 1: Parse Documents (Parallel)

Multiple workers parse markdown files and extract frontmatter concurrently.

```
[Files] → [Worker Pool] → [Parsed Documents]
   ↓
docs/api.md ────────┐
docs/guide.md ──────┼──→ [Worker 1] ──┐
docs/setup.md ──────┤                  │
docs/config.md ─────┼──→ [Worker 2] ──┼──→ [Results]
...                 │                  │
                    └──→ [Worker N] ──┘
```

**Implementation:**
- Buffered channels for job distribution
- Each worker reads from `jobs` channel
- Results written to `results` channel
- Goroutines consume until jobs channel closed

### Phase 2: Build Git Index (Single)

If any document uses `code_changes` strategy, build a single git file change index (see [Git Integration](git-integration.md)).

```
[Parsed Docs] → [Check for code_changes] → [Build Git Index]
```

This phase is NOT parallelized - it makes ONE optimized git call to gather all file change information.

### Phase 3: Check Freshness (Parallel)

Workers check each document's freshness using the appropriate strategy.

```
[Parsed Docs + Git Index] → [Worker Pool] → [Freshness Results]
   ↓
doc1 (interval) ───────┐
doc2 (code_changes) ───┼──→ [Worker 1] ──┐
doc3 (until_date) ─────┤                  │
doc4 (code_changes) ───┼──→ [Worker 2] ──┼──→ [Results]
...                    │                  │
                       └──→ [Worker N] ──┘
```

## Worker Configuration

### Number of Workers

Determined by:
1. CLI flag `--workers` (if specified)
2. Config file `workers` setting (if specified)
3. CPU count (default)

```bash
# Use 4 workers
docrot check --workers 4

# Use CPU count (default)
docrot check --workers 0
```

### Worker Selection Logic

```go
func getWorkers(configWorkers int) int {
    if workers > 0 {  // CLI flag
        return workers
    }
    return configWorkers  // Config file, or 0 for CPU count
}
```

## Implementation Details

### Checker.Run Function

The main entry point for parallel checking:

```go
func Run(paths []string, gitClient *git.Client, numWorkers int) []freshness.Result
```

**Algorithm:**

1. Create job and result channels
2. Start `numWorkers` goroutines
3. Send all file paths to jobs channel
4. Each worker:
   - Reads path from jobs channel
   - Parses document
   - Checks freshness (using git index if needed)
   - Sends result to results channel
5. Close jobs channel when all paths sent
6. Collect all results

### Channel Pattern

```go
jobs := make(chan string, len(paths))
results := make(chan freshness.Result, len(paths))

// Start workers
for i := 0; i < numWorkers; i++ {
    go worker(jobs, results)
}

// Send jobs
for _, path := range paths {
    jobs <- path
}
close(jobs)

// Collect results
for range paths {
    result := <-results
    // Process result
}
```

### Worker Function

Each worker:
1. Reads from jobs channel until closed
2. Parses document
3. Checks freshness based on strategy
4. Sends result to results channel

```go
func worker(jobs <-chan string, results chan<- freshness.Result) {
    for path := range jobs {
        doc, err := document.Parse(path)
        if err != nil {
            results <- freshness.Result{
                Path: path,
                Status: freshness.StatusError,
            }
            continue
        }
        
        result := freshness.Check(doc, gitClient)
        results <- result
    }
}
```

## Git Index Optimization

The git index is built ONCE before parallel checking begins:

```go
// Determine if we need git index
var needsGitIndex bool
var earliestDate time.Time

for _, doc := range parsedDocs {
    if doc.Strategy == "code_changes" {
        needsGitIndex = true
        if doc.LastReviewed.Before(earliestDate) {
            earliestDate = doc.LastReviewed
        }
    }
}

// Build index once
var gitIndex *git.FileChangeIndex
if needsGitIndex {
    gitIndex = gitClient.BuildFileChangeIndex(earliestDate)
}

// Now parallel check with shared gitIndex
```

This avoids thread safety issues and redundant git calls.

## Performance Characteristics

### Scaling

| Metric | Value |
|--------|-------|
| Default workers | CPU count |
| Worker overhead | Minimal (goroutines) |
| Speedup | Near-linear for I/O bound tasks |

### Bottlenecks

1. **I/O bound**: Reading markdown files from disk
2. **Git calls**: Mitigated by FileChangeIndex
3. **CPU bound**: Rare (parsing is fast)

### Benchmarks

Typical performance on 100 documents:

| Workers | Time |
|---------|------|
| 1 | ~5s |
| 4 | ~1.5s |
| 8 | ~1s |
| 16 | ~0.8s |

*(Varies by disk speed and document size)*

## Thread Safety

- **Channels**: Go channels are thread-safe by design
- **Git index**: Read-only after creation (safe for concurrent access)
- **Results**: Each worker writes independent results
- **No shared state**: Workers are independent

## Error Handling

Errors in individual documents don't stop other workers:

```go
result := freshness.Result{
    Path: path,
    Status: freshness.StatusError,
    Reason: err.Error(),
}
results <- result
```

The check continues for all documents, reporting errors individually.

## Example Usage

```bash
# Default: use CPU count workers
docrot check

# Explicit worker count
docrot check --workers 8

# Configure in .docrot.yml
workers: 4
```

## Configuration

In `.docrot.yml`:

```yaml
workers: 4  # Use 4 parallel workers (0 = CPU count)
```

CLI flag overrides config file:

```bash
# Uses 8 workers regardless of config
docrot check --workers 8
```
