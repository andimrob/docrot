package cmd

import (
	"os"

	"github.com/andimrob/docrot/internal/config"
	"github.com/andimrob/docrot/internal/freshness"
	"github.com/spf13/cobra"
)

var (
	configPath  string
	format      string
	workers     int
	patternFlag []string
)

var rootCmd = &cobra.Command{
	Use:   "docrot",
	Short: "Detect stale documentation",
	Long: `docrot checks documentation files for staleness based on frontmatter configuration.

It supports multiple freshness strategies:
  - interval: Doc expires after a specified duration since last review
  - until_date: Doc expires on a specific date
  - code_changes: Doc expires when related code files change`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".docrot.yml", "Path to config file")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "text", "Output format: text, json")
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "w", 0, "Number of parallel workers (0 = use CPU count)")
	rootCmd.PersistentFlags().StringArrayVarP(&patternFlag, "pattern", "p", nil, "Glob pattern(s) for doc discovery (overrides config; may be repeated)")
}

// getWorkers returns the number of workers to use.
// CLI flag takes precedence over config file. 0 means use CPU count.
func getWorkers(configWorkers int) int {
	if workers > 0 {
		return workers
	}
	return configWorkers
}

// getPatterns returns the patterns to use for doc discovery.
// CLI flag takes precedence over config file.
func getPatterns(cfgPatterns []string) []string {
	if len(patternFlag) > 0 {
		return patternFlag
	}
	return cfgPatterns
}

// defaultsFromConfig extracts watch/ignore defaults from config into a freshness.DefaultPatterns.
func defaultsFromConfig(cfg *config.Config) *freshness.DefaultPatterns {
	if cfg.Defaults != nil && (len(cfg.Defaults.Watch) > 0 || len(cfg.Defaults.Ignore) > 0) {
		return &freshness.DefaultPatterns{
			Watch:  cfg.Defaults.Watch,
			Ignore: cfg.Defaults.Ignore,
		}
	}
	return nil
}
