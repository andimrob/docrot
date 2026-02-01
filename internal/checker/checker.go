package checker

import (
	"runtime"
	"sync"
	"time"

	"github.com/andimrob/docrot/internal/document"
	"github.com/andimrob/docrot/internal/freshness"
	"github.com/andimrob/docrot/internal/git"
)

// parsedDoc holds a parsed document and any parse error
type parsedDoc struct {
	path string
	doc  *document.Document
	err  error
}

// Run processes files in parallel and returns freshness results.
// It optimizes by building a FileChangeIndex once instead of making
// individual git calls per file for code change detection.
func Run(paths []string, gitClient *git.Client, workers int) []freshness.Result {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	// Phase 1: Parse all documents in parallel
	parsedDocs := parseAllDocs(paths, workers)

	// Phase 2: Build FileChangeIndex for all docs (all strategies check for code changes)
	var index *git.FileChangeIndex
	var repoRoot string
	if gitClient != nil {
		repoRoot = gitClient.RepoRoot()
		oldestDate := findOldestLastReviewedDate(parsedDocs)
		if !oldestDate.IsZero() {
			// Build index starting from oldest date to minimize git output
			var err error
			index, err = gitClient.BuildFileChangeIndex(oldestDate)
			if err != nil {
				// Fall back to no index (individual git calls)
				index = nil
			}
		}
	}

	// Phase 3: Check all documents in parallel using the index
	return checkAllDocs(parsedDocs, gitClient, repoRoot, index, workers)
}

// parseAllDocs parses all documents in parallel
func parseAllDocs(paths []string, workers int) []parsedDoc {
	jobs := make(chan string, len(paths))
	results := make(chan parsedDoc, len(paths))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				doc, err := document.Parse(path)
				results <- parsedDoc{path: path, doc: doc, err: err}
			}
		}()
	}

	for _, path := range paths {
		jobs <- path
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var parsed []parsedDoc
	for pd := range results {
		parsed = append(parsed, pd)
	}

	return parsed
}

// findOldestLastReviewedDate finds the oldest last_reviewed date among all docs
// with freshness configuration. Returns zero time if none found.
// All strategies check for code changes, so we need the oldest date across all docs.
func findOldestLastReviewedDate(docs []parsedDoc) time.Time {
	var oldest time.Time

	for _, pd := range docs {
		if pd.err != nil || pd.doc == nil || pd.doc.Freshness == nil {
			continue
		}

		date, err := time.Parse("2006-01-02", pd.doc.Freshness.LastReviewed)
		if err != nil {
			continue
		}

		if oldest.IsZero() || date.Before(oldest) {
			oldest = date
		}
	}

	return oldest
}

// checkAllDocs checks all parsed documents in parallel
func checkAllDocs(docs []parsedDoc, gitClient *git.Client, repoRoot string, index *git.FileChangeIndex, workers int) []freshness.Result {
	jobs := make(chan parsedDoc, len(docs))
	results := make(chan freshness.Result, len(docs))

	var wg sync.WaitGroup
	checker := freshness.NewChecker(gitClient, repoRoot)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pd := range jobs {
				var result freshness.Result
				if pd.err != nil {
					result = freshness.Result{
						Path:   pd.path,
						Status: freshness.StatusStale,
						Reason: "Failed to parse: " + pd.err.Error(),
					}
				} else {
					result = checker.CheckWithIndex(pd.doc, index)
				}
				results <- result
			}
		}()
	}

	for _, pd := range docs {
		jobs <- pd
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []freshness.Result
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

// RunWithCallback processes files in parallel and calls callback for each result
// Useful for streaming output
func RunWithCallback(paths []string, gitClient *git.Client, workers int, callback func(freshness.Result)) {
	results := Run(paths, gitClient, workers)
	for _, result := range results {
		callback(result)
	}
}
