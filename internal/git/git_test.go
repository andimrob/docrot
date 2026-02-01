package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Initialize git repo
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgsign", "false"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}

	return tmpDir
}

func gitCommit(t *testing.T, dir, file, content, message string) {
	t.Helper()

	fullPath := filepath.Join(dir, file)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmds := [][]string{
		{"git", "add", file},
		{"git", "commit", "-m", message},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}
}

func TestLastCommitDate(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create and commit a file
	gitCommit(t, tmpDir, "doc/readme.md", "# Readme", "Initial commit")

	client, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	date, err := client.LastCommitDate("doc/readme.md")
	if err != nil {
		t.Fatalf("LastCommitDate() error = %v", err)
	}

	// Date should be today (within reasonable margin)
	now := time.Now()
	if date.Year() != now.Year() || date.Month() != now.Month() || date.Day() != now.Day() {
		t.Errorf("LastCommitDate() = %v, expected today", date)
	}
}

func TestLastCommitDate_FileNotInGit(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create initial commit with different file
	gitCommit(t, tmpDir, "other.txt", "content", "Initial")

	client, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.LastCommitDate("nonexistent.md")
	if err == nil {
		t.Error("LastCommitDate() expected error for file not in git")
	}
}

func TestFilesChangedSince(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create initial doc
	gitCommit(t, tmpDir, "doc/readme.md", "# Readme", "Add docs")

	// Wait a tiny bit and create code file
	gitCommit(t, tmpDir, "src/main.go", "package main", "Add code")

	client, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Get date of doc commit
	docDate, err := client.LastCommitDate("doc/readme.md")
	if err != nil {
		t.Fatal(err)
	}

	// Find files changed since doc was committed
	changed, err := client.FilesChangedSince(docDate, []string{"src/**/*.go"}, nil, tmpDir)
	if err != nil {
		t.Fatalf("FilesChangedSince() error = %v", err)
	}

	if len(changed) != 1 {
		t.Errorf("FilesChangedSince() found %d files, want 1", len(changed))
	}

	if len(changed) > 0 && changed[0].Path != "src/main.go" {
		t.Errorf("FilesChangedSince() path = %v, want src/main.go", changed[0].Path)
	}
}

// Tests for FileChangeIndex (batch git query optimization)

func TestBuildFileChangeIndex(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create multiple files
	gitCommit(t, tmpDir, "src/main.go", "package main", "Add main")
	gitCommit(t, tmpDir, "src/lib.go", "package lib", "Add lib")
	gitCommit(t, tmpDir, "doc/readme.md", "# Readme", "Add docs")

	client, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Build index since epoch (get all files)
	index, err := client.BuildFileChangeIndex(time.Time{})
	if err != nil {
		t.Fatalf("BuildFileChangeIndex() error = %v", err)
	}

	// Should have 3 files
	if len(index.files) != 3 {
		t.Errorf("Index has %d files, want 3", len(index.files))
	}

	// Check specific files exist
	if _, ok := index.files["src/main.go"]; !ok {
		t.Error("Index missing src/main.go")
	}
	if _, ok := index.files["src/lib.go"]; !ok {
		t.Error("Index missing src/lib.go")
	}
	if _, ok := index.files["doc/readme.md"]; !ok {
		t.Error("Index missing doc/readme.md")
	}
}

func TestFileChangeIndex_HasChangedSince(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create files
	gitCommit(t, tmpDir, "src/old.go", "package old", "Add old")
	gitCommit(t, tmpDir, "src/new.go", "package new", "Add new")

	client, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Build index
	index, err := client.BuildFileChangeIndex(time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	// Use a date far in the past - should find changes
	pastDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	changed := index.HasChangedSince(pastDate, []string{"**/*.go"}, nil)
	if !changed {
		t.Error("HasChangedSince(past) = false, want true")
	}

	// Use a date in the future - should NOT find changes
	futureDate := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	changed = index.HasChangedSince(futureDate, []string{"**/*.go"}, nil)
	if changed {
		t.Error("HasChangedSince(future) = true, want false")
	}

	// Check with pattern that doesn't match any files
	changed = index.HasChangedSince(pastDate, []string{"**/*.rb"}, nil)
	if changed {
		t.Error("HasChangedSince() = true, want false (no .rb files)")
	}
}

func TestFileChangeIndex_GetChangedFiles(t *testing.T) {
	tmpDir := setupGitRepo(t)

	gitCommit(t, tmpDir, "src/main.go", "package main", "Add main")
	gitCommit(t, tmpDir, "src/lib.rb", "class Lib", "Add lib")
	gitCommit(t, tmpDir, "doc/readme.md", "# Readme", "Add docs")

	client, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	mainDate, _ := client.LastCommitDate("src/main.go")

	index, err := client.BuildFileChangeIndex(time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	// Get .go files changed since main.go was committed
	// (should not include main.go itself since it wasn't changed AFTER that date)
	changed := index.GetChangedFiles(mainDate, []string{"**/*.go"}, nil)

	// main.go should not be included (it was committed at mainDate, not after)
	for _, f := range changed {
		if f.Path == "src/main.go" {
			t.Error("GetChangedFiles() should not include main.go (not changed AFTER mainDate)")
		}
	}
}

func TestFilesChangedSince_WithIgnorePatterns(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create code files and doc files
	gitCommit(t, tmpDir, "subsystem/src/main.go", "package main", "Add main")
	gitCommit(t, tmpDir, "subsystem/src/lib.go", "package lib", "Add lib")
	gitCommit(t, tmpDir, "subsystem/docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "subsystem/docs/guide.md", "# Guide", "Add guide")

	client, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	pastDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	// Watch all files in subsystem, ignore docs directory
	watch := []string{"subsystem/**/*"}
	ignore := []string{"subsystem/docs/**"}

	changed, err := client.FilesChangedSince(pastDate, watch, ignore, "")
	if err != nil {
		t.Fatalf("FilesChangedSince() error = %v", err)
	}

	// Should find the .go files but not the .md files in docs
	goFiles := 0
	mdFiles := 0
	for _, f := range changed {
		if filepath.Ext(f.Path) == ".go" {
			goFiles++
		}
		if filepath.Ext(f.Path) == ".md" {
			mdFiles++
		}
	}

	if goFiles != 2 {
		t.Errorf("Expected 2 .go files, got %d", goFiles)
	}
	if mdFiles != 0 {
		t.Errorf("Expected 0 .md files (ignored), got %d", mdFiles)
	}
}

func TestFileChangeIndex_GetChangedFiles_WithIgnore(t *testing.T) {
	tmpDir := setupGitRepo(t)

	gitCommit(t, tmpDir, "subsystem/src/main.go", "package main", "Add main")
	gitCommit(t, tmpDir, "subsystem/src/test_helper.go", "package main", "Add test helper")
	gitCommit(t, tmpDir, "subsystem/docs/readme.md", "# Readme", "Add docs")

	client, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	index, err := client.BuildFileChangeIndex(time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	pastDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	// Watch everything, ignore docs and test files
	watch := []string{"subsystem/**/*"}
	ignore := []string{"subsystem/docs/**", "**/*_helper.go"}

	changed := index.GetChangedFiles(pastDate, watch, ignore)

	// Should only find main.go (not test_helper.go or docs)
	if len(changed) != 1 {
		t.Errorf("Expected 1 file, got %d", len(changed))
	}
	if len(changed) > 0 && changed[0].Path != "subsystem/src/main.go" {
		t.Errorf("Expected subsystem/src/main.go, got %s", changed[0].Path)
	}
}

func TestFileChangeIndex_HasChangedSince_WithIgnore(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Only create files that will be ignored
	gitCommit(t, tmpDir, "subsystem/docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "subsystem/docs/guide.md", "# Guide", "Add guide")

	client, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	index, err := client.BuildFileChangeIndex(time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	pastDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	// Watch everything but ignore docs - should return false since all files are ignored
	watch := []string{"subsystem/**/*"}
	ignore := []string{"subsystem/docs/**"}

	changed := index.HasChangedSince(pastDate, watch, ignore)
	if changed {
		t.Error("HasChangedSince() = true, want false (all matching files are ignored)")
	}
}
