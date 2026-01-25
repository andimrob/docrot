package git

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

type Client struct {
	repoRoot string
}

type ChangedFile struct {
	Path string
	Date time.Time
}

// FileChangeIndex holds a precomputed map of file paths to their last change dates.
// This enables efficient batch queries instead of one git call per document.
type FileChangeIndex struct {
	files map[string]time.Time
}

func New(repoRoot string) (*Client, error) {
	// Verify it's a git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return nil, errors.New("not a git repository")
	}

	return &Client{repoRoot: repoRoot}, nil
}

// LastCommitDate returns the date of the last commit that touched the given path
func (c *Client) LastCommitDate(path string) (time.Time, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%aI", "--", path)
	cmd.Dir = c.repoRoot

	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}

	dateStr := strings.TrimSpace(string(out))
	if dateStr == "" {
		return time.Time{}, errors.New("file has no git history")
	}

	return time.Parse(time.RFC3339, dateStr)
}

// FilesChangedSince returns files matching the patterns that were changed after the given date
func (c *Client) FilesChangedSince(since time.Time, patterns []string, relativeTo string) ([]ChangedFile, error) {
	sinceStr := since.Format("2006-01-02T15:04:05")
	cmd := exec.Command("git", "log", "--since="+sinceStr, "--name-only", "--format=%aI", "--diff-filter=ACMR")
	cmd.Dir = c.repoRoot

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	fileChanges := parseGitLogOutput(string(out))

	// Filter by patterns
	var results []ChangedFile
	for path, date := range fileChanges {
		for _, pattern := range patterns {
			if matched, _ := doublestar.Match(pattern, path); matched {
				results = append(results, ChangedFile{Path: path, Date: date})
				break
			}
		}
	}

	return results, nil
}

// parseGitLogOutput parses git log output with --name-only --format=%aI
// Returns a map of file paths to their most recent change date
func parseGitLogOutput(out string) map[string]time.Time {
	files := make(map[string]time.Time)
	lines := strings.Split(out, "\n")
	var currentDate time.Time

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as date
		if t, err := time.Parse(time.RFC3339, line); err == nil {
			currentDate = t
			continue
		}

		// It's a file path - keep the most recent date for each file
		if !currentDate.IsZero() {
			if existing, ok := files[line]; !ok || currentDate.After(existing) {
				files[line] = currentDate
			}
		}
	}

	return files
}

// FindRepoRoot finds the root of the git repository starting from the given path
func FindRepoRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = absPath

	out, err := cmd.Output()
	if err != nil {
		return "", errors.New("not a git repository")
	}

	return strings.TrimSpace(string(out)), nil
}

// BuildFileChangeIndex builds an index of all file changes since the given date.
// Pass time.Time{} (zero time) to get all files ever committed.
// This makes ONE git call and enables efficient queries for multiple documents.
func (c *Client) BuildFileChangeIndex(since time.Time) (*FileChangeIndex, error) {
	var cmd *exec.Cmd

	if since.IsZero() {
		cmd = exec.Command("git", "log", "--name-only", "--format=%aI", "--diff-filter=ACMR")
	} else {
		sinceStr := since.Format("2006-01-02T15:04:05")
		cmd = exec.Command("git", "log", "--since="+sinceStr, "--name-only", "--format=%aI", "--diff-filter=ACMR")
	}
	cmd.Dir = c.repoRoot

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return &FileChangeIndex{files: parseGitLogOutput(string(out))}, nil
}

// HasChangedSince returns true if any file matching the patterns changed after the given date
func (idx *FileChangeIndex) HasChangedSince(since time.Time, patterns []string) bool {
	for path, date := range idx.files {
		if !date.After(since) {
			continue
		}

		for _, pattern := range patterns {
			if matched, _ := doublestar.Match(pattern, path); matched {
				return true
			}
		}
	}
	return false
}

// GetChangedFiles returns all files matching patterns that changed after the given date
func (idx *FileChangeIndex) GetChangedFiles(since time.Time, patterns []string) []ChangedFile {
	var results []ChangedFile

	for path, date := range idx.files {
		if !date.After(since) {
			continue
		}

		for _, pattern := range patterns {
			if matched, _ := doublestar.Match(pattern, path); matched {
				results = append(results, ChangedFile{Path: path, Date: date})
				break
			}
		}
	}

	return results
}

// FileCount returns the number of files in the index
func (idx *FileChangeIndex) FileCount() int {
	return len(idx.files)
}
