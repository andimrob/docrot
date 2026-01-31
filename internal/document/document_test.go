package document

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse_WithNoFrontmatter(t *testing.T) {
	content := `# My Documentation

This is content without frontmatter.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	doc, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if doc.Freshness != nil {
		t.Errorf("Freshness = %v, want nil", doc.Freshness)
	}
}

func TestParse_WithUntilDateStrategy(t *testing.T) {
	content := `---
docrot:
  last_reviewed: "2024-01-15"
  strategy: until_date
  expires: "2024-06-01"
---
# My Documentation
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	doc, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if doc.Freshness.Strategy != "until_date" {
		t.Errorf("Strategy = %v, want %v", doc.Freshness.Strategy, "until_date")
	}

	if doc.Freshness.Expires != "2024-06-01" {
		t.Errorf("Expires = %v, want %v", doc.Freshness.Expires, "2024-06-01")
	}
}

func TestParse_WithCodeChangesStrategy(t *testing.T) {
	content := `---
docrot:
  last_reviewed: "2024-01-15"
  strategy: code_changes
  watch:
    - "../**/*.rb"
    - "../**/*.go"
---
# My Documentation
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	doc, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if doc.Freshness.Strategy != "code_changes" {
		t.Errorf("Strategy = %v, want %v", doc.Freshness.Strategy, "code_changes")
	}

	if len(doc.Freshness.Watch) != 2 {
		t.Errorf("Watch length = %v, want %v", len(doc.Freshness.Watch), 2)
	}
}

func TestParse_FileNotFound(t *testing.T) {
	_, err := Parse("/nonexistent/path/file.md")
	if err == nil {
		t.Error("Parse() expected error for nonexistent file, got nil")
	}
}

func TestParse_WithValidFrontmatter(t *testing.T) {
	// Create a temp file with frontmatter
	content := `---
docrot:
  last_reviewed: "2024-01-15"
  strategy: interval
  interval: 90d
---
# My Documentation

This is the content.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	doc, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if doc.Path != tmpFile {
		t.Errorf("Path = %v, want %v", doc.Path, tmpFile)
	}

	if doc.Freshness == nil {
		t.Fatal("Freshness is nil, expected parsed frontmatter")
	}

	if doc.Freshness.LastReviewed != "2024-01-15" {
		t.Errorf("LastReviewed = %v, want %v", doc.Freshness.LastReviewed, "2024-01-15")
	}

	if doc.Freshness.Strategy != "interval" {
		t.Errorf("Strategy = %v, want %v", doc.Freshness.Strategy, "interval")
	}

	if doc.Freshness.Interval != "90d" {
		t.Errorf("Interval = %v, want %v", doc.Freshness.Interval, "90d")
	}
}
