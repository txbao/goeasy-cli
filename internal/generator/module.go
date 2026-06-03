package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ModuleOptions struct {
	ProjectDir string
	ModuleName string
	Force      bool
}

func GenerateModule(opts ModuleOptions) error {
	if opts.ModuleName == "" {
		return fmt.Errorf("module name is required")
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	repl := map[string]string{
		"MODULE": data["ModuleSnake"].(string),
	}
	return renderScoped("module", opts.ProjectDir, repl, data, opts.Force)
}

func GenerateCRUD(opts ModuleOptions) error {
	if err := GenerateModule(opts); err != nil {
		return err
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	repl := map[string]string{"MODULE": data["ModuleSnake"].(string)}
	return renderScoped("crud", opts.ProjectDir, repl, data, true)
}

func GenerateRepository(opts ModuleOptions) error {
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	repl := map[string]string{"MODULE": data["ModuleSnake"].(string)}
	return renderScoped("repository", opts.ProjectDir, repl, data, opts.Force)
}

func GenerateProto(opts ModuleOptions) error {
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	repl := map[string]string{"MODULE": data["ModuleSnake"].(string)}
	return renderScoped("proto", opts.ProjectDir, repl, data, opts.Force)
}

func GenerateEvent(opts ModuleOptions) error {
	if opts.ModuleName == "" {
		return fmt.Errorf("event name is required")
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := toMap(BuildEventData(opts.ModuleName, projectModule))
	repl := map[string]string{"EVENT": data["EventSnake"].(string)}
	return renderScoped("event", opts.ProjectDir, repl, data, opts.Force)
}

func GenerateAggregate(opts ModuleOptions) error {
	if opts.ModuleName == "" {
		return fmt.Errorf("aggregate name is required")
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	repl := map[string]string{"MODULE": data["ModuleSnake"].(string)}
	return renderScoped("aggregate", opts.ProjectDir, repl, data, opts.Force)
}

func renderScoped(templateRoot, projectDir string, repl map[string]string, data map[string]any, force bool) error {
	sub, err := fsSub(templateRoot)
	if err != nil {
		return err
	}
	return fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel := applyPathReplacements(path, repl)
		targetRel := rel
		if strings.HasSuffix(targetRel, ".tmpl") {
			targetRel = targetRel[:len(targetRel)-5]
		}
		targetPath := filepath.Join(projectDir, targetRel)
		if !force {
			if _, err := os.Stat(targetPath); err == nil {
				return fmt.Errorf("file exists: %s (use --force)", targetRel)
			}
		}
		content, err := fs.ReadFile(sub, path)
		if err != nil {
			return err
		}
		var out []byte
		if strings.HasSuffix(path, ".tmpl") {
			out, err = executeTemplate(d.Name(), content, data)
			if err != nil {
				return err
			}
		} else {
			out = content
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, out, 0644); err != nil {
			return err
		}
		fmt.Printf("  created %s\n", targetRel)
		return nil
	})
}

func readModulePath(projectDir string) (string, error) {
	gomod := filepath.Join(projectDir, "go.mod")
	b, err := os.ReadFile(gomod)
	if err != nil {
		return "", fmt.Errorf("read go.mod in %s: %w", projectDir, err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module path not found in go.mod")
}
