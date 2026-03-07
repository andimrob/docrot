package cmd

import (
	"fmt"
	"os"

	"github.com/andimrob/docrot/internal/checker"
	"github.com/andimrob/docrot/internal/config"
	"github.com/andimrob/docrot/internal/git"
	"github.com/andimrob/docrot/internal/output"
	"github.com/andimrob/docrot/internal/scanner"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [paths...]",
	Short: "List all docs and their freshness status",
	Long:  `List all documentation files and their current freshness status.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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

	gitClient, _ := git.New(root)

	// Check docs in parallel
	numWorkers := getWorkers(cfg.Workers)
	results := checker.Run(paths, gitClient, numWorkers, defaultsFromConfig(cfg))

	var formatter output.Formatter
	switch format {
	case "json":
		formatter = output.NewJSONFormatter(os.Stdout)
	default:
		formatter = output.NewTextFormatter(os.Stdout, false)
	}

	formatter.Write(results)
	return nil
}
