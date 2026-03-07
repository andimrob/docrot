package files

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestListFiles_WithWatchPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "lib"), 0755)

	// Create files
	os.WriteFile(filepath.Join(tmpDir, "src/main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src/util.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "lib/helper.rb"), []byte("class Helper"), 0644)

	files, err := ListFiles(tmpDir, []string{"src/**/*.go"}, nil)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d: %v", len(files), files)
	}

	// Check that both .go files are included
	sort.Strings(files)
	if len(files) >= 2 {
		if !contains(files, "src/main.go") {
			t.Errorf("Expected src/main.go in results: %v", files)
		}
		if !contains(files, "src/util.go") {
			t.Errorf("Expected src/util.go in results: %v", files)
		}
	}
}

func TestListFiles_WithIgnorePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "docs"), 0755)

	os.WriteFile(filepath.Join(tmpDir, "src/main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "docs/readme.md"), []byte("# Readme"), 0644)

	// Watch everything, ignore docs
	files, err := ListFiles(tmpDir, []string{"**/*"}, []string{"docs/**"})
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	// Should include main.go but not readme.md
	if !contains(files, "src/main.go") {
		t.Errorf("Expected src/main.go in results: %v", files)
	}
	if contains(files, "docs/readme.md") {
		t.Errorf("docs/readme.md should be ignored: %v", files)
	}
}

func TestListFiles_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "src/main.go"), []byte("package main"), 0644)

	// Pattern that matches nothing
	files, err := ListFiles(tmpDir, []string{"**/*.rb"}, nil)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files, got %d: %v", len(files), files)
	}
}

func TestListFiles_MultipleWatchPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "lib"), 0755)

	os.WriteFile(filepath.Join(tmpDir, "src/main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "lib/helper.rb"), []byte("class Helper"), 0644)

	// Multiple watch patterns
	files, err := ListFiles(tmpDir, []string{"**/*.go", "**/*.rb"}, nil)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d: %v", len(files), files)
	}
}

func TestListFiles_ExcludedDirectoryIsPruned(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a deeply nested file inside an excluded directory
	os.MkdirAll(filepath.Join(tmpDir, "ignored_dir/nested/deep"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ignored_dir/nested/deep/file.go"), []byte("package x"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src/main.go"), []byte("package main"), 0644)

	files, err := ListFiles(tmpDir, []string{"**/*"}, []string{"ignored_dir/**"})
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if !contains(files, "src/main.go") {
		t.Errorf("Expected src/main.go in results: %v", files)
	}
	if contains(files, "ignored_dir/nested/deep/file.go") {
		t.Errorf("ignored_dir/nested/deep/file.go should be excluded: %v", files)
	}
}

func TestListFiles_EmptyWatchPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "src/main.go"), []byte("package main"), 0644)

	files, err := ListFiles(tmpDir, []string{}, nil)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files with empty watch patterns, got %d: %v", len(files), files)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
