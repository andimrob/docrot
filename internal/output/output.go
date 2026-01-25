package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/betterment/docrot/internal/freshness"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	Write(results []freshness.Result)
}

// TextFormatter outputs human-readable text
type TextFormatter struct {
	w     io.Writer
	quiet bool
}

func NewTextFormatter(w io.Writer, quiet bool) *TextFormatter {
	return &TextFormatter{w: w, quiet: quiet}
}

func (f *TextFormatter) Write(results []freshness.Result) {
	var fresh, stale, missing int

	for _, r := range results {
		switch r.Status {
		case freshness.StatusFresh:
			fresh++
			if f.quiet {
				continue
			}
		case freshness.StatusStale:
			stale++
		case freshness.StatusMissingFrontmatter:
			missing++
		}

		f.writeResult(r)
	}

	if !f.quiet {
		fmt.Fprintf(f.w, "\nSummary: %d fresh, %d stale, %d missing frontmatter\n",
			fresh, stale, missing)
	}
}

func (f *TextFormatter) writeResult(r freshness.Result) {
	statusIcon := "✓"
	switch r.Status {
	case freshness.StatusStale:
		statusIcon = "✗"
	case freshness.StatusMissingFrontmatter:
		statusIcon = "?"
	}

	fmt.Fprintf(f.w, "%s %s [%s]\n", statusIcon, r.Path, r.Status)

	if r.Reason != "" {
		fmt.Fprintf(f.w, "  └─ %s\n", r.Reason)
	}
}

// JSONFormatter outputs machine-readable JSON
type JSONFormatter struct {
	w io.Writer
}

type JSONOutput struct {
	Summary Summary            `json:"summary"`
	Docs    []freshness.Result `json:"docs"`
}

type Summary struct {
	Total              int `json:"total"`
	Fresh              int `json:"fresh"`
	Stale              int `json:"stale"`
	MissingFrontmatter int `json:"missing_frontmatter"`
}

func NewJSONFormatter(w io.Writer) *JSONFormatter {
	return &JSONFormatter{w: w}
}

func (f *JSONFormatter) Write(results []freshness.Result) {
	output := JSONOutput{
		Docs: results,
	}

	for _, r := range results {
		output.Summary.Total++
		switch r.Status {
		case freshness.StatusFresh:
			output.Summary.Fresh++
		case freshness.StatusStale:
			output.Summary.Stale++
		case freshness.StatusMissingFrontmatter:
			output.Summary.MissingFrontmatter++
		}
	}

	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(output) // Error unlikely for stdout; silently ignore
}
