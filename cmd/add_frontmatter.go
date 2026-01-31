package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/andimrob/docrot/internal/config"
	"github.com/andimrob/docrot/internal/document"
	"github.com/andimrob/docrot/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	addFrontmatterStrategy string
	addFrontmatterInterval string
	addFrontmatterDryRun   bool
)

var addFrontmatterCmd = &cobra.Command{
	Use:   "add-frontmatter [paths...]",
	Short: "Add freshness frontmatter to docs missing it",
	Long:  `Add freshness frontmatter to documentation files that don't have it.`,
	RunE:  runAddFrontmatter,
}

func init() {
	addFrontmatterCmd.Flags().StringVarP(&addFrontmatterStrategy, "strategy", "s", "interval", "Default strategy: interval, until_date, code_changes")
	addFrontmatterCmd.Flags().StringVarP(&addFrontmatterInterval, "interval", "i", "180d", "Default interval (for interval strategy)")
	addFrontmatterCmd.Flags().BoolVarP(&addFrontmatterDryRun, "dry-run", "n", false, "Show what would be changed without modifying files")
	rootCmd.AddCommand(addFrontmatterCmd)
}

func runAddFrontmatter(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	s := scanner.New(root, cfg.Patterns, cfg.Exclude)
	paths, err := s.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan for docs: %w", err)
	}

	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "No documentation files found")
		return nil
	}

	today := time.Now().Format("2006-01-02")
	interval := addFrontmatterInterval
	if cfg.Defaults != nil && cfg.Defaults.Interval != "" {
		interval = cfg.Defaults.Interval
	}
	strategy := addFrontmatterStrategy
	if cfg.Defaults != nil && cfg.Defaults.Strategy != "" {
		strategy = cfg.Defaults.Strategy
	}

	for _, path := range paths {
		doc, err := document.Parse(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", path, err)
			continue
		}

		if doc.Freshness != nil {
			// Already has freshness frontmatter
			continue
		}

		if addFrontmatterDryRun {
			fmt.Printf("Would add frontmatter to: %s\n", path)
			continue
		}

		if err := addFrontmatter(path, today, strategy, interval); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating %s: %v\n", path, err)
			continue
		}

		fmt.Printf("Added frontmatter to: %s\n", path)
	}

	return nil
}

func addFrontmatter(path, date, strategy, interval string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	var frontmatterLines []string
	inFrontmatter := false
	hasFrontmatter := false
	frontmatterClosed := false

	for scanner.Scan() {
		line := scanner.Text()

		if !hasFrontmatter && line == "---" && !inFrontmatter {
			inFrontmatter = true
			hasFrontmatter = true
			continue
		}

		if inFrontmatter && line == "---" {
			inFrontmatter = false
			frontmatterClosed = true
			continue
		}

		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Build the docrot block
	freshnessBlock := fmt.Sprintf("docrot:\n  last_reviewed: \"%s\"\n  strategy: %s", date, strategy)
	if strategy == "interval" {
		freshnessBlock += fmt.Sprintf("\n  interval: %s", interval)
	}

	// Build new content
	var newContent strings.Builder
	newContent.WriteString("---\n")

	if hasFrontmatter && frontmatterClosed {
		// Merge with existing frontmatter
		for _, line := range frontmatterLines {
			newContent.WriteString(line)
			newContent.WriteString("\n")
		}
	}

	newContent.WriteString(freshnessBlock)
	newContent.WriteString("\n---\n")

	// Add rest of content
	for _, line := range lines {
		newContent.WriteString(line)
		newContent.WriteString("\n")
	}

	return os.WriteFile(path, []byte(newContent.String()), 0644)
}
