package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func ensureDBXPackage(opts ModuleOptions) error {
	marker := filepath.Join(opts.ProjectDir, "internal", "infrastructure", "shared", "dbx", "dialect.go")
	if _, err := os.Stat(marker); err == nil {
		return nil
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := map[string]any{
		"ModuleName":   projectModule,
		"GoEasyModule": currentGoEasyModule(),
	}
	sub, err := fsSub("project")
	if err != nil {
		return err
	}
	return walkDBXTemplates(sub, "internal/infrastructure/shared/dbx", opts, data)
}

func walkDBXTemplates(sub fs.FS, root string, opts ModuleOptions, data map[string]any) error {
	info, err := fs.Stat(sub, root)
	if err != nil || !info.IsDir() {
		return nil
	}
	return fs.WalkDir(sub, root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || path == root {
			return nil
		}
		targetRel := path
		if filepath.Ext(targetRel) == ".tmpl" {
			targetRel = targetRel[:len(targetRel)-5]
		}
		targetPath := filepath.Join(opts.ProjectDir, targetRel)
		if !opts.Force {
			if _, err := os.Stat(targetPath); err == nil {
				return nil
			}
		}
		content, err := fs.ReadFile(sub, path)
		if err != nil {
			return err
		}
		var out []byte
		if filepath.Ext(path) == ".tmpl" {
			out, err = executeTemplate(filepath.Base(path), content, data)
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
		fmt.Printf("  created %s\n", filepath.ToSlash(targetRel))
		return nil
	})
}
