---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/document/**/*.go"
---

# Document Parsing

This document describes how docrot parses markdown files and extracts frontmatter configuration.

## Overview

The `document` package provides functionality to parse markdown files and extract docrot-specific frontmatter. The main entry point is the `Parse()` function which returns a `Document` struct containing the file path, content, and freshness configuration.

## Document Structure

A parsed document is represented by the following struct:

```go
type Document struct {
    Path      string      // File path to the markdown document
    Freshness *Freshness  // Parsed frontmatter configuration (nil if not present)
    Content   string      // Document content (excluding frontmatter)
}
```

## Frontmatter Format

Docrot uses YAML frontmatter delimited by `---` markers at the beginning of markdown files. The frontmatter must:

- Start with a `---` line at the beginning of the file
- End with another `---` line
- Contain a `docrot:` key with nested configuration

Example:
```yaml
---
docrot:
  last_reviewed: "2026-02-01"
  strategy: interval
  interval: 90d
---
```

## Parsing Process

The `Parse()` function performs the following steps:

1. **Open the file**: Opens the specified file path for reading
2. **Scan line by line**: Uses `bufio.Scanner` to read the file line by line
3. **Extract frontmatter**: 
   - Detects the opening `---` delimiter
   - Collects all lines between the delimiters
   - Detects the closing `---` delimiter
   - All subsequent lines become the document content
4. **Parse YAML**: If frontmatter is found, it's parsed as YAML into a `frontmatterWrapper` struct
5. **Return Document**: Returns a `Document` struct with the parsed data

## Freshness Configuration

The frontmatter can contain a `docrot` section with the following fields:

```go
type Freshness struct {
    LastReviewed string   `yaml:"last_reviewed"`  // Last review date (YYYY-MM-DD)
    Strategy     string   `yaml:"strategy"`       // Freshness strategy: interval, until_date, or code_changes
    Interval     string   `yaml:"interval,omitempty"`  // For interval strategy (e.g., "90d", "3m", "1y")
    Expires      string   `yaml:"expires,omitempty"`   // For until_date strategy (YYYY-MM-DD)
    Watch        []string `yaml:"watch,omitempty"`     // For code_changes strategy (file patterns)
}
```

## Error Handling

The parser returns an error if:
- The file cannot be opened or read
- The YAML frontmatter is malformed and cannot be parsed

Documents without frontmatter are valid and return a `Document` with `Freshness` set to `nil`.

## Usage Example

```go
import "github.com/andimrob/docrot/internal/document"

doc, err := document.Parse("docs/api.md")
if err != nil {
    // Handle error
}

if doc.Freshness != nil {
    // Process freshness configuration
    fmt.Printf("Strategy: %s\n", doc.Freshness.Strategy)
    fmt.Printf("Last reviewed: %s\n", doc.Freshness.LastReviewed)
}

// Access document content
fmt.Println(doc.Content)
```
