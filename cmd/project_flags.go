package cmd

import (
	"github.com/txbao/goeasy-cli/internal/generator"

	"github.com/spf13/cobra"
)

var (
	moduleName      string
	templateName    string
	templateVersion string
	useDownload     bool
	outputDir       string
	goeasyReplace   string
)

func bindProjectFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&moduleName, "module", "", "Go module name")
	cmd.Flags().StringVar(&templateName, "template", "default", "Template: default|monolith|auth|system|payment")
	cmd.Flags().StringVar(&templateVersion, "version", "v1.0.0", "Remote template version")
	cmd.Flags().BoolVar(&useDownload, "download", false, "Download template from remote (fallback to embed)")
	cmd.Flags().StringVar(&outputDir, "output", ".", "Output parent directory")
	cmd.Flags().StringVar(&goeasyReplace, "goeasy-replace", "", "replace path for local goeasy module")
}

func runInitProject(cmd *cobra.Command, args []string) error {
	projectName := args[0]
	parent := outputDir
	return generator.GenerateProject(generator.Options{
		ProjectName:     projectName,
		ModuleName:      moduleName,
		TemplateName:    templateName,
		TemplateVersion: templateVersion,
		UseDownload:     useDownload,
		OutputDir:       parent,
		ZdgfReplace:     goeasyReplace,
	})
}
