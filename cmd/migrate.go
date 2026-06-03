package cmd

import (
	"path/filepath"
	"strconv"

	"github.com/txbao/goeasy-cli/internal/migrate"

	"github.com/spf13/cobra"
)

var (
	migrateDir        string
	migrateConfig     string
	migrateMigrations string
	migrateSteps      int
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database SQL migrations (_sqlx_migrations)",
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the latest migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return migrate.Down(migrateOpts())
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.PersistentFlags().StringVar(&migrateDir, "dir", ".", "Project root directory")
	migrateCmd.PersistentFlags().StringVarP(&migrateConfig, "config", "f", "configs/config.yaml", "Config file path")
	migrateCmd.PersistentFlags().StringVar(&migrateMigrations, "migrations", "migrations", "Migrations directory")

	migrateCmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Apply pending .up.sql migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return migrate.Up(migrateOpts())
		},
	})
	migrateDownCmd.Flags().IntVar(&migrateSteps, "steps", 1, "Number of migrations to roll back")
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return migrate.Status(migrateOpts())
		},
	})
	migrateCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show current migration version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return migrate.Version(migrateOpts())
		},
	})
	migrateCmd.AddCommand(&cobra.Command{
		Use:   "goto <version>",
		Short: "Migrate to a specific version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			return migrate.Goto(migrateOpts(), uint(target))
		},
	})
	migrateCmd.AddCommand(&cobra.Command{
		Use:   "force <version>",
		Short: "Force set migration version without running migrations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			return migrate.Force(migrateOpts(), target)
		},
	})
	migrateCmd.AddCommand(&cobra.Command{
		Use:   "create [name]",
		Short: "Create a new up/down migration pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(migrateDir)
			if err != nil {
				return err
			}
			return migrate.Create(filepath.Join(abs, migrateMigrations), args[0])
		},
	})
}

func migrateOpts() migrate.Options {
	abs, err := filepath.Abs(migrateDir)
	if err != nil {
		abs = migrateDir
	}
	return migrate.Options{
		ProjectDir:    abs,
		ConfigPath:    migrateConfig,
		MigrationsDir: filepath.Join(abs, migrateMigrations),
		Steps:         migrateSteps,
	}
}
