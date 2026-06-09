package cmd

import (
	"github.com/txbao/goeasy-cli/internal/configpath"

	"github.com/spf13/cobra"
)

var (
	addConfigPath string
	addAppStyle   string
)

func bindAddSharedFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&addConfigPath, "config", "f", "", "Config file path (default: configs/config.yaml or GOEASY_CONFIG)")
	cmd.PersistentFlags().StringVar(&addAppStyle, "app-style", "", "Application layer style: service|light_cqrs|full_cqrs (default: service; aliases: light, full)")
}

func resolvedConfigPath(projectDir, flagValue string) string {
	return configpath.Resolve(projectDir, flagValue)
}
