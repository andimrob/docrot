package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/betterment/docrot/internal/checker"
	"github.com/betterment/docrot/internal/config"
	"github.com/betterment/docrot/internal/freshness"
	"github.com/betterment/docrot/internal/git"
	"github.com/betterment/docrot/internal/output"
	"github.com/betterment/docrot/internal/scanner"
	"github.com/spf13/cobra"
)

// ErrStaleDocsFound is returned when stale docs are found
var ErrStaleDocsFound = errors.New("stale documentation found")

var quiet bool

var checkCmd = &cobra.Command{
	Use:           "check [paths...]",
	Short:         "Check documentation freshness",
	Long:          `Check all documentation files for staleness. Exits with code 1 if any docs are stale.`,
	RunE:          runCheck,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	checkCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only output stale docs")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine root directory
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	// Find all docs
	s := scanner.New(root, cfg.Patterns, cfg.Exclude)
	paths, err := s.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan for docs: %w", err)
	}

	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "No documentation files found")
		return nil
	}

	// Set up git client
	gitClient, _ := git.New(root)

	// Check docs in parallel
	numWorkers := getWorkers(cfg.Workers)
	rawResults := checker.Run(paths, gitClient, numWorkers)

	// Post-process results based on config
	var results []freshness.Result
	for _, result := range rawResults {
		if result.Status == freshness.StatusMissingFrontmatter {
			switch cfg.OnMissingFrontmatter {
			case "skip":
				continue
			case "fail":
				result.Status = freshness.StatusStale
			}
		}
		results = append(results, result)
	}

	// Output results
	var formatter output.Formatter
	switch format {
	case "json":
		formatter = output.NewJSONFormatter(os.Stdout)
	default:
		formatter = output.NewTextFormatter(os.Stdout, quiet)
	}

	formatter.Write(results)

	// Return error if any stale
	for _, r := range results {
		if r.Status == freshness.StatusStale {
			return ErrStaleDocsFound
		}
	}

	return nil
}
