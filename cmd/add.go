package cmd

import (
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/generator"

	"github.com/spf13/cobra"
)

var (
	addForce      bool
	addProjectDir string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add module, crud, repository, proto, event, or aggregate",
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.PersistentFlags().BoolVar(&addForce, "force", false, "Overwrite existing files")
	addCmd.PersistentFlags().StringVar(&addProjectDir, "dir", ".", "Existing project directory")

	addCmd.AddCommand(&cobra.Command{
		Use:   "module [name]",
		Short: "Generate DDD module skeleton",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateModule(genOpts(args[0]))
		},
	})
	addCmd.AddCommand(&cobra.Command{
		Use:   "crud [name]",
		Short: "Generate CRUD placeholders",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateCRUD(genOpts(args[0]))
		},
	})
	addCmd.AddCommand(&cobra.Command{
		Use:   "repository [name]",
		Short: "Generate repository interface and infra stub",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateRepository(genOpts(args[0]))
		},
	})
	addCmd.AddCommand(&cobra.Command{
		Use:   "proto [name]",
		Short: "Generate api/proto definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateProto(genOpts(args[0]))
		},
	})
	addCmd.AddCommand(&cobra.Command{
		Use:   "event [name]",
		Short: "Generate domain event and publisher stub",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateEvent(genOpts(args[0]))
		},
	})
	addCmd.AddCommand(&cobra.Command{
		Use:   "aggregate [name]",
		Short: "Generate aggregate skeleton",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateAggregate(genOpts(args[0]))
		},
	})
}

func genOpts(name string) generator.ModuleOptions {
	abs, err := filepath.Abs(addProjectDir)
	if err != nil {
		abs = addProjectDir
	}
	return generator.ModuleOptions{
		ProjectDir: abs,
		ModuleName: name,
		Force:      addForce,
	}
}
