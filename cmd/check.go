package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

// ErrMissingFrontmatterFound is returned when docs missing frontmatter are found in strict mode
var ErrMissingFrontmatterFound = errors.New("documentation missing frontmatter")

// ErrMultipleRepos is returned when paths span multiple git repositories
var ErrMultipleRepos = errors.New("paths are in different git repositories")

var quiet bool
var strictMode bool

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
	checkCmd.Flags().BoolVar(&strictMode, "strict", false, "Exit with error if any docs are missing frontmatter")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	effectivePolicy := cfg.OnMissingFrontmatter
	if strictMode {
		effectivePolicy = "strict"
	}

	var paths []string
	var inputPaths []string // Track original input paths for git root validation

	if len(args) > 0 {
		inputPaths = args
		// Check if args are files or directories
		for _, arg := range args {
			info, err := os.Stat(arg)
			if err != nil {
				return fmt.Errorf("failed to stat %s: %w", arg, err)
			}
			if info.IsDir() {
				// Scan directory for docs
				s := scanner.New(arg, getPatterns(cfg.Patterns), cfg.Exclude)
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
		inputPaths = []string{"."}
		// Default: scan current directory
		s := scanner.New(".", getPatterns(cfg.Patterns), cfg.Exclude)
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

	// Find git root from input paths, ensuring all are in the same repo
	repoRoot, err := findCommonGitRoot(inputPaths)
	if err != nil {
		return err
	}

	// Set up git client using the derived repo root
	var gitClient *git.Client
	if repoRoot != "" {
		gitClient, _ = git.New(repoRoot)
	}

	// Check docs in parallel
	numWorkers := getWorkers(cfg.Workers)
	rawResults := checker.Run(paths, gitClient, numWorkers, defaultsFromConfig(cfg))

	// Post-process results based on config
	var results []freshness.Result
	var missingCount int
	for _, result := range rawResults {
		if result.Status == freshness.StatusMissingFrontmatter {
			switch effectivePolicy {
			case "skip":
				continue
			case "fail":
				result.Status = freshness.StatusStale
			default:
				// "strict" and "warn": pass through with StatusMissingFrontmatter
				missingCount++
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

	if effectivePolicy == "strict" && missingCount > 0 {
		fmt.Fprintf(os.Stderr, "Error: %d file(s) missing frontmatter (strict mode)\n", missingCount)
		return ErrMissingFrontmatterFound
	}

	return nil
}

// findCommonGitRoot finds the git repository root for the given paths.
// Returns an error if paths are in different git repositories.
// Returns empty string if no paths are in a git repository.
func findCommonGitRoot(paths []string) (string, error) {
	var commonRoot string

	for _, p := range paths {
		// Convert to absolute path
		absPath, err := filepath.Abs(p)
		if err != nil {
			continue
		}

		// For files, use parent directory for git root lookup
		info, err := os.Stat(absPath)
		if err != nil {
			continue
		}
		lookupPath := absPath
		if !info.IsDir() {
			lookupPath = filepath.Dir(absPath)
		}

		// Find git root for this path
		root, err := git.FindRepoRoot(lookupPath)
		if err != nil {
			// Not in a git repo, skip
			continue
		}

		if commonRoot == "" {
			commonRoot = root
		} else if commonRoot != root {
			return "", fmt.Errorf("%w: %s and %s", ErrMultipleRepos, commonRoot, root)
		}
	}

	return commonRoot, nil
}
