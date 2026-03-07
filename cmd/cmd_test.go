package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckCommand_WithStaleDoc(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create a stale doc
	staleDoc := `---
docrot:
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
	freshDoc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Fresh Doc
`, recentDate())
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

	doc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Doc
`, recentDate())
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

func TestAddFrontmatterCommand_DryRun(t *testing.T) {
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
	addFrontmatterDryRun = true

	err := runAddFrontmatter(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runAddFrontmatter() error = %v", err)
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

func TestInitCommand_CreatesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runInit(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Errorf("runInit() error = %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, ".docrot.yml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "patterns:") {
		t.Errorf("Config should contain patterns, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "defaults:") {
		t.Errorf("Config should contain defaults section, got: %s", contentStr)
	}
}

func TestInitCommand_DoesNotOverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create existing config
	existingConfig := "# My existing config\npatterns:\n  - custom/*.md\n"
	os.WriteFile(filepath.Join(tmpDir, ".docrot.yml"), []byte(existingConfig), 0644)

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := runInit(nil, []string{})

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err == nil {
		t.Error("runInit() should return error when config exists")
	}

	// Verify config file was not overwritten
	content, _ := os.ReadFile(filepath.Join(tmpDir, ".docrot.yml"))
	if string(content) != existingConfig {
		t.Errorf("Config should not be overwritten, got: %s", string(content))
	}
}

func TestCheckCommand_PatternFlag(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a doc in a non-standard directory (not "doc" or "docs")
	rfcsDir := filepath.Join(tmpDir, "rfcs")
	os.MkdirAll(rfcsDir, 0755)

	doc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# RFC 001
`, recentDate())
	os.WriteFile(filepath.Join(rfcsDir, "rfc-001.md"), []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false
	patternFlag = []string{"**/*.md"} // Override config patterns via flag

	err := runCheck(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout
	patternFlag = nil // reset global

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runCheck() error = %v", err)
	}

	if !strings.Contains(output, "rfc-001.md") {
		t.Errorf("Output should contain rfc-001.md when using --pattern flag, got: %s", output)
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

func TestAddFrontmatterCommand_MergesWithExistingFrontmatter(t *testing.T) {
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
	addFrontmatterDryRun = false
	addFrontmatterStrategy = "interval"
	addFrontmatterInterval = "90d"

	err := runAddFrontmatter(nil, []string{tmpDir})
	if err != nil {
		t.Errorf("runInit() error = %v", err)
	}

	content, _ := os.ReadFile(docPath)
	contentStr := string(content)

	// Should have freshness added
	if !strings.Contains(contentStr, "docrot:") {
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
docrot:
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

func TestFilesCommand_WithExplicitPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a real git repo so root detection works
	exec.Command("git", "-C", tmpDir, "init").Run()

	// Create directory structure
	os.MkdirAll(filepath.Join(tmpDir, "subsystem/src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "subsystem/docs"), 0755)

	// Create source files
	os.WriteFile(filepath.Join(tmpDir, "subsystem/src/main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subsystem/src/lib.go"), []byte("package lib"), 0644)

	// Create doc with explicit watch/ignore patterns
	doc := `---
docrot:
  last_reviewed: "2024-01-01"
  strategy: interval
  interval: 90d
  watch:
    - "subsystem/**/*"
  ignore:
    - "subsystem/docs/**"
---
# Doc
`
	docPath := filepath.Join(tmpDir, "subsystem/docs/readme.md")
	os.WriteFile(docPath, []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	format = "text"
	err := runFiles(nil, []string{docPath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runFiles() error = %v", err)
	}

	// Should include source files
	if !strings.Contains(output, "main.go") {
		t.Errorf("Output should contain main.go, got: %s", output)
	}
	if !strings.Contains(output, "lib.go") {
		t.Errorf("Output should contain lib.go, got: %s", output)
	}

	// Should NOT include docs (ignored)
	if strings.Contains(output, "readme.md") {
		t.Errorf("Output should NOT contain readme.md (ignored), got: %s", output)
	}
}

func TestCheckCommand_NonexistentFile(t *testing.T) {
	configPath = ""
	format = "text"
	quiet = false

	err := runCheck(nil, []string{"/nonexistent/path/file.md"})

	if err == nil {
		t.Error("runCheck() should return error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "failed to stat") {
		t.Errorf("Error should mention 'failed to stat', got: %v", err)
	}
}

func TestCheckCommand_MixedFilesAndDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a doc directory with a file
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	dirDoc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Doc in directory
`, recentDate())
	os.WriteFile(filepath.Join(docDir, "in-dir.md"), []byte(dirDoc), 0644)

	// Create a standalone file
	standaloneDoc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Standalone Doc
`, recentDate())
	standalonePath := filepath.Join(tmpDir, "standalone.md")
	os.WriteFile(standalonePath, []byte(standaloneDoc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false

	// Pass both a directory and a file
	err := runCheck(nil, []string{tmpDir, standalonePath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runCheck() error = %v", err)
	}

	// Both files should be in output
	if !strings.Contains(output, "in-dir.md") {
		t.Errorf("Output should contain in-dir.md, got: %s", output)
	}
	if !strings.Contains(output, "standalone.md") {
		t.Errorf("Output should contain standalone.md, got: %s", output)
	}
}

func TestCheckCommand_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two doc files
	freshDoc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Fresh Doc
`, recentDate())
	staleDoc := `---
docrot:
  last_reviewed: "2020-01-01"
  strategy: interval
  interval: 30d
---
# Stale Doc
`
	freshPath := filepath.Join(tmpDir, "fresh.md")
	stalePath := filepath.Join(tmpDir, "stale.md")
	os.WriteFile(freshPath, []byte(freshDoc), 0644)
	os.WriteFile(stalePath, []byte(staleDoc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false

	// Pass multiple file paths
	err := runCheck(nil, []string{freshPath, stalePath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should return error because one doc is stale
	if err != ErrStaleDocsFound {
		t.Errorf("runCheck() error = %v, want ErrStaleDocsFound", err)
	}

	// Both files should be in output
	if !strings.Contains(output, "fresh.md") {
		t.Errorf("Output should contain fresh.md, got: %s", output)
	}
	if !strings.Contains(output, "stale.md") {
		t.Errorf("Output should contain stale.md, got: %s", output)
	}
}

func TestCheckCommand_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a doc file (not in a doc/ subdirectory)
	doc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Single File Test
`, recentDate())
	docPath := filepath.Join(tmpDir, "test.md")
	os.WriteFile(docPath, []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false

	// Pass the file path directly, not a directory
	err := runCheck(nil, []string{docPath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runCheck() error = %v", err)
	}

	if !strings.Contains(output, "test.md") {
		t.Errorf("Output should contain test.md, got: %s", output)
	}

	if !strings.Contains(output, "fresh") {
		t.Errorf("Output should contain 'fresh', got: %s", output)
	}
}

func TestFilesCommand_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a real git repo so root detection works
	exec.Command("git", "-C", tmpDir, "init").Run()

	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "docs"), 0755)

	os.WriteFile(filepath.Join(tmpDir, "src/main.go"), []byte("package main"), 0644)

	doc := `---
docrot:
  last_reviewed: "2024-01-01"
  strategy: interval
  interval: 90d
  watch:
    - "src/**/*"
---
# Doc
`
	docPath := filepath.Join(tmpDir, "docs/readme.md")
	os.WriteFile(docPath, []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	format = "json"
	err := runFiles(nil, []string{docPath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runFiles() error = %v", err)
	}

	// Should be valid JSON with files array
	if !strings.Contains(output, `"files"`) {
		t.Errorf("JSON output should contain files field, got: %s", output)
	}
	if !strings.Contains(output, "main.go") {
		t.Errorf("JSON output should contain main.go, got: %s", output)
	}
}

func TestCheckCommand_PathsInDifferentRepos_ReturnsError(t *testing.T) {
	// Create two separate git repos
	repo1 := t.TempDir()
	repo2 := t.TempDir()

	// Initialize both as git repos
	for _, repo := range []string{repo1, repo2} {
		initGitRepo(t, repo)
	}

	// Create doc directories and files in each repo
	doc1Dir := filepath.Join(repo1, "doc")
	doc2Dir := filepath.Join(repo2, "doc")
	os.MkdirAll(doc1Dir, 0755)
	os.MkdirAll(doc2Dir, 0755)

	doc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Doc
`, recentDate())
	doc1Path := filepath.Join(doc1Dir, "readme.md")
	doc2Path := filepath.Join(doc2Dir, "readme.md")
	os.WriteFile(doc1Path, []byte(doc), 0644)
	os.WriteFile(doc2Path, []byte(doc), 0644)

	configPath = ""
	format = "text"
	quiet = false

	// Pass paths from two different repos
	err := runCheck(nil, []string{doc1Path, doc2Path})

	if err == nil {
		t.Error("runCheck() should return error when paths are in different repos")
	}

	if err != nil && !strings.Contains(err.Error(), "different git repositories") {
		t.Errorf("Error should mention 'different git repositories', got: %v", err)
	}
}

func TestCheckCommand_PathsInSameRepo_Succeeds(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create two doc directories in the same repo
	doc1Dir := filepath.Join(tmpDir, "module1/doc")
	doc2Dir := filepath.Join(tmpDir, "module2/doc")
	os.MkdirAll(doc1Dir, 0755)
	os.MkdirAll(doc2Dir, 0755)

	doc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Doc
`, recentDate())
	doc1Path := filepath.Join(doc1Dir, "readme.md")
	doc2Path := filepath.Join(doc2Dir, "readme.md")
	os.WriteFile(doc1Path, []byte(doc), 0644)
	os.WriteFile(doc2Path, []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false

	err := runCheck(nil, []string{doc1Path, doc2Path})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runCheck() error = %v", err)
	}

	// Both files should be checked
	if !strings.Contains(output, "module1") {
		t.Errorf("Output should contain module1, got: %s", output)
	}
	if !strings.Contains(output, "module2") {
		t.Errorf("Output should contain module2, got: %s", output)
	}
}

func TestCheckCommand_DirectoryInDifferentRepo_ReturnsError(t *testing.T) {
	// Create two separate git repos
	repo1 := t.TempDir()
	repo2 := t.TempDir()

	initGitRepo(t, repo1)
	initGitRepo(t, repo2)

	// Create doc in repo1 and directory with docs in repo2
	doc1Dir := filepath.Join(repo1, "doc")
	doc2Dir := filepath.Join(repo2, "doc")
	os.MkdirAll(doc1Dir, 0755)
	os.MkdirAll(doc2Dir, 0755)

	doc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Doc
`, recentDate())
	doc1Path := filepath.Join(doc1Dir, "readme.md")
	os.WriteFile(doc1Path, []byte(doc), 0644)
	os.WriteFile(filepath.Join(doc2Dir, "other.md"), []byte(doc), 0644)

	configPath = ""
	format = "text"
	quiet = false

	// Pass a file from repo1 and directory from repo2
	err := runCheck(nil, []string{doc1Path, repo2})

	if err == nil {
		t.Error("runCheck() should return error when paths are in different repos")
	}

	if err != nil && !strings.Contains(err.Error(), "different git repositories") {
		t.Errorf("Error should mention 'different git repositories', got: %v", err)
	}
}

func TestCheckCommand_SingleFileUsesCorrectGitRoot(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create and commit a source file
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	srcPath := filepath.Join(srcDir, "main.go")
	os.WriteFile(srcPath, []byte("package main"), 0644)
	gitAdd(t, tmpDir, "src/main.go")
	gitCommit(t, tmpDir, "Add source")

	// Create doc that watches src
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Use a very old date so that any code change makes it stale
	doc := `---
docrot:
  last_reviewed: "2020-01-01"
  strategy: code_changes
  watch:
    - "src/**/*.go"
---
# Doc
`
	docPath := filepath.Join(docDir, "readme.md")
	os.WriteFile(docPath, []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false

	// Run from a different directory but pass absolute path
	oldWd, _ := os.Getwd()
	os.Chdir(t.TempDir()) // Change to unrelated directory
	defer os.Chdir(oldWd)

	err := runCheck(nil, []string{docPath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should detect the doc as stale because src/main.go was committed
	if err != ErrStaleDocsFound {
		t.Errorf("runCheck() error = %v, want ErrStaleDocsFound (git should detect code changes)", err)
	}

	if !strings.Contains(output, "stale") {
		t.Errorf("Output should show doc as stale, got: %s", output)
	}
}

// Helper to initialize a git repo for testing
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgsign", "false"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}
	// Create initial commit so repo is valid
	placeholderPath := filepath.Join(dir, ".gitkeep")
	os.WriteFile(placeholderPath, []byte(""), 0644)
	gitAdd(t, dir, ".gitkeep")
	gitCommit(t, dir, "Initial commit")
}

func gitAdd(t *testing.T, dir, file string) {
	t.Helper()
	cmd := exec.Command("git", "add", file)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}
}

func gitCommit(t *testing.T, dir, message string) {
	t.Helper()
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}
}

func TestFilesCommand_SmartDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create structure: subsystem/docs/readme.md and subsystem/src/main.go
	os.MkdirAll(filepath.Join(tmpDir, "subsystem/src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "subsystem/docs"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subsystem/src/main.go"), []byte("package main"), 0644)

	// Doc with no watch/ignore — should use smart defaults
	doc := `---
docrot:
  last_reviewed: "2024-01-01"
  strategy: interval
  interval: 90d
---
# Doc
`
	docPath := filepath.Join(tmpDir, "subsystem/docs/readme.md")
	os.WriteFile(docPath, []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	format = "text"
	err := runFiles(nil, []string{docPath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runFiles() error = %v", err)
	}

	// Smart defaults: watch subsystem/**/* , ignore subsystem/docs/**
	// src/main.go should be included, readme.md should not
	if !strings.Contains(output, "main.go") {
		t.Errorf("Output should contain main.go (watched by smart defaults), got: %s", output)
	}
	if strings.Contains(output, "readme.md") {
		t.Errorf("Output should NOT contain readme.md (ignored by smart defaults), got: %s", output)
	}
}

func TestCheckCommand_QuietFlag(t *testing.T) {
	tmpDir := t.TempDir()

	freshDoc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Fresh Doc
`, recentDate())
	staleDoc := `---
docrot:
  last_reviewed: "2020-01-01"
  strategy: interval
  interval: 30d
---
# Stale Doc
`
	freshPath := filepath.Join(tmpDir, "fresh.md")
	stalePath := filepath.Join(tmpDir, "stale.md")
	os.WriteFile(freshPath, []byte(freshDoc), 0644)
	os.WriteFile(stalePath, []byte(staleDoc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = true

	err := runCheck(nil, []string{freshPath, stalePath})

	w.Close()
	os.Stdout = oldStdout
	quiet = false // reset global

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != ErrStaleDocsFound {
		t.Errorf("runCheck() error = %v, want ErrStaleDocsFound", err)
	}

	// Quiet mode: stale doc should appear, fresh doc should not
	if !strings.Contains(output, "stale.md") {
		t.Errorf("Quiet output should contain stale.md, got: %s", output)
	}
	if strings.Contains(output, "fresh.md") {
		t.Errorf("Quiet output should NOT contain fresh.md, got: %s", output)
	}
}

func TestListCommand_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	doc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Fresh Doc
`, recentDate())
	os.WriteFile(filepath.Join(docDir, "fresh.md"), []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"

	err := runList(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runList() error = %v", err)
	}
	if !strings.Contains(output, "fresh.md") {
		t.Errorf("Output should contain 'fresh.md', got: %s", output)
	}
	if !strings.Contains(output, "fresh") {
		t.Errorf("Output should contain 'fresh', got: %s", output)
	}
}

func TestListCommand_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	doc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Fresh Doc
`, recentDate())
	os.WriteFile(filepath.Join(docDir, "fresh.md"), []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "json"

	err := runList(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout
	format = "text" // reset

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runList() error = %v", err)
	}
	if !strings.Contains(output, `"status"`) {
		t.Errorf("JSON output should contain 'status' field, got: %s", output)
	}
}

func TestCheckCommand_StrictMode_MissingFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create doc without frontmatter
	doc := `# No Frontmatter
Some content.
`
	os.WriteFile(filepath.Join(docDir, "missing.md"), []byte(doc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false
	strictMode = true

	err := runCheck(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout
	strictMode = false // reset

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != ErrMissingFrontmatterFound {
		t.Errorf("runCheck() error = %v, want ErrMissingFrontmatterFound", err)
	}

	// Doc should appear in output (not skipped)
	if !strings.Contains(output, "missing.md") {
		t.Errorf("Output should contain missing.md, got: %s", output)
	}
}

func TestCheckCommand_StrictMode_AllFresh_Succeeds(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	freshDoc := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Fresh Doc
`, recentDate())
	os.WriteFile(filepath.Join(docDir, "fresh.md"), []byte(freshDoc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false
	strictMode = true

	err := runCheck(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout
	strictMode = false // reset

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Errorf("runCheck() error = %v, want nil", err)
	}
}

func TestCheckCommand_StrictMode_StaleAndMissing_ReturnsStaleError(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	staleDoc := `---
docrot:
  last_reviewed: "2020-01-01"
  strategy: interval
  interval: 30d
---
# Stale Doc
`
	missingDoc := `# No Frontmatter
Some content.
`
	os.WriteFile(filepath.Join(docDir, "stale.md"), []byte(staleDoc), 0644)
	os.WriteFile(filepath.Join(docDir, "missing.md"), []byte(missingDoc), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = ""
	format = "text"
	quiet = false
	strictMode = true

	err := runCheck(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout
	strictMode = false // reset

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Stale error takes priority over missing frontmatter error
	if err != ErrStaleDocsFound {
		t.Errorf("runCheck() error = %v, want ErrStaleDocsFound", err)
	}
}

func TestCheckCommand_StrictMode_ViaConfig(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create doc without frontmatter
	doc := `# No Frontmatter
Some content.
`
	os.WriteFile(filepath.Join(docDir, "missing.md"), []byte(doc), 0644)

	// Write config with on_missing_frontmatter: strict
	cfgContent := "on_missing_frontmatter: strict\n"
	cfgPath := filepath.Join(tmpDir, ".docrot.yml")
	os.WriteFile(cfgPath, []byte(cfgContent), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configPath = cfgPath
	format = "text"
	quiet = false
	strictMode = false // flag not set; config drives it

	err := runCheck(nil, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout
	configPath = "" // reset

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != ErrMissingFrontmatterFound {
		t.Errorf("runCheck() error = %v, want ErrMissingFrontmatterFound", err)
	}
}

// recentDate returns yesterday's date formatted as YYYY-MM-DD.
// Using a recent past date ensures docs with a 90d interval are always fresh.
func recentDate() string {
	return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
}
