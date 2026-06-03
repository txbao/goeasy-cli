package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade templates or framework dependency",
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.AddCommand(upgradeTemplateCmd)
	upgradeCmd.AddCommand(upgradeFrameworkCmd)
}

var upgradeTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Show embedded template upgrade guidance",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Embedded scaffold templates ship with goeasy-cli.")
		fmt.Println("Re-install CLI: go install github.com/txbao/goeasy-cli@latest")
		fmt.Println("Manual merge recommended for customized project files.")
	},
}

var upgradeFrameworkCmd = &cobra.Command{
	Use:   "framework",
	Short: "Show goeasy module version in go.mod",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}
		gomod := filepath.Join(dir, "go.mod")
		b, err := os.ReadFile(gomod)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(b), "\n") {
			if strings.Contains(line, "github.com/txbao/goeasy") {
				fmt.Println(strings.TrimSpace(line))
				fmt.Println("Update version in go.mod, then: go mod tidy")
				return nil
			}
		}
		fmt.Println("goeasy module not found in go.mod (optional for generated projects)")
		return nil
	},
}
