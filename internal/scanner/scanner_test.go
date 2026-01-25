package scanner

import (
	"os"
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
