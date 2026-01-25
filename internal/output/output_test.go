package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/betterment/docrot/internal/freshness"
)

func TestTextFormatter_SingleResult(t *testing.T) {
	results := []freshness.Result{
		{
			Path:         "doc/readme.md",
			Status:       freshness.StatusFresh,
			Strategy:     "interval",
			LastReviewed: "2024-01-15",
			Expires:      "2024-07-15",
		},
	}

	var buf bytes.Buffer
	f := NewTextFormatter(&buf, false)
	f.Write(results)

	output := buf.String()
	if !strings.Contains(output, "doc/readme.md") {
		t.Errorf("Output should contain file path, got: %s", output)
	}
	if !strings.Contains(output, "fresh") {
		t.Errorf("Output should contain status, got: %s", output)
	}
}

func TestTextFormatter_QuietMode(t *testing.T) {
	results := []freshness.Result{
		{Path: "doc/fresh.md", Status: freshness.StatusFresh},
		{Path: "doc/stale.md", Status: freshness.StatusStale, Reason: "Expired"},
	}

	var buf bytes.Buffer
	f := NewTextFormatter(&buf, true) // quiet mode
	f.Write(results)

	output := buf.String()
	if strings.Contains(output, "fresh.md") {
		t.Errorf("Quiet mode should not show fresh docs, got: %s", output)
	}
	if !strings.Contains(output, "stale.md") {
		t.Errorf("Quiet mode should show stale docs, got: %s", output)
	}
}

func TestJSONFormatter(t *testing.T) {
	results := []freshness.Result{
		{
			Path:         "doc/readme.md",
			Status:       freshness.StatusFresh,
			Strategy:     "interval",
			LastReviewed: "2024-01-15",
		},
		{
			Path:         "doc/api.md",
			Status:       freshness.StatusStale,
			Strategy:     "until_date",
			LastReviewed: "2024-01-01",
			Reason:       "Expired",
		},
	}

	var buf bytes.Buffer
	f := NewJSONFormatter(&buf)
	f.Write(results)

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if output.Summary.Total != 2 {
		t.Errorf("Summary.Total = %d, want 2", output.Summary.Total)
	}
	if output.Summary.Fresh != 1 {
		t.Errorf("Summary.Fresh = %d, want 1", output.Summary.Fresh)
	}
	if output.Summary.Stale != 1 {
		t.Errorf("Summary.Stale = %d, want 1", output.Summary.Stale)
	}
	if len(output.Docs) != 2 {
		t.Errorf("Docs length = %d, want 2", len(output.Docs))
	}
}
