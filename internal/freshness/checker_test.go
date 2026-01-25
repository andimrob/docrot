package freshness

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/betterment/docrot/internal/document"
	"github.com/betterment/docrot/internal/git"
)

func TestCheck_IntervalStrategy_Fresh(t *testing.T) {
	// Doc reviewed 30 days ago with 90 day interval = fresh
	lastReviewed := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: lastReviewed,
			Strategy:     "interval",
			Interval:     "90d",
		},
	}

	checker := NewChecker(nil) // No git client needed for interval
	result := checker.Check(doc)

	if result.Status != StatusFresh {
		t.Errorf("Status = %v, want %v", result.Status, StatusFresh)
	}
}

func TestCheck_IntervalStrategy_Stale(t *testing.T) {
	// Doc reviewed 100 days ago with 90 day interval = stale
	lastReviewed := time.Now().AddDate(0, 0, -100).Format("2006-01-02")
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: lastReviewed,
			Strategy:     "interval",
			Interval:     "90d",
		},
	}

	checker := NewChecker(nil)
	result := checker.Check(doc)

	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v", result.Status, StatusStale)
	}

	if result.Reason == "" {
		t.Error("Expected reason to be set for stale doc")
	}
}

func TestCheck_UntilDateStrategy_Fresh(t *testing.T) {
	// Expires in the future = fresh
	expires := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: "2024-01-01",
			Strategy:     "until_date",
			Expires:      expires,
		},
	}

	checker := NewChecker(nil)
	result := checker.Check(doc)

	if result.Status != StatusFresh {
		t.Errorf("Status = %v, want %v", result.Status, StatusFresh)
	}
}

func TestCheck_UntilDateStrategy_Stale(t *testing.T) {
	// Expires in the past = stale
	expires := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: "2024-01-01",
			Strategy:     "until_date",
			Expires:      expires,
		},
	}

	checker := NewChecker(nil)
	result := checker.Check(doc)

	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v", result.Status, StatusStale)
	}
}

func TestCheck_MissingFrontmatter(t *testing.T) {
	doc := &document.Document{
		Path:      "doc/readme.md",
		Freshness: nil,
	}

	checker := NewChecker(nil)
	result := checker.Check(doc)

	if result.Status != StatusMissingFrontmatter {
		t.Errorf("Status = %v, want %v", result.Status, StatusMissingFrontmatter)
	}
}

func TestCheck_IntervalParsing(t *testing.T) {
	tests := []struct {
		interval string
		daysAgo  int
		want     Status
	}{
		{"30d", 25, StatusFresh},
		{"30d", 35, StatusStale},
		{"4w", 25, StatusFresh},  // 4 weeks = 28 days
		{"4w", 30, StatusStale},
		{"3m", 80, StatusFresh},  // 3 months = ~90 days
		{"3m", 100, StatusStale},
		{"1y", 300, StatusFresh}, // 1 year = 365 days
		{"1y", 400, StatusStale},
	}

	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			lastReviewed := time.Now().AddDate(0, 0, -tt.daysAgo).Format("2006-01-02")
			doc := &document.Document{
				Path: "doc/readme.md",
				Freshness: &document.Freshness{
					LastReviewed: lastReviewed,
					Strategy:     "interval",
					Interval:     tt.interval,
				},
			}

			checker := NewChecker(nil)
			result := checker.Check(doc)

			if result.Status != tt.want {
				t.Errorf("interval=%s, daysAgo=%d: Status = %v, want %v",
					tt.interval, tt.daysAgo, result.Status, tt.want)
			}
		})
	}
}

func setupGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

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

func TestCheck_CodeChangesStrategy_Fresh(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create doc first
	gitCommit(t, tmpDir, "doc/readme.md", "# Readme", "Add docs")

	// Create code file BEFORE the doc's last_reviewed date (simulate no changes)
	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Use today as last_reviewed, so no code was changed after
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "doc/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().Format("2006-01-02"),
			Strategy:     "code_changes",
			Watch:        []string{"**/*.go"},
		},
	}

	checker := NewChecker(gitClient)
	result := checker.Check(doc)

	if result.Status != StatusFresh {
		t.Errorf("Status = %v, want %v (reason: %s)", result.Status, StatusFresh, result.Reason)
	}
}

func TestCheck_CodeChangesStrategy_Stale(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create initial commit
	gitCommit(t, tmpDir, "doc/readme.md", "# Readme", "Add docs")

	// Create code file after (simulate code change)
	gitCommit(t, tmpDir, "src/main.go", "package main", "Add code")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Use yesterday as last_reviewed
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "doc/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
			Watch:        []string{"**/*.go"},
		},
	}

	checker := NewChecker(gitClient)
	result := checker.Check(doc)

	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v", result.Status, StatusStale)
	}

	if result.Reason == "" {
		t.Error("Expected reason mentioning changed file")
	}
}

// Tests for CheckWithIndex (batch optimization)

func TestCheckWithIndex_CodeChanges_Fresh(t *testing.T) {
	tmpDir := setupGitRepo(t)

	gitCommit(t, tmpDir, "src/main.go", "package main", "Add code")
	gitCommit(t, tmpDir, "doc/readme.md", "# Readme", "Add docs")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Build index
	index, err := gitClient.BuildFileChangeIndex(time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	// Doc reviewed "tomorrow" - so no code changes are after that
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "doc/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: tomorrow,
			Strategy:     "code_changes",
			Watch:        []string{"**/*.go"},
		},
	}

	checker := NewChecker(gitClient)
	result := checker.CheckWithIndex(doc, index)

	if result.Status != StatusFresh {
		t.Errorf("Status = %v, want %v (reason: %s)", result.Status, StatusFresh, result.Reason)
	}
}

func TestCheckWithIndex_CodeChanges_Stale(t *testing.T) {
	tmpDir := setupGitRepo(t)

	gitCommit(t, tmpDir, "doc/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "src/main.go", "package main", "Add code")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	index, err := gitClient.BuildFileChangeIndex(time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	// Doc reviewed in the past
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "doc/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: "2020-01-01",
			Strategy:     "code_changes",
			Watch:        []string{"**/*.go"},
		},
	}

	checker := NewChecker(gitClient)
	result := checker.CheckWithIndex(doc, index)

	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v", result.Status, StatusStale)
	}
}

func TestCheckWithIndex_IntervalStrategy_StillWorks(t *testing.T) {
	// Non-code_changes strategies should still work with index (index is ignored)
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: time.Now().Format("2006-01-02"),
			Strategy:     "interval",
			Interval:     "90d",
		},
	}

	checker := NewChecker(nil)
	result := checker.CheckWithIndex(doc, nil)

	if result.Status != StatusFresh {
		t.Errorf("Status = %v, want %v", result.Status, StatusFresh)
	}
}
