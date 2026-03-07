package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestScan_FindsMatchingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create doc structure
	docDir := filepath.Join(tmpDir, "project", "doc")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create some markdown files
	files := []string{
		filepath.Join(docDir, "readme.md"),
		filepath.Join(docDir, "api.md"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("# Doc"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	s := New(tmpDir, []string{"**/doc/**/*.md"}, nil)
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Scan() found %d files, want 2", len(results))
	}
}

func TestScan_RespectsExcludePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create docs in regular location
	docDir := filepath.Join(tmpDir, "project", "doc")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docDir, "api.md"), []byte("# Doc"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create docs in node_modules (should be excluded)
	nodeDocDir := filepath.Join(tmpDir, "node_modules", "pkg", "doc")
	if err := os.MkdirAll(nodeDocDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeDocDir, "readme.md"), []byte("# Doc"), 0644); err != nil {
		t.Fatal(err)
	}

	s := New(tmpDir, []string{"**/doc/**/*.md"}, []string{"**/node_modules/**"})
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Scan() found %d files, want 1 (node_modules should be excluded)", len(results))
	}
}

func TestScan_MultiplePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create doc/ and docs/ directories
	docDir := filepath.Join(tmpDir, "project", "doc")
	docsDir := filepath.Join(tmpDir, "project", "docs")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(docDir, "api.md"), []byte("# Doc"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "guide.md"), []byte("# Guide"), 0644); err != nil {
		t.Fatal(err)
	}

	s := New(tmpDir, []string{"**/doc/**/*.md", "**/docs/**/*.md"}, nil)
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Scan() found %d files, want 2", len(results))
	}
}

func TestScan_NoMatchesReturnsEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	s := New(tmpDir, []string{"**/doc/**/*.md"}, nil)
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Scan() found %d files, want 0", len(results))
	}
}

func TestScan_CustomPatternNonStandardDir(t *testing.T) {
	tmpDir := t.TempDir()

	rfcsDir := filepath.Join(tmpDir, "rfcs", "2024")
	if err := os.MkdirAll(rfcsDir, 0755); err != nil {
		t.Fatal(err)
	}

	files := []string{
		filepath.Join(rfcsDir, "rfc-001.md"),
		filepath.Join(rfcsDir, "rfc-002.md"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("# RFC"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	s := New(tmpDir, []string{"rfcs/**/*.md"}, nil)
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Scan() found %d files, want 2", len(results))
	}
}

func TestScan_WildcardPattern(t *testing.T) {
	tmpDir := t.TempDir()

	dirs := []string{
		filepath.Join(tmpDir, "alpha"),
		filepath.Join(tmpDir, "beta", "sub"),
		filepath.Join(tmpDir, "gamma"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	files := []string{
		filepath.Join(tmpDir, "alpha", "a.md"),
		filepath.Join(tmpDir, "beta", "sub", "b.md"),
		filepath.Join(tmpDir, "gamma", "c.md"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("# Doc"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	s := New(tmpDir, []string{"**/*.md"}, nil)
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Scan() found %d files, want 3", len(results))
	}
}

func TestScan_RespectsGitIgnore(t *testing.T) {
	tmpDir := t.TempDir()

	// Init a git repo
	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}

	// Create two doc dirs: one gitignored, one not
	normalDir := filepath.Join(tmpDir, "docs")
	generatedDir := filepath.Join(tmpDir, "generated", "docs")
	os.MkdirAll(normalDir, 0755)
	os.MkdirAll(generatedDir, 0755)

	os.WriteFile(filepath.Join(normalDir, "guide.md"), []byte("# Guide"), 0644)
	os.WriteFile(filepath.Join(generatedDir, "api.md"), []byte("# API"), 0644)

	// .gitignore that ignores the generated/ directory
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("generated/\n"), 0644)

	s := New(tmpDir, []string{"**/*.md"}, nil)
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Scan() found %d files, want 1 (generated/ should be excluded by .gitignore): %v", len(results), results)
	}
	if len(results) == 1 && !filepath.IsAbs(results[0]) || (len(results) == 1 && filepath.Base(results[0]) != "guide.md") {
		t.Errorf("Scan() should return guide.md, got: %v", results)
	}
}

func TestScan_GitRepo_NothingIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	// Init a git repo with no .gitignore
	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}

	docDir := filepath.Join(tmpDir, "docs")
	os.MkdirAll(docDir, 0755)
	os.WriteFile(filepath.Join(docDir, "guide.md"), []byte("# Guide"), 0644)
	os.WriteFile(filepath.Join(docDir, "api.md"), []byte("# API"), 0644)

	s := New(tmpDir, []string{"**/*.md"}, nil)
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// No .gitignore so all docs should be found
	if len(results) != 2 {
		t.Errorf("Scan() found %d files, want 2 (no .gitignore, nothing should be excluded): %v", len(results), results)
	}
}

func TestScan_RespectsGitIgnore_FileLevel(t *testing.T) {
	tmpDir := t.TempDir()

	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}

	docDir := filepath.Join(tmpDir, "docs")
	os.MkdirAll(docDir, 0755)
	os.WriteFile(filepath.Join(docDir, "guide.md"), []byte("# Guide"), 0644)
	os.WriteFile(filepath.Join(docDir, "generated.md"), []byte("# Generated"), 0644)

	// .gitignore that ignores a specific file
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("docs/generated.md\n"), 0644)

	s := New(tmpDir, []string{"**/*.md"}, nil)
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Scan() found %d files, want 1 (generated.md should be excluded): %v", len(results), results)
	}
	if len(results) == 1 && filepath.Base(results[0]) != "guide.md" {
		t.Errorf("Scan() should return guide.md, got: %v", results)
	}
}

func TestScan_BuildDirNoLongerBlocked(t *testing.T) {
	tmpDir := t.TempDir()

	buildDocsDir := filepath.Join(tmpDir, "build", "docs")
	if err := os.MkdirAll(buildDocsDir, 0755); err != nil {
		t.Fatal(err)
	}

	docPath := filepath.Join(buildDocsDir, "guide.md")
	if err := os.WriteFile(docPath, []byte("# Guide"), 0644); err != nil {
		t.Fatal(err)
	}

	s := New(tmpDir, []string{"build/docs/**/*.md"}, nil)
	results, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Scan() found %d files, want 1 (build/ should no longer be blocked)", len(results))
	}
}
