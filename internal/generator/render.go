package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	tplfs "github.com/txbao/goeasy-cli/internal/templates"
)

func EmbeddedFS() fs.FS {
	return tplfs.AllTemplates
}

func fsSub(root string) (fs.FS, error) {
	return fs.Sub(EmbeddedFS(), strings.Trim(root, "/"))
}

func RenderEmbeddedTemplates(templateRoot, dstDir string, data map[string]any) error {
	root := strings.Trim(templateRoot, "/")
	sub, err := fs.Sub(EmbeddedFS(), root)
	if err != nil {
		return fmt.Errorf("template %q not found: %w", templateRoot, err)
	}
	return renderFS(sub, "", dstDir, data, nil)
}

// RenderEmbeddedOverlay renders optional overlay files on top of an existing project.
func RenderEmbeddedOverlay(overlayRoot, dstDir string, data map[string]any, pathRepl map[string]string) error {
	root := strings.Trim(overlayRoot, "/")
	sub, err := fs.Sub(EmbeddedFS(), root)
	if err != nil {
		return nil
	}
	return renderFS(sub, "", dstDir, data, pathRepl)
}

func renderFS(tplFS fs.FS, srcBase, dstDir string, data map[string]any, pathRepl map[string]string) error {
	return fs.WalkDir(tplFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		rel := path
		if srcBase != "" {
			rel = strings.TrimPrefix(path, srcBase+"/")
		}
		rel = applyPathReplacements(rel, pathRepl)

		targetRel := rel
		if strings.HasSuffix(targetRel, ".tmpl") {
			targetRel = targetRel[:len(targetRel)-5]
		}

		targetPath := filepath.Join(dstDir, targetRel)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		content, err := fs.ReadFile(tplFS, path)
		if err != nil {
			return err
		}

		if strings.HasSuffix(d.Name(), ".gitkeep") {
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			return os.WriteFile(targetPath, content, 0644)
		}

		if strings.HasSuffix(path, ".tmpl") {
			out, err := executeTemplate(d.Name(), content, data)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			return os.WriteFile(targetPath, out, 0644)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, content, 0644)
	})
}

func applyPathReplacements(rel string, repl map[string]string) string {
	if repl == nil {
		return rel
	}
	out := rel
	for k, v := range repl {
		out = strings.ReplaceAll(out, k, v)
	}
	return out
}

func executeTemplate(name string, content []byte, data map[string]any) ([]byte, error) {
	baseName := name
	if strings.HasSuffix(baseName, ".tmpl") {
		baseName = baseName[:len(baseName)-5]
	}
	tpl, err := template.New(baseName).Parse(string(content))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func RenderDir(srcDir, dstDir string, data map[string]any) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dstDir, rel)
		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".tmpl") {
			out, err := executeTemplate(info.Name(), content, data)
			if err != nil {
				return err
			}
			targetPath = targetPath[:len(targetPath)-5]
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			return os.WriteFile(targetPath, out, 0644)
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, content, 0644)
	})
}
