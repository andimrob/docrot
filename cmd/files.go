package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/andimrob/docrot/internal/document"
	"github.com/andimrob/docrot/internal/files"
	"github.com/andimrob/docrot/internal/freshness"
	"github.com/andimrob/docrot/internal/git"
	"github.com/spf13/cobra"
)

var filesCmd = &cobra.Command{
	Use:   "files <doc-path> [doc-path...]",
	Short: "List files in a document's domain",
	Long: `List all files that fall within a document's watch patterns.

This is useful for understanding what code a document covers, or for
feeding to an LLM to help generate or update documentation.

The command reads the watch and ignore patterns from the document's
frontmatter. If not specified, smart defaults are computed based on
the document's location (watch parent directory, ignore docs directory).`,
	Args: cobra.MinimumNArgs(1),
	RunE: runFiles,
}

func init() {
	rootCmd.AddCommand(filesCmd)
}

type filesOutput struct {
	Document string   `json:"document"`
	Watch    []string `json:"watch"`
	Ignore   []string `json:"ignore"`
	Files    []string `json:"files"`
}

func runFiles(cmd *cobra.Command, args []string) error {
	var allOutputs []filesOutput

	for _, docPath := range args {
		output, err := getFilesForDoc(docPath)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", docPath, err)
		}
		allOutputs = append(allOutputs, output)
	}

	switch format {
	case "json":
		return outputFilesJSON(allOutputs)
	default:
		return outputFilesText(allOutputs)
	}
}

func getFilesForDoc(docPath string) (filesOutput, error) {
	// Parse document to get watch/ignore patterns
	doc, err := document.Parse(docPath)
	if err != nil {
		return filesOutput{}, fmt.Errorf("failed to parse document: %w", err)
	}

	// Determine the root directory (parent of doc file's directory for pattern matching)
	absDocPath, err := filepath.Abs(docPath)
	if err != nil {
		return filesOutput{}, err
	}

	// Find repo root or use doc's parent directory structure
	root := findRoot(absDocPath)

	// Get watch/ignore patterns
	var watchPatterns, ignorePatterns []string

	if doc.Freshness != nil {
		watchPatterns = doc.Freshness.Watch
		ignorePatterns = doc.Freshness.Ignore
	}

	// If no patterns specified, use smart defaults
	if len(watchPatterns) == 0 && len(ignorePatterns) == 0 {
		watchPatterns, ignorePatterns = freshness.ComputeDefaultPatterns(absDocPath, root)
	} else if len(watchPatterns) == 0 {
		// If only ignore is specified, watch everything
		watchPatterns = []string{"**/*"}
	}

	// List files matching patterns
	matchedFiles, err := files.ListFiles(root, watchPatterns, ignorePatterns)
	if err != nil {
		return filesOutput{}, fmt.Errorf("failed to list files: %w", err)
	}

	// Sort for consistent output
	sort.Strings(matchedFiles)

	return filesOutput{
		Document: docPath,
		Watch:    watchPatterns,
		Ignore:   ignorePatterns,
		Files:    matchedFiles,
	}, nil
}

// findRoot finds the git repo root for a document, falling back to the doc's parent directory.
func findRoot(absDocPath string) string {
	dir := filepath.Dir(absDocPath)
	root, err := git.FindRepoRoot(dir)
	if err != nil {
		// Not in a git repo; fall back to parent of doc's directory
		return filepath.Dir(dir)
	}
	return root
}

func outputFilesJSON(outputs []filesOutput) error {
	result := map[string]interface{}{
		"documents": outputs,
	}

	// For single document, also add top-level files for convenience
	if len(outputs) == 1 {
		result["files"] = outputs[0].Files
		result["watch"] = outputs[0].Watch
		result["ignore"] = outputs[0].Ignore
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func outputFilesText(outputs []filesOutput) error {
	for i, output := range outputs {
		if len(outputs) > 1 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("# %s\n", output.Document)
		}

		for _, f := range output.Files {
			fmt.Println(f)
		}
	}
	return nil
}
