package freshness

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/andimrob/docrot/internal/document"
	"github.com/andimrob/docrot/internal/git"
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

	checker := NewChecker(nil, "", nil) // No git client needed for interval
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

	checker := NewChecker(nil, "", nil)
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

	checker := NewChecker(nil, "", nil)
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

	checker := NewChecker(nil, "", nil)
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

	checker := NewChecker(nil, "", nil)
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

			checker := NewChecker(nil, "", nil)
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

	checker := NewChecker(gitClient, tmpDir, nil)
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

	checker := NewChecker(gitClient, tmpDir, nil)
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

	checker := NewChecker(gitClient, tmpDir, nil)
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

	checker := NewChecker(gitClient, tmpDir, nil)
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

	checker := NewChecker(nil, "", nil)
	result := checker.CheckWithIndex(doc, nil)

	if result.Status != StatusFresh {
		t.Errorf("Status = %v, want %v", result.Status, StatusFresh)
	}
}

// Tests for smart default patterns based on document location

func TestComputeDefaultPatterns(t *testing.T) {
	tests := []struct {
		name       string
		docPath    string
		repoRoot   string
		wantWatch  string
		wantIgnore string // empty string means expect no ignore patterns
	}{
		{
			name:       "doc in subsystem/docs",
			docPath:    "/repo/subsystem/docs/readme.md",
			repoRoot:   "/repo",
			wantWatch:  "subsystem/**/*",
			wantIgnore: "subsystem/docs/**",
		},
		{
			name:       "doc in subsystem/doc (singular)",
			docPath:    "/repo/subsystem/doc/readme.md",
			repoRoot:   "/repo",
			wantWatch:  "subsystem/**/*",
			wantIgnore: "subsystem/doc/**",
		},
		{
			name:       "doc in subsystem/documentation",
			docPath:    "/repo/subsystem/documentation/readme.md",
			repoRoot:   "/repo",
			wantWatch:  "subsystem/**/*",
			wantIgnore: "subsystem/documentation/**",
		},
		{
			name:       "doc in nested docs dir",
			docPath:    "/repo/subsystem/docs/guides/setup.md",
			repoRoot:   "/repo",
			wantWatch:  "subsystem/**/*",
			wantIgnore: "subsystem/docs/**",
		},
		{
			name:       "doc at repo root docs",
			docPath:    "/repo/docs/readme.md",
			repoRoot:   "/repo",
			wantWatch:  "**/*",
			wantIgnore: "docs/**",
		},
		{
			name:       "doc directly in subsystem (no docs subdir)",
			docPath:    "/repo/subsystem/readme.md",
			repoRoot:   "/repo",
			wantWatch:  "subsystem/**/*",
			wantIgnore: "",
		},
		{
			name:       "doc at repo root (no docs dir)",
			docPath:    "/repo/readme.md",
			repoRoot:   "/repo",
			wantWatch:  "**/*",
			wantIgnore: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watch, ignore := ComputeDefaultPatterns(tt.docPath, tt.repoRoot)

			if len(watch) != 1 || watch[0] != tt.wantWatch {
				t.Errorf("watch = %v, want [%s]", watch, tt.wantWatch)
			}
			if tt.wantIgnore == "" {
				if len(ignore) != 0 {
					t.Errorf("ignore = %v, want []", ignore)
				}
			} else {
				if len(ignore) != 1 || ignore[0] != tt.wantIgnore {
					t.Errorf("ignore = %v, want [%s]", ignore, tt.wantIgnore)
				}
			}
		})
	}
}

// Tests for code_changes strategy with ignore patterns

func TestCheck_CodeChangesStrategy_WithIgnore(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create files - only docs, no code
	gitCommit(t, tmpDir, "subsystem/docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "subsystem/docs/guide.md", "# Guide", "Add guide")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Watch all, ignore docs - should be fresh since only ignored files exist
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "subsystem/docs/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
			Watch:        []string{"subsystem/**/*"},
			Ignore:       []string{"subsystem/docs/**"},
		},
	}

	checker := NewChecker(gitClient, tmpDir, nil)
	result := checker.Check(doc)

	if result.Status != StatusFresh {
		t.Errorf("Status = %v, want %v (only ignored files changed)", result.Status, StatusFresh)
	}
}

// Test using smart defaults (no explicit watch/ignore in frontmatter)

func TestCheck_SmartDefaults(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create structure: subsystem/docs/readme.md and subsystem/src/main.go
	gitCommit(t, tmpDir, "subsystem/docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "subsystem/src/main.go", "package main", "Add code")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// No watch/ignore specified - should use smart defaults
	// (watch subsystem/**/* , ignore subsystem/docs/**)
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "subsystem/docs/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
			// No Watch or Ignore - uses smart defaults
		},
	}

	checker := NewChecker(gitClient, tmpDir, nil)
	result := checker.Check(doc)

	// Code file changed after last_reviewed, should be stale
	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v (smart defaults should detect code change)", result.Status, StatusStale)
	}
}

func TestCheck_SmartDefaults_OnlyDocsChanged(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Only create docs, no code files
	gitCommit(t, tmpDir, "subsystem/docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "subsystem/docs/guide.md", "# Guide", "Add guide")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// No watch/ignore specified - smart defaults should ignore docs dir
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "subsystem/docs/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
		},
	}

	checker := NewChecker(gitClient, tmpDir, nil)
	result := checker.Check(doc)

	// Only docs changed (ignored by default), should be fresh
	if result.Status != StatusFresh {
		t.Errorf("Status = %v, want %v (smart defaults should ignore docs dir)", result.Status, StatusFresh)
	}
}

// Tests for config defaults merging with document frontmatter

func TestCheck_ConfigDefaults_InheritIgnoreWhenWatchSet(t *testing.T) {
	// When doc has watch but no ignore, it should inherit ignore from config defaults
	tmpDir := setupGitRepo(t)

	// Create code and test files
	gitCommit(t, tmpDir, "src/main.go", "package main", "Add code")
	gitCommit(t, tmpDir, "src/main_test.go", "package main", "Add test")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Doc has watch but no ignore
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "docs/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
			Watch:        []string{"src/**/*.go"}, // Watch all Go files
			// No Ignore - should inherit from config defaults
		},
	}

	// Config defaults ignore test files
	defaults := &DefaultPatterns{
		Watch:  []string{"**/*.go"},
		Ignore: []string{"**/*_test.go"},
	}

	checker := NewChecker(gitClient, tmpDir, defaults)
	result := checker.Check(doc)

	// Only test file changed after main.go, but tests are ignored via config defaults
	// The main.go was committed, so it should be stale due to main.go change
	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v", result.Status, StatusStale)
	}
}

func TestCheck_ConfigDefaults_IgnoreTestFiles(t *testing.T) {
	// Config defaults should allow ignoring test files globally
	tmpDir := setupGitRepo(t)

	// Create only a test file change
	gitCommit(t, tmpDir, "docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "src/main_test.go", "package main", "Add test")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Doc has watch but no ignore - should inherit ignore from defaults
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "docs/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
			Watch:        []string{"src/**/*.go"},
			// No Ignore - inherits from config defaults
		},
	}

	defaults := &DefaultPatterns{
		Ignore: []string{"**/*_test.go"},
	}

	checker := NewChecker(gitClient, tmpDir, defaults)
	result := checker.Check(doc)

	// Only test file changed, which is ignored - should be fresh
	if result.Status != StatusFresh {
		t.Errorf("Status = %v, want %v (test file should be ignored via config defaults)", result.Status, StatusFresh)
	}
}

func TestCheck_ConfigDefaults_InheritWatchWhenIgnoreSet(t *testing.T) {
	// When doc has ignore but no watch, it should inherit watch from config defaults
	tmpDir := setupGitRepo(t)

	gitCommit(t, tmpDir, "docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "cmd/main.go", "package main", "Add cmd")
	gitCommit(t, tmpDir, "scripts/build.sh", "#!/bin/bash", "Add script")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Doc has ignore but no watch - should inherit watch from config defaults
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "docs/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
			// No Watch - inherits from config defaults
			Ignore: []string{"scripts/**"}, // Ignore scripts
		},
	}

	defaults := &DefaultPatterns{
		Watch: []string{"cmd/**/*.go", "internal/**/*.go"},
	}

	checker := NewChecker(gitClient, tmpDir, defaults)
	result := checker.Check(doc)

	// cmd/main.go changed and is watched via defaults - should be stale
	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v (should detect cmd change via inherited watch)", result.Status, StatusStale)
	}
}

func TestCheck_ConfigDefaults_UsedWhenNoFrontmatterPatterns(t *testing.T) {
	// When doc has neither watch nor ignore, config defaults should be used
	tmpDir := setupGitRepo(t)

	gitCommit(t, tmpDir, "docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "internal/pkg/lib.go", "package pkg", "Add lib")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Doc has no watch or ignore
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "docs/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
			// No Watch, no Ignore - uses config defaults
		},
	}

	defaults := &DefaultPatterns{
		Watch:  []string{"internal/**/*.go"},
		Ignore: []string{"**/*_test.go"},
	}

	checker := NewChecker(gitClient, tmpDir, defaults)
	result := checker.Check(doc)

	// internal/pkg/lib.go changed and matches config defaults watch
	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v (config defaults should detect change)", result.Status, StatusStale)
	}
}

func TestCheck_ConfigDefaults_FrontmatterTakesPriority(t *testing.T) {
	// Explicit frontmatter patterns should take priority over config defaults
	tmpDir := setupGitRepo(t)

	gitCommit(t, tmpDir, "docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "internal/pkg/lib.go", "package pkg", "Add lib")
	gitCommit(t, tmpDir, "cmd/main.go", "package main", "Add cmd")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Doc has explicit watch and ignore - should NOT use config defaults
	doc := &document.Document{
		Path: filepath.Join(tmpDir, "docs/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
			Watch:        []string{"cmd/**/*.go"}, // Only watch cmd
			Ignore:       []string{"cmd/debug/**"}, // Ignore debug
		},
	}

	// Config defaults watch internal, but frontmatter overrides
	defaults := &DefaultPatterns{
		Watch:  []string{"internal/**/*.go"},
		Ignore: []string{"**/*_test.go"},
	}

	checker := NewChecker(gitClient, tmpDir, defaults)
	result := checker.Check(doc)

	// cmd/main.go changed (watched by frontmatter) - should be stale
	// internal/pkg/lib.go changed but NOT watched (frontmatter overrides defaults)
	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v", result.Status, StatusStale)
	}
	if result.Reason == "" || !contains(result.Reason, "cmd/main.go") {
		t.Errorf("Reason should mention cmd/main.go, got: %s", result.Reason)
	}
}

func TestCheck_ConfigDefaults_NilDefaultsStillWorks(t *testing.T) {
	// When defaults is nil, existing behavior (smart defaults) should work
	tmpDir := setupGitRepo(t)

	gitCommit(t, tmpDir, "subsystem/docs/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "subsystem/src/main.go", "package main", "Add code")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	doc := &document.Document{
		Path: filepath.Join(tmpDir, "subsystem/docs/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     "code_changes",
			// No Watch, no Ignore
		},
	}

	// nil defaults - should fall back to smart defaults
	checker := NewChecker(gitClient, tmpDir, nil)
	result := checker.Check(doc)

	// Smart defaults should detect subsystem/src/main.go change
	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v (smart defaults should work with nil config defaults)", result.Status, StatusStale)
	}
}

// Tests for error paths (invalid dates, intervals, unknown strategies)

func TestCheck_InvalidLastReviewedDate(t *testing.T) {
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: "not-a-date",
			Strategy:     StrategyInterval,
			Interval:     "90d",
		},
	}

	checker := NewChecker(nil, "", nil)
	result := checker.Check(doc)

	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v for invalid date", result.Status, StatusStale)
	}
	if result.Reason == "" {
		t.Error("Expected reason to be set for invalid date")
	}
}

func TestCheck_InvalidInterval(t *testing.T) {
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: time.Now().Format("2006-01-02"),
			Strategy:     StrategyInterval,
			Interval:     "90x", // invalid unit
		},
	}

	checker := NewChecker(nil, "", nil)
	result := checker.Check(doc)

	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v for invalid interval", result.Status, StatusStale)
	}
	if result.Reason == "" {
		t.Error("Expected reason to be set for invalid interval")
	}
}

func TestCheck_InvalidExpiresDate(t *testing.T) {
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: "2024-01-01",
			Strategy:     StrategyUntilDate,
			Expires:      "not-a-date",
		},
	}

	checker := NewChecker(nil, "", nil)
	result := checker.Check(doc)

	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v for invalid expires date", result.Status, StatusStale)
	}
	if result.Reason == "" {
		t.Error("Expected reason to be set for invalid expires date")
	}
}

func TestCheck_UnknownStrategy(t *testing.T) {
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: "2024-01-01",
			Strategy:     "bogus",
		},
	}

	checker := NewChecker(nil, "", nil)
	result := checker.Check(doc)

	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v for unknown strategy", result.Status, StatusStale)
	}
	if !contains(result.Reason, "bogus") {
		t.Errorf("Reason should mention unknown strategy name, got: %s", result.Reason)
	}
}

func TestCheck_CodeChanges_InvalidLastReviewedDate(t *testing.T) {
	doc := &document.Document{
		Path: "doc/readme.md",
		Freshness: &document.Freshness{
			LastReviewed: "not-a-date",
			Strategy:     StrategyCodeChanges,
			Watch:        []string{"**/*.go"},
		},
	}

	checker := NewChecker(nil, "", nil)
	result := checker.Check(doc)

	if result.Status != StatusStale {
		t.Errorf("Status = %v, want %v for invalid date in code_changes", result.Status, StatusStale)
	}
	if result.Reason == "" {
		t.Error("Expected reason to be set for invalid date")
	}
}

func TestCheck_CodeChanges_ChangedFilesPopulated(t *testing.T) {
	tmpDir := setupGitRepo(t)

	gitCommit(t, tmpDir, "doc/readme.md", "# Readme", "Add docs")
	gitCommit(t, tmpDir, "src/main.go", "package main", "Add main")
	gitCommit(t, tmpDir, "src/lib.go", "package lib", "Add lib")

	gitClient, err := git.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	doc := &document.Document{
		Path: filepath.Join(tmpDir, "doc/readme.md"),
		Freshness: &document.Freshness{
			LastReviewed: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			Strategy:     StrategyCodeChanges,
			Watch:        []string{"src/**/*.go"},
		},
	}

	checker := NewChecker(gitClient, tmpDir, nil)
	result := checker.Check(doc)

	if result.Status != StatusStale {
		t.Fatalf("Status = %v, want %v", result.Status, StatusStale)
	}
	if len(result.ChangedFiles) != 2 {
		t.Errorf("ChangedFiles = %v, want 2 entries", result.ChangedFiles)
	}
	if !containsStr(result.ChangedFiles, "src/main.go") {
		t.Errorf("ChangedFiles should contain src/main.go, got: %v", result.ChangedFiles)
	}
	if !containsStr(result.ChangedFiles, "src/lib.go") {
		t.Errorf("ChangedFiles should contain src/lib.go, got: %v", result.ChangedFiles)
	}
}

func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
