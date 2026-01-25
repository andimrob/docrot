package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCommand_WithStaleDoc(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create a stale doc
	staleDoc := `---
freshness:
  last_reviewed: "2020-01-01"
  strategy: interval
  interval: 30d
---
# Stale Doc
`
	os.WriteFile(filepath.Join(docDir, "stale.md"), []byte(staleDoc), 0644)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Reset flags and run
	configPath = ""
	format = "text"
	quiet = false

	err := runCheck(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should return ErrStaleDocsFound for stale docs
	if err != ErrStaleDocsFound {
		t.Errorf("runCheck() error = %v, want ErrStaleDocsFound", err)
	}

	if !strings.Contains(output, "stale") {
		t.Errorf("Output should contain 'stale', got: %s", output)
	}
}

func TestCheckCommand_WithFreshDoc(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create a fresh doc (using current date)
	freshDoc := `---
freshness:
  last_reviewed: "2026-01-20"
  strategy: interval
  interval: 90d
---
# Fresh Doc
`
	os.WriteFile(filepath.Join(docDir, "fresh.md"), []byte(freshDoc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false

	err := runCheck(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runCheck() error = %v", err)
	}

	if !strings.Contains(output, "fresh") {
		t.Errorf("Output should contain 'fresh', got: %s", output)
	}
}

func TestCheckCommand_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	doc := `---
freshness:
  last_reviewed: "2026-01-20"
  strategy: interval
  interval: 90d
---
# Doc
`
	os.WriteFile(filepath.Join(docDir, "test.md"), []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "json"
	quiet = false

	err := runCheck(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runCheck() error = %v", err)
	}

	if !strings.Contains(output, `"status": "fresh"`) {
		t.Errorf("JSON output should contain status field, got: %s", output)
	}
}

func TestInitCommand_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create doc without frontmatter
	doc := `# My Doc
Some content.
`
	docPath := filepath.Join(docDir, "test.md")
	os.WriteFile(docPath, []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	dryRun = true

	err := runInit(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runInit() error = %v", err)
	}

	if !strings.Contains(output, "Would add frontmatter") {
		t.Errorf("Output should mention dry run, got: %s", output)
	}

	// Verify file wasn't changed
	content, _ := os.ReadFile(docPath)
	if strings.Contains(string(content), "freshness") {
		t.Error("Dry run should not modify files")
	}
}

func TestGetWorkers_FlagOverridesConfig(t *testing.T) {
	// When flag is set, it should override config
	workers = 4 // CLI flag
	configWorkers := 8

	result := getWorkers(configWorkers)
	if result != 4 {
		t.Errorf("getWorkers() = %d, want 4 (flag value)", result)
	}

	// Reset
	workers = 0
}

func TestGetWorkers_ConfigUsedWhenNoFlag(t *testing.T) {
	// When flag is 0, use config value
	workers = 0
	configWorkers := 8

	result := getWorkers(configWorkers)
	if result != 8 {
		t.Errorf("getWorkers() = %d, want 8 (config value)", result)
	}
}

func TestGetWorkers_ZeroMeansDefault(t *testing.T) {
	// When both are 0, return 0 (which means use CPU count at runtime)
	workers = 0
	configWorkers := 0

	result := getWorkers(configWorkers)
	if result != 0 {
		t.Errorf("getWorkers() = %d, want 0 (use CPU count)", result)
	}
}

func TestInitCommand_MergesWithExistingFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create doc with existing frontmatter but no freshness
	doc := `---
title: My Document
author: Jane Doe
tags:
  - api
  - docs
---
# My Doc
Some content.
`
	docPath := filepath.Join(docDir, "test.md")
	os.WriteFile(docPath, []byte(doc), 0644)

	configPath = ""
	dryRun = false
	initStrategy = "interval"
	initInterval = "90d"

	err := runInit(nil, []string{tmpDir})
	if err != nil {
		t.Errorf("runInit() error = %v", err)
	}

	content, _ := os.ReadFile(docPath)
	contentStr := string(content)

	// Should have freshness added
	if !strings.Contains(contentStr, "freshness:") {
		t.Errorf("File should contain freshness block, got: %s", contentStr)
	}

	// Should preserve existing frontmatter
	if !strings.Contains(contentStr, "title: My Document") {
		t.Errorf("File should preserve title, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "author: Jane Doe") {
		t.Errorf("File should preserve author, got: %s", contentStr)
	}

	// Should have exactly one frontmatter block (count --- occurrences)
	delimCount := strings.Count(contentStr, "---")
	if delimCount != 2 {
		t.Errorf("File should have exactly 2 --- delimiters (one frontmatter block), got %d: %s", delimCount, contentStr)
	}

	// Content should still be there
	if !strings.Contains(contentStr, "# My Doc") {
		t.Errorf("File should preserve content, got: %s", contentStr)
	}
}

func TestReviewCommand_UpdatesDate(t *testing.T) {
	tmpDir := t.TempDir()

	doc := `---
freshness:
  last_reviewed: "2020-01-01"
  strategy: interval
  interval: 90d
---
# Doc
`
	docPath := filepath.Join(tmpDir, "test.md")
	os.WriteFile(docPath, []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	reviewDate = "2025-06-15"

	err := runReview(nil, []string{docPath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Errorf("runReview() error = %v", err)
	}

	content, _ := os.ReadFile(docPath)
	if !strings.Contains(string(content), "2025-06-15") {
		t.Errorf("File should contain new date, got: %s", string(content))
	}
}
