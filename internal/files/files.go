package files

import (
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

// ListFiles returns all files under root that match any of the watch patterns
// and don't match any of the ignore patterns. Returned paths are relative to root.
func ListFiles(root string, watchPatterns []string, ignorePatterns []string) ([]string, error) {
	var results []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		if info.IsDir() {
			// Prune excluded directories early to avoid walking their contents
			if relPath != "." && matchesAny(relPath, ignorePatterns) {
				return filepath.SkipDir
			}
			return nil
		}

		if !matchesAny(relPath, watchPatterns) {
			return nil
		}
		if matchesAny(relPath, ignorePatterns) {
			return nil
		}

		results = append(results, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// matchesAny returns true if path matches any of the patterns.
// Pattern errors (malformed patterns) are treated as non-matches.
func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := doublestar.Match(pattern, path)
		if err != nil {
			// Malformed pattern - skip it (treated as non-match)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}
