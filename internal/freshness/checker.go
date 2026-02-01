package freshness

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andimrob/docrot/internal/document"
	"github.com/andimrob/docrot/internal/git"
)

type Checker struct {
	git      *git.Client
	repoRoot string
}

func NewChecker(gitClient *git.Client, repoRoot string) *Checker {
	return &Checker{git: gitClient, repoRoot: repoRoot}
}

// ComputeDefaultPatterns computes smart default watch/ignore patterns based on document location.
// Documents in a "docs" or "doc" directory will watch the parent subsystem and ignore the docs dir.
func ComputeDefaultPatterns(docPath, repoRoot string) (watch []string, ignore []string) {
	// Get path relative to repo root
	relPath, err := filepath.Rel(repoRoot, docPath)
	if err != nil {
		// Fallback to watching everything
		return []string{"**/*"}, nil
	}

	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	// Find docs directory in path (docs, doc, documentation)
	parts := strings.Split(filepath.Dir(relPath), "/")
	docsIdx := -1
	for i, part := range parts {
		lower := strings.ToLower(part)
		if lower == "docs" || lower == "doc" || lower == "documentation" {
			docsIdx = i
			break
		}
	}

	if docsIdx == -1 {
		// No docs directory found, watch parent directory
		parentDir := filepath.Dir(relPath)
		if parentDir == "." {
			return []string{"**/*"}, nil
		}
		return []string{parentDir + "/**/*"}, nil
	}

	// Build paths
	var subsystemPath string
	if docsIdx == 0 {
		// Docs at root level (e.g., docs/readme.md)
		subsystemPath = ""
	} else {
		// Docs in subsystem (e.g., subsystem/docs/readme.md)
		subsystemPath = strings.Join(parts[:docsIdx], "/")
	}

	docsPath := strings.Join(parts[:docsIdx+1], "/")

	if subsystemPath == "" {
		return []string{"**/*"}, []string{docsPath + "/**"}
	}
	return []string{subsystemPath + "/**/*"}, []string{docsPath + "/**"}
}

// getWatchIgnorePatterns returns the watch and ignore patterns for a document,
// using explicit patterns from frontmatter if provided, otherwise smart defaults
func (c *Checker) getWatchIgnorePatterns(doc *document.Document) (watch []string, ignore []string) {
	watch = doc.Freshness.Watch
	ignore = doc.Freshness.Ignore

	// If neither watch nor ignore is specified, use smart defaults
	if len(watch) == 0 && len(ignore) == 0 && c.repoRoot != "" {
		return ComputeDefaultPatterns(doc.Path, c.repoRoot)
	}

	// If watch is specified but ignore is not, use empty ignore
	// If ignore is specified but watch is not, use watch all
	if len(watch) == 0 {
		watch = []string{"**/*"}
	}

	return watch, ignore
}

func (c *Checker) Check(doc *document.Document) Result {
	return c.CheckWithIndex(doc, nil)
}

// CheckWithIndex checks a document using an optional precomputed FileChangeIndex.
// All strategies check for code changes in addition to their primary logic.
// Using an index avoids individual git calls for code change detection.
func (c *Checker) CheckWithIndex(doc *document.Document, index *git.FileChangeIndex) Result {
	result := Result{
		Path: doc.Path,
	}

	if doc.Freshness == nil {
		result.Status = StatusMissingFrontmatter
		result.Reason = "No freshness frontmatter found"
		return result
	}

	result.Strategy = doc.Freshness.Strategy
	result.LastReviewed = doc.Freshness.LastReviewed

	switch doc.Freshness.Strategy {
	case "interval":
		return c.checkInterval(doc, result)
	case "until_date":
		return c.checkUntilDate(doc, result)
	case "code_changes":
		return c.checkCodeChangesOnly(doc, result, index)
	default:
		result.Status = StatusStale
		result.Reason = fmt.Sprintf("Unknown strategy: %s", doc.Freshness.Strategy)
		return result
	}
}

// checkCodeChanges checks if any watched files have changed since last_reviewed.
// Returns the changed files, or nil if no code change check should be performed.
func (c *Checker) checkCodeChanges(doc *document.Document, lastReviewed time.Time, index *git.FileChangeIndex) ([]git.ChangedFile, error) {
	watchPatterns, ignorePatterns := c.getWatchIgnorePatterns(doc)

	// If no patterns are configured and no smart defaults available, skip code change check
	if len(watchPatterns) == 0 {
		return nil, nil
	}

	// Use index if available, otherwise make direct git call
	if index != nil {
		return index.GetChangedFiles(lastReviewed, watchPatterns, ignorePatterns), nil
	}

	if c.git == nil {
		return nil, nil // No git client, skip code change check
	}

	return c.git.FilesChangedSince(lastReviewed, watchPatterns, ignorePatterns, "")
}

func (c *Checker) checkInterval(doc *document.Document, result Result) Result {
	lastReviewed, err := time.Parse("2006-01-02", doc.Freshness.LastReviewed)
	if err != nil {
		result.Status = StatusStale
		result.Reason = fmt.Sprintf("Invalid last_reviewed date: %v", err)
		return result
	}

	interval, err := parseInterval(doc.Freshness.Interval)
	if err != nil {
		result.Status = StatusStale
		result.Reason = fmt.Sprintf("Invalid interval: %v", err)
		return result
	}

	expiresAt := lastReviewed.Add(interval)
	now := time.Now()

	if now.After(expiresAt) {
		result.Status = StatusStale
		result.StaleSince = expiresAt.Format("2006-01-02")
		result.Reason = fmt.Sprintf("Interval of %s exceeded (expired %s)",
			doc.Freshness.Interval, expiresAt.Format("2006-01-02"))
	} else {
		result.Status = StatusFresh
		result.Expires = expiresAt.Format("2006-01-02")
	}

	return result
}

func (c *Checker) checkUntilDate(doc *document.Document, result Result) Result {
	expires, err := time.Parse("2006-01-02", doc.Freshness.Expires)
	if err != nil {
		result.Status = StatusStale
		result.Reason = fmt.Sprintf("Invalid expires date: %v", err)
		return result
	}

	result.Expires = doc.Freshness.Expires
	now := time.Now()

	if now.After(expires) {
		result.Status = StatusStale
		result.StaleSince = doc.Freshness.Expires
		result.Reason = fmt.Sprintf("Expired on %s", doc.Freshness.Expires)
	} else {
		result.Status = StatusFresh
	}

	return result
}

// checkCodeChangesOnly implements the code_changes strategy (only checks for code changes)
func (c *Checker) checkCodeChangesOnly(doc *document.Document, result Result, index *git.FileChangeIndex) Result {
	lastReviewed, err := time.Parse("2006-01-02", doc.Freshness.LastReviewed)
	if err != nil {
		result.Status = StatusStale
		result.Reason = fmt.Sprintf("Invalid last_reviewed date: %v", err)
		return result
	}

	if c.git == nil && index == nil {
		result.Status = StatusStale
		result.Reason = "Git client not available for code_changes strategy"
		return result
	}

	// Use start of next day for comparison, so commits on last_reviewed date are not flagged.
	// This handles timezone differences (e.g., a commit at 7pm EST on Jan 31 is Feb 1 00:xx UTC).
	sinceDate := lastReviewed.AddDate(0, 0, 1)
	changed, err := c.checkCodeChanges(doc, sinceDate, index)
	if err != nil {
		result.Status = StatusStale
		result.Reason = fmt.Sprintf("Error checking git history: %v", err)
		return result
	}

	if len(changed) > 0 {
		result.Status = StatusStale
		result.StaleSince = changed[0].Date.Format("2006-01-02")
		result.Reason = fmt.Sprintf("Code changed: %s (%s)",
			changed[0].Path, changed[0].Date.Format("2006-01-02"))
		// Collect all changed file paths
		result.ChangedFiles = make([]string, len(changed))
		for i, f := range changed {
			result.ChangedFiles[i] = f.Path
		}
	} else {
		result.Status = StatusFresh
	}

	return result
}

// parseInterval converts interval strings like "90d", "12w", "3m", "1y" to duration
func parseInterval(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("interval too short: %s", s)
	}

	numStr := s[:len(s)-1]
	unit := strings.ToLower(s[len(s)-1:])

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid number in interval: %s", s)
	}

	var days int
	switch unit {
	case "d":
		days = num
	case "w":
		days = num * 7
	case "m":
		days = num * 30 // Approximate
	case "y":
		days = num * 365
	default:
		return 0, fmt.Errorf("unknown interval unit: %s", unit)
	}

	return time.Duration(days) * 24 * time.Hour, nil
}
