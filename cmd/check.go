package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/andimrob/docrot/internal/checker"
	"github.com/andimrob/docrot/internal/config"
	"github.com/andimrob/docrot/internal/freshness"
	"github.com/andimrob/docrot/internal/git"
	"github.com/andimrob/docrot/internal/output"
	"github.com/andimrob/docrot/internal/scanner"
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

	var paths []string

	if len(args) > 0 {
		// Check if args are files or directories
		for _, arg := range args {
			info, err := os.Stat(arg)
			if err != nil {
				return fmt.Errorf("failed to stat %s: %w", arg, err)
			}
			if info.IsDir() {
				// Scan directory for docs
				s := scanner.New(arg, cfg.Patterns, cfg.Exclude)
				dirPaths, err := s.Scan()
				if err != nil {
					return fmt.Errorf("failed to scan for docs: %w", err)
				}
				paths = append(paths, dirPaths...)
			} else {
				// Use file directly
				paths = append(paths, arg)
			}
		}
	} else {
		// Default: scan current directory
		s := scanner.New(".", cfg.Patterns, cfg.Exclude)
		var err error
		paths, err = s.Scan()
		if err != nil {
			return fmt.Errorf("failed to scan for docs: %w", err)
		}
	}

	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "No documentation files found")
		return nil
	}

	// Set up git client
	gitClient, _ := git.New(".")

	// Convert config defaults to freshness defaults
	var defaults *freshness.DefaultPatterns
	if cfg.Defaults != nil && (len(cfg.Defaults.Watch) > 0 || len(cfg.Defaults.Ignore) > 0) {
		defaults = &freshness.DefaultPatterns{
			Watch:  cfg.Defaults.Watch,
			Ignore: cfg.Defaults.Ignore,
		}
	}

	// Check docs in parallel
	numWorkers := getWorkers(cfg.Workers)
	rawResults := checker.Run(paths, gitClient, numWorkers, defaults)

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
