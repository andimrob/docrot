package cmd

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

var reviewDate string

var reviewCmd = &cobra.Command{
	Use:   "review <file> [files...]",
	Short: "Update last_reviewed date to today",
	Long:  `Update the last_reviewed date in the frontmatter of one or more documentation files.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runReview,
}

func init() {
	reviewCmd.Flags().StringVarP(&reviewDate, "date", "d", "", "Use specific date instead of today (YYYY-MM-DD)")
	rootCmd.AddCommand(reviewCmd)
}

func runReview(cmd *cobra.Command, args []string) error {
	date := time.Now().Format("2006-01-02")
	if reviewDate != "" {
		// Validate date format
		if _, err := time.Parse("2006-01-02", reviewDate); err != nil {
			return fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", reviewDate)
		}
		date = reviewDate
	}

	for _, path := range args {
		if err := updateLastReviewed(path, date); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating %s: %v\n", path, err)
			continue
		}
		fmt.Printf("Updated %s: last_reviewed = %s\n", path, date)
	}

	return nil
}

func updateLastReviewed(path, date string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Pattern to match last_reviewed in frontmatter
	pattern := regexp.MustCompile(`(last_reviewed:\s*)"?(\d{4}-\d{2}-\d{2})"?`)

	newContent := pattern.ReplaceAllString(string(content), fmt.Sprintf(`${1}"%s"`, date))

	if newContent == string(content) {
		return fmt.Errorf("no last_reviewed field found in frontmatter")
	}

	return os.WriteFile(path, []byte(newContent), 0644)
}
