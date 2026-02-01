---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "**/*.go"
---

# What is Docrot?

Docrot is a command-line tool that helps you **detect stale documentation** and **keep your docs fresh**. It acts as a documentation freshness checker that can be integrated into your development workflow and CI/CD pipelines.

## The Problem

Documentation often becomes outdated as code evolves. Traditional approaches to keeping docs fresh rely on manual reviews, which are:
- **Time-consuming**: Requires manual tracking of what needs review
- **Error-prone**: Easy to forget which docs relate to which code
- **Inconsistent**: No standard way to track documentation freshness across projects

## The Solution

Docrot solves this by:

1. **Embedding freshness metadata in documentation files** using YAML frontmatter
2. **Automatically tracking documentation staleness** using multiple strategies
3. **Integrating with your git history** to detect when related code changes
4. **Providing clear, actionable feedback** about which docs need attention
5. **Supporting CI/CD integration** to enforce documentation quality gates

## How It Works

### 1. Add Frontmatter to Your Docs

Docrot uses YAML frontmatter in your Markdown files to track freshness. You can choose from three strategies:

**Interval Strategy** - Doc expires after a fixed duration:
```yaml
---
docrot:
  strategy: interval
  last_reviewed: "2026-02-01"
  interval: 90d
---
```

**Until Date Strategy** - Doc expires on a specific date:
```yaml
---
docrot:
  strategy: until_date
  last_reviewed: "2026-02-01"
  expires: "2024-06-01"
---
```

**Code Changes Strategy** - Doc expires when related code files change:
```yaml
---
docrot:
  strategy: code_changes
  last_reviewed: "2026-02-01"
  watch:
    - "internal/**/*.go"
    - "cmd/**/*.go"
---
```

### 2. Check Documentation Freshness

Run `docrot check` to verify all your documentation is up-to-date:

```bash
$ docrot check
✓ docs/setup.md - fresh (expires 2024-04-15)
✗ docs/api.md - stale (expired 2024-01-10)
✗ docs/guide.md - stale (code changed: internal/api/handler.go on 2024-02-01)
```

The tool scans your repository for documentation files, parses their frontmatter, and checks each document's freshness status based on its configured strategy.

### 3. Update Review Dates

When you review and update documentation, use `docrot review` to update the `last_reviewed` date:

```bash
$ docrot review docs/api.md
Updated last_reviewed to 2024-02-15 in docs/api.md
```

This resets the freshness timer for that document.

## Key Features

### Multiple Freshness Strategies

- **Interval**: Perfect for general documentation that needs periodic review (e.g., every 90 days)
- **Until Date**: Great for time-sensitive docs (e.g., deprecation notices, roadmap items)
- **Code Changes**: Ideal for technical docs that should be reviewed when related code changes

### Parallel Processing

Docrot uses parallel workers to efficiently check large documentation sets. It automatically uses all available CPU cores, or you can specify a custom worker count:

```bash
$ docrot check --workers 4
```

### Git Integration

For the `code_changes` strategy, docrot integrates with your git repository to:
- Track when files matching watch patterns were last modified
- Build an efficient index of file changes for batch processing
- Identify exactly which code changes made a doc stale

### Flexible Configuration

Create a `.docrot.yml` configuration file to customize:
- File patterns for finding documentation
- Exclusion patterns for directories like `node_modules` or `vendor`
- Default freshness settings for new docs
- Behavior when frontmatter is missing
- Number of parallel workers

### CI/CD Integration

Docrot exits with code 1 if any docs are stale, making it perfect for CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Check documentation freshness
  run: docrot check
```

### Multiple Output Formats

- **Text format** (default): Human-readable output for interactive use
- **JSON format**: Machine-readable output for scripting and automation

```bash
$ docrot check --format json
```

## Architecture

Docrot is built with several key components:

- **Scanner**: Discovers documentation files using configurable glob patterns with optimized directory traversal
- **Document Parser**: Extracts and validates YAML frontmatter from Markdown files
- **Freshness Checker**: Evaluates document staleness using the appropriate strategy
- **Git Client**: Interfaces with git to track file changes for the `code_changes` strategy
- **Parallel Checker**: Coordinates worker pools for efficient batch processing
- **Output Formatters**: Renders results in text or JSON format

## Use Cases

### Development Workflow

Add docrot checks to pre-commit hooks to catch stale docs before they're committed:

```bash
#!/bin/sh
docrot check --quiet
```

### Continuous Integration

Enforce documentation freshness as part of your CI pipeline to prevent merging PRs with stale docs.

### Periodic Audits

Run `docrot list` to see the freshness status of all documentation:

```bash
$ docrot list
docs/setup.md         fresh    interval     2024-01-15  2024-04-15
docs/api.md          stale    code_changes  2023-12-01  -
docs/guide.md        fresh    until_date    2024-01-01  2024-12-31
```

### Documentation Initialization

Use `docrot init` to add freshness frontmatter to docs that don't have it yet:

```bash
$ docrot init --strategy interval --interval 90d --dry-run
Would add frontmatter to: docs/new-feature.md
Would add frontmatter to: docs/tutorial.md
```

## Benefits

- **Automated tracking**: No more manual spreadsheets or calendar reminders
- **Code-aware**: Automatically detect when docs are outdated due to code changes
- **Flexible**: Choose the freshness strategy that fits each document's needs
- **Fast**: Parallel processing and optimized scanning for large codebases
- **Actionable**: Clear feedback on what needs attention and why
- **Integrable**: Works seamlessly with existing tools and workflows
- **Configurable**: Adapt to your project's specific needs and conventions
