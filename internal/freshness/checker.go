package freshness

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/andimrob/docrot/internal/document"
	"github.com/andimrob/docrot/internal/git"
)

// DefaultWatchPatterns are the default file patterns to watch for code_changes strategy
var DefaultWatchPatterns = []string{"**/*.rb", "**/*.go", "**/*.ts", "**/*.tsx"}

type Checker struct {
	git *git.Client
}

func NewChecker(gitClient *git.Client) *Checker {
	return &Checker{git: gitClient}
}

func (c *Checker) Check(doc *document.Document) Result {
	return c.CheckWithIndex(doc, nil)
}

// CheckWithIndex checks a document using an optional precomputed FileChangeIndex.
// For code_changes strategy, using an index avoids individual git calls.
// For other strategies, the index is ignored.
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
		if index != nil {
			return c.checkCodeChangesWithIndex(doc, result, index)
		}
		return c.checkCodeChanges(doc, result)
	default:
		result.Status = StatusStale
		result.Reason = fmt.Sprintf("Unknown strategy: %s", doc.Freshness.Strategy)
		return result
	}
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

func (c *Checker) checkCodeChanges(doc *document.Document, result Result) Result {
	if c.git == nil {
		result.Status = StatusStale
		result.Reason = "Git client not available for code_changes strategy"
		return result
	}

	lastReviewed, err := time.Parse("2006-01-02", doc.Freshness.LastReviewed)
	if err != nil {
		result.Status = StatusStale
		result.Reason = fmt.Sprintf("Invalid last_reviewed date: %v", err)
		return result
	}

	watchPatterns := doc.Freshness.Watch
	if len(watchPatterns) == 0 {
		watchPatterns = DefaultWatchPatterns
	}

	changed, err := c.git.FilesChangedSince(lastReviewed, watchPatterns, "")
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

// checkCodeChangesWithIndex uses a precomputed index instead of making a git call
func (c *Checker) checkCodeChangesWithIndex(doc *document.Document, result Result, index *git.FileChangeIndex) Result {
	lastReviewed, err := time.Parse("2006-01-02", doc.Freshness.LastReviewed)
	if err != nil {
		result.Status = StatusStale
		result.Reason = fmt.Sprintf("Invalid last_reviewed date: %v", err)
		return result
	}

	watchPatterns := doc.Freshness.Watch
	if len(watchPatterns) == 0 {
		watchPatterns = DefaultWatchPatterns
	}

	changed := index.GetChangedFiles(lastReviewed, watchPatterns)

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
