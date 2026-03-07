package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a .docrot.yml config file",
	Long:  `Create a default .docrot.yml configuration file in the current directory.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configFile := ".docrot.yml"

	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil {
		return errors.New("config file already exists: " + configFile)
	}

	config := `# docrot configuration
# See: https://github.com/andimrob/docrot

patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"

exclude:
  - "**/node_modules/**"
  - "**/vendor/**"

# What to do when a doc has no freshness frontmatter: warn, skip, fail, or strict
on_missing_frontmatter: warn

# Default freshness settings for add-frontmatter command
defaults:
  strategy: interval
  interval: 180d
  # watch patterns for code_changes strategy:
  # watch:
  #   - "**/*.go"
  #   - "**/*.rb"
`

	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Created %s\n", configFile)
	return nil
}
