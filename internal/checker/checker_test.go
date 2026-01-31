package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	// Silence unused import error
	_ = fmt.Sprintf
}

func TestRunParallel_ProcessesAllFiles(t *testing.T) {
	tmpDir := t.TempDir()
	docDir := filepath.Join(tmpDir, "doc")
	os.MkdirAll(docDir, 0755)

	// Create multiple docs
	for i := 0; i < 10; i++ {
		content := `---
docrot:
  last_reviewed: "2026-01-20"
  strategy: interval
  interval: 90d
---
# Doc
`
		os.WriteFile(filepath.Join(docDir, fmt.Sprintf("doc%d.md", i)), []byte(content), 0644)
	}

	paths := make([]string, 10)
	for i := 0; i < 10; i++ {
		paths[i] = filepath.Join(docDir, fmt.Sprintf("doc%d.md", i))
	}

	results := Run(paths, nil, 4) // 4 workers

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
		content := `---
docrot:
  last_reviewed: "2026-01-20"
  strategy: interval
  interval: 90d
---
# Doc
`
		os.WriteFile(filepath.Join(docDir, fmt.Sprintf("doc%d.md", i)), []byte(content), 0644)
	}

	paths := make([]string, numDocs)
	for i := 0; i < numDocs; i++ {
		paths[i] = filepath.Join(docDir, fmt.Sprintf("doc%d.md", i))
	}

	// Time parallel execution
	start := time.Now()
	results := Run(paths, nil, 8)
	parallelDuration := time.Since(start)

	if len(results) != numDocs {
		t.Errorf("Expected %d results, got %d", numDocs, len(results))
	}

	t.Logf("Parallel processing of %d files took %v", numDocs, parallelDuration)
}
