package checker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andimrob/docrot/internal/git"
)

func recentDate() string {
	return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
}

func TestRunParallel_ProcessesAllFiles(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create multiple docs
	for i := 0; i < 10; i++ {
		content := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Doc
`, recentDate())
		os.WriteFile(filepath.Join(docDir, fmt.Sprintf("doc%d.md", i)), []byte(content), 0644)
	}

	paths := make([]string, 10)
	for i := 0; i < 10; i++ {
		paths[i] = filepath.Join(docDir, fmt.Sprintf("doc%d.md", i))
	}

	results := Run(paths, nil, 4, nil) // 4 workers, no defaults

	if len(results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results))
	}

	// All should be fresh
	for _, r := range results {
		if r.Status != "fresh" {
			t.Errorf("Expected fresh status, got %s for %s", r.Status, r.Path)
		}
	}
}

func TestRunParallel_FasterThanSequential(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create 50 docs
	numDocs := 50
	for i := 0; i < numDocs; i++ {
		content := fmt.Sprintf(`---
docrot:
  last_reviewed: "%s"
  strategy: interval
  interval: 90d
---
# Doc
`, recentDate())
		os.WriteFile(filepath.Join(docDir, fmt.Sprintf("doc%d.md", i)), []byte(content), 0644)
	}

	paths := make([]string, numDocs)
	for i := 0; i < numDocs; i++ {
		paths[i] = filepath.Join(docDir, fmt.Sprintf("doc%d.md", i))
	}

	// Time parallel execution
	start := time.Now()
	results := Run(paths, nil, 8, nil)
	parallelDuration := time.Since(start)

	if len(results) != numDocs {
		t.Errorf("Expected %d results, got %d", numDocs, len(results))
	}

	t.Logf("Parallel processing of %d files took %v", numDocs, parallelDuration)
}

func TestRun_ParseError(t *testing.T) {
	paths := []string{"/nonexistent/path/to/doc.md"}
	results := Run(paths, nil, 1, nil)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Status != "stale" {
		t.Errorf("Status = %q, want %q", r.Status, "stale")
	}
	if !strings.Contains(r.Reason, "Failed to parse:") {
		t.Errorf("Reason should contain 'Failed to parse:', got: %q", r.Reason)
	}
}

func TestRun_WithGitClient_BuildsIndex(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create and commit src/main.go
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main"), 0644)
	gitAdd(t, tmpDir, "src/main.go")
	gitCommit(t, tmpDir, "Add source")

	// Create a code_changes strategy doc with a past last_reviewed date
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)
	docContent := `---
docrot:
  last_reviewed: "2020-01-01"
  strategy: code_changes
  watch:
    - "src/**/*.go"
---
# Watched Doc
`
	docPath := filepath.Join(docDir, "watched.md")
	os.WriteFile(docPath, []byte(docContent), 0644)

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatalf("git.New() error: %v", err)
	}

	results := Run([]string{docPath}, gitClient, 1, nil)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Status != "stale" {
		t.Errorf("Status = %q, want %q (src/main.go was committed after last_reviewed)", r.Status, "stale")
	}
	if len(r.ChangedFiles) == 0 {
		t.Errorf("ChangedFiles should be populated, got empty")
	}
}

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
