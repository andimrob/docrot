package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type Scanner struct {
	root     string
	patterns []string
	exclude  []string
}

func New(root string, patterns []string, exclude []string) *Scanner {
	return &Scanner{
		root:     root,
		patterns: patterns,
		exclude:  exclude,
	}
}

// Scan finds documentation files using an optimized walk that prunes excluded directories early
func (s *Scanner) Scan() ([]string, error) {
	var results []string

	// Extract target directory names from patterns (e.g., "doc", "docs" from "**/doc/**/*.md")
	targetDirs := s.extractTargetDirs()

	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		relPath, _ := filepath.Rel(s.root, path)

		// For directories, check if we should skip entirely
		if d.IsDir() {
			// Skip hidden directories (except root)
			if path != s.root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}

			// Skip excluded directories early - this is the key optimization
			if s.shouldSkipDir(relPath) {
				return filepath.SkipDir
			}

			return nil
		}

		// For files, check if they match our patterns
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		// Check if file is in a target directory
		if !s.isInTargetDir(relPath, targetDirs) {
			return nil
		}

		// Check if file matches any pattern
		if !s.matchesPattern(relPath) {
			return nil
		}

		// Check exclude patterns for the file itself
		if s.isExcluded(relPath) {
			return nil
		}

		results = append(results, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// extractTargetDirs extracts directory names we're looking for from patterns
// e.g., "**/doc/**/*.md" -> "doc"
func (s *Scanner) extractTargetDirs() []string {
	dirs := make(map[string]bool)
	for _, pattern := range s.patterns {
		parts := strings.Split(pattern, "/")
		for _, part := range parts {
			// Look for literal directory names (not wildcards)
			if part != "" && part != "**" && part != "*" && !strings.Contains(part, "*") {
				dirs[part] = true
			}
		}
	}

	result := make([]string, 0, len(dirs))
	for dir := range dirs {
		if !strings.Contains(dir, ".") { // Skip file extensions
			result = append(result, dir)
		}
	}
	return result
}

// isInTargetDir checks if the path is inside one of our target directories.
// If no target directories were extracted from patterns (e.g., patterns like "**/*.md"),
// this returns true to allow all files through.
func (s *Scanner) isInTargetDir(relPath string, targetDirs []string) bool {
	// If no target directories specified, allow all files
	if len(targetDirs) == 0 {
		return true
	}

	parts := strings.Split(relPath, string(os.PathSeparator))
	for _, part := range parts {
		for _, target := range targetDirs {
			if part == target {
				return true
			}
		}
	}
	return false
}

// shouldSkipDir checks if we should skip this directory entirely
func (s *Scanner) shouldSkipDir(relPath string) bool {
	dirName := filepath.Base(relPath)

	// Common directories to always skip for performance
	skipDirs := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".git":         true,
		".svn":         true,
		"__pycache__":  true,
		".cache":       true,
	}

	if skipDirs[dirName] {
		return true
	}

	// Check exclude patterns
	for _, pattern := range s.exclude {
		// For directory patterns like "**/node_modules/**", check if current dir matches
		// Pattern errors (malformed patterns) are treated as non-matches
		if matched, err := doublestar.Match(pattern, relPath); err == nil && matched {
			return true
		}
		if matched, err := doublestar.Match(pattern, relPath+"/"); err == nil && matched {
			return true
		}
	}

	return false
}

// matchesAny returns true if path matches any of the given patterns.
// Pattern errors (malformed patterns) are treated as non-matches.
func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, err := doublestar.Match(pattern, path); err == nil && matched {
			return true
		}
	}
	return false
}

// matchesPattern checks if the file matches any of our search patterns.
func (s *Scanner) matchesPattern(relPath string) bool {
	return matchesAny(relPath, s.patterns)
}

// isExcluded checks if a file path matches any exclude pattern.
func (s *Scanner) isExcluded(relPath string) bool {
	return matchesAny(relPath, s.exclude)
}
