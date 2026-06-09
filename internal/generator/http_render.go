package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func clientNames(clients []ClientSurface) []string {
	out := make([]string, len(clients))
	for i, c := range clients {
		out[i] = c.Name
	}
	return out
}

func renderScopedFiltered(templateRoot, projectDir string, repl map[string]string, data map[string]any, force bool, includeSubstr string, excludeSubstrs []string, skipTargets ...string) error {
	skip := make(map[string]bool, len(skipTargets))
	for _, t := range skipTargets {
		skip[filepath.ToSlash(t)] = true
	}
	sub, err := fsSub(templateRoot)
	if err != nil {
		return err
	}
	return fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." || d.IsDir() {
			return nil
		}
		if strings.Contains(path, "register_MODULE") || strings.Contains(path, "register_DOMAIN") {
			return nil
		}
		tplPath := filepath.ToSlash(path)
		if includeSubstr != "" && !strings.Contains(tplPath, includeSubstr) {
			return nil
		}
		for _, excludeSubstr := range excludeSubstrs {
			if excludeSubstr != "" && strings.Contains(tplPath, excludeSubstr) {
				return nil
			}
		}
		rel := applyPathReplacements(path, repl)
		targetRel := rel
		if strings.HasSuffix(targetRel, ".tmpl") {
			targetRel = targetRel[:len(targetRel)-5]
		}
		targetRel = filepath.ToSlash(targetRel)
		if skip[targetRel] {
			return nil
		}
		targetPath := filepath.Join(projectDir, targetRel)
		if !force {
			if _, err := os.Stat(targetPath); err == nil {
				fmt.Fprintf(os.Stderr, "info: skip existing %s (use --force)\n", targetRel)
				return nil
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

func renderHTTPClientLayer(templateRoot, projectDir, moduleSnake, projectModule string, clients []ClientSurface, baseData map[string]any, force bool) error {
	meta := metaFromData(baseData)
	if meta.ModuleID == "" {
		meta.ModuleID = moduleSnake
	}
	for _, cl := range clients {
		data := copyMap(baseData)
		data["HTTPClient"] = cl.Name
		data["ClientFullCRUD"] = cl.FullCRUD
		repl, include := httpTemplateRepl(cl.Name, meta)
		if err := renderScopedFiltered(templateRoot, projectDir, repl, data, force, include, []string{"handler.go.tmpl"}); err != nil {
			return err
		}
	}
	return nil
}

func renderHTTPCrudOverlay(projectDir, moduleSnake, projectModule string, clients []ClientSurface, baseData map[string]any, force bool) error {
	meta := metaFromData(baseData)
	if meta.ModuleID == "" {
		meta.ModuleID = moduleSnake
	}
	for _, cl := range clients {
		data := copyMap(baseData)
		if data == nil {
			data = toMap(BuildModuleData(moduleSnake, projectModule))
			enrichModuleMetaData(data, meta)
		}
		data["HTTPClient"] = cl.Name
		data["ClientFullCRUD"] = cl.FullCRUD
		repl, include := httpTemplateRepl(cl.Name, meta)
		if err := renderScopedFiltered("crud", projectDir, repl, data, force, include, []string{"handler_crud.go.tmpl"}); err != nil {
			if !strings.Contains(err.Error(), "file exists") {
				return err
			}
			fmt.Fprintf(os.Stderr, "info: crud overlay skipped for client %s (files exist)\n", cl.Name)
		}
	}
	return nil
}

func copyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
