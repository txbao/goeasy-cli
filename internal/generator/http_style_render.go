package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

// writeHTTPHandlersFromCodegen 名字驱动 module/crud 的 HTTP handler 一律由 codegen 生成（覆盖静态模板）。
func writeHTTPHandlersFromCodegen(opts ModuleOptions, projectModule string, clients []ClientSurface, meta ModuleMeta, style AppStyle, crudOverlay bool) error {
	goeasy := currentGoEasyModule()
	pascal := utils.ToPascal(meta.ModuleID)
	alias := utils.ToIdent(meta.ModuleID)
	ct := schema.ClassifiedTable{ModuleName: meta.ModuleID}

	for _, cl := range clients {
		targets := map[string]string{
			meta.HTTPRel(cl.Name, "handler.go"): genHandler(projectModule, cl.Name, ct, pascal, alias, goeasy, cl.FullCRUD, meta, style),
		}
		if crudOverlay {
			targets[meta.HTTPRel(cl.Name, "handler_crud.go")] = genHandlerCrud(projectModule, cl.Name, ct, pascal, goeasy, cl.FullCRUD, meta, style)
		}
		for rel, content := range targets {
			path := filepath.Join(opts.ProjectDir, rel)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return err
			}
			fmt.Printf("  created %s\n", filepath.ToSlash(rel))
		}
	}
	return nil
}
