package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "mathrock",
	Short: "Individual Mathrock Node",
}

// Execute executes the root command.
func Execute() error {
	rootCmd.AddCommand(nodeCmd)
	return rootCmd.Execute()
}
