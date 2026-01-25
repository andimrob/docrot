package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set during build via ldflags
	Version = "dev"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("docrot %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
