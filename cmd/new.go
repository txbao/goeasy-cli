package cmd

import "github.com/spf13/cobra"

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new DDD Lite project",
	Args:  cobra.ExactArgs(1),
	RunE:  runInitProject,
}

func init() {
	rootCmd.AddCommand(newCmd)
	bindProjectFlags(newCmd)
}
