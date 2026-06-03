package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

type Options struct {
	ProjectName     string
	ModuleName      string
	ServiceName     string
	TemplateName    string
	TemplateVersion string
	UseDownload     bool
	OutputDir       string
	ZdgfReplace     string
}

func GenerateProject(opts Options) error {
	if opts.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}
	if opts.ModuleName == "" {
		opts.ModuleName = opts.ProjectName
		fmt.Fprintf(os.Stderr, "warn: --module not set, using %q; prefer github.com/org/%s\n",
			opts.ModuleName, opts.ProjectName)
	}
	if opts.TemplateName == "" {
		opts.TemplateName = "default"
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}

	targetDir := filepath.Join(opts.OutputDir, opts.ProjectName)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	opts.OutputDir = targetDir

	data := toMap(BuildProjectData(opts))
	root := resolveTemplateRoot(opts.TemplateName)

	var renderErr error
	if opts.UseDownload {
		tmpDir := filepath.Join(os.TempDir(), "goeasy-"+opts.TemplateVersion)
		if err := DownloadTemplate(opts.TemplateName, opts.TemplateVersion, tmpDir); err == nil {
			renderErr = RenderDir(tmpDir, targetDir, data)
		} else {
			fmt.Fprintf(os.Stderr, "warn: remote template failed (%v), fallback to embed\n", err)
			renderErr = renderProjectTemplates(opts, targetDir, data, root)
		}
	} else {
		renderErr = renderProjectTemplates(opts, targetDir, data, root)
	}
	if renderErr != nil {
		return renderErr
	}

	fmt.Printf("project %s created at %s\n", opts.ProjectName, targetDir)
	return nil
}

func renderProjectTemplates(opts Options, targetDir string, data map[string]any, root string) error {
	if err := RenderEmbeddedTemplates("project", targetDir, data); err != nil {
		return err
	}
	overlay := opts.TemplateName
	if overlay == "default" || overlay == "project" {
		return nil
	}
	repl := map[string]string{"MODULE": ""}
	return RenderEmbeddedOverlay(overlay, targetDir, data, repl)
}
