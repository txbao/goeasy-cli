package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/utils"
)

const modulesRegistryMarker = "// business modules (goeasy add module / add crud appends below)"

func domainRegisterFuncName(domain string) string {
	return "Register" + utils.ToPascal(domain)
}

func moduleRegisterCallLine(funcName string) string {
	return fmt.Sprintf("\tif err := %s(engine, infra); err != nil {\n\t\treturn err\n\t}", funcName)
}

func enrichModuleData(data map[string]any, moduleName string, withCRUD, withPG bool) {
	data["ModuleAlias"] = utils.ToIdent(moduleName)
	data["WithCRUD"] = withCRUD
	data["WithPG"] = withPG
}

func repositoryPGExists(projectDir string, meta ModuleMeta) bool {
	paths := []string{
		persistenceRepoRel(meta, "repository_pg.go"),
		filepath.Join("internal", "infrastructure", "persistence", "repository", meta.ModuleID, "repository_pg.go"),
		filepath.Join("internal", "infrastructure", "persistence", meta.ModuleID, "repository_pg.go"),
	}
	for _, rel := range paths {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err == nil {
			return true
		}
	}
	return false
}

func renderRegisterFile(opts ModuleOptions, withCRUD, force bool) error {
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	targetRel := filepath.ToSlash(filepath.Join("internal", "bootstrap", "register_"+meta.Domain+".go"))
	targetPath := filepath.Join(opts.ProjectDir, targetRel)

	moduleMarker := "// goeasy-module: " + meta.ModuleID
	if b, err := os.ReadFile(targetPath); err == nil {
		content := string(b)
		if strings.Contains(content, moduleMarker) {
			if withCRUD && !strings.Contains(content, "RegisterCRUDRoutes") {
				force = true
			}
			if !force {
				fmt.Fprintf(os.Stderr, "info: %s already registers %s, skipping\n", targetRel, meta.ModuleID)
				return nil
			}
		}
	}

	moduleIDs := collectDomainModuleIDs(opts.ProjectDir, targetPath, meta.ModuleID)
	data, err := buildDomainRegisterData(opts.ProjectDir, meta.Domain, moduleIDs, opts, withCRUD)
	if err != nil {
		return err
	}
	return renderDomainRegisterFile(opts.ProjectDir, meta.Domain, data)
}

func ensureModulesRegistry(opts ModuleOptions) error {
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	funcName := domainRegisterFuncName(meta.Domain)
	callLine := moduleRegisterCallLine(funcName)

	modulesPath := filepath.Join(opts.ProjectDir, "internal", "bootstrap", "modules.go")
	if _, err := os.Stat(modulesPath); os.IsNotExist(err) {
		if err := renderModulesFile(opts); err != nil {
			return err
		}
	}

	b, err := os.ReadFile(modulesPath)
	if err != nil {
		return err
	}
	content := string(b)
	if strings.Contains(content, funcName+"(") {
		return nil
	}
	if !strings.Contains(content, modulesRegistryMarker) {
		repaired, ok := insertRegistryMarker(content, modulesRegistryMarker)
		if !ok {
			fmt.Fprintf(os.Stderr, "warn: modules.go missing registry marker; skip %s (%s)\n", funcName, callLine)
			return nil
		}
		content = repaired
	}
	updated := strings.Replace(content, modulesRegistryMarker+"\n", modulesRegistryMarker+"\n"+callLine+"\n", 1)
	if updated == content {
		updated = strings.Replace(content, modulesRegistryMarker, modulesRegistryMarker+"\n"+callLine, 1)
	}
	if updated == content {
		return nil
	}
	if err := os.WriteFile(modulesPath, []byte(updated), 0644); err != nil {
		return err
	}
	fmt.Printf("  updated internal/bootstrap/modules.go (+%s)\n", funcName)
	return nil
}

func insertRegistryMarker(content, marker string) (string, bool) {
	if strings.Contains(content, marker) {
		return content, true
	}
	const fn = "func registerAllModules"
	idx := strings.Index(content, fn)
	if idx < 0 {
		return content, false
	}
	ret := strings.Index(content[idx:], "return nil")
	if ret < 0 {
		return content, false
	}
	pos := idx + ret
	return content[:pos] + "\t" + marker + "\n\t" + content[pos:], true
}

func renderModulesFile(opts ModuleOptions) error {
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
	tplPath := "internal/bootstrap/modules.go.tmpl"
	content, err := fs.ReadFile(sub, tplPath)
	if err != nil {
		return err
	}
	out, err := executeTemplate(filepath.Base(tplPath), content, data)
	if err != nil {
		return err
	}
	target := filepath.Join(opts.ProjectDir, "internal", "bootstrap", "modules.go")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(target, out, 0644); err != nil {
		return err
	}
	fmt.Printf("  created internal/bootstrap/modules.go\n")
	return nil
}
