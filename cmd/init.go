package cmd

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new project (alias of new)",
	Args:  cobra.ExactArgs(1),
	RunE:  runInitProject,
}

func init() {
	rootCmd.AddCommand(initCmd)
	bindProjectFlags(initCmd)
}
