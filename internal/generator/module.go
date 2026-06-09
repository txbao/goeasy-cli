package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/utils"
)

type ModuleOptions struct {
	ProjectDir    string
	ModuleName    string
	Force         bool
	WithMigration bool
	Clients       []string // HTTP 客户端：admin（默认）, h5, app
	PublicClients []string // 不挂鉴权中间件的 client（--public h5）
	Domain        string   // 限界上下文，如 system
	Group         string   // 已废弃，等同 Domain（兼容 --group）
	Resource      string   // URL/包资源名，如 roles
	ConfigPath    string   // 读取 codegen.domains，默认 configs/config.yaml
	AppStyle      string   // CLI --app-style；空则读 config
}

func GenerateModule(opts ModuleOptions) error {
	if opts.ModuleName == "" {
		return fmt.Errorf("module name is required")
	}
	style, err := resolveAppStyleForModule(opts)
	if err != nil {
		return err
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	snake := utils.ToSnake(opts.ModuleName)
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	if style.IsLightCQRS() {
		removeModuleListQuery(opts.ProjectDir, meta)
	}
	clients, err := ResolveHTTPClients(opts.Clients, opts.PublicClients)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	enrichModuleData(data, opts.ModuleName, false, false)
	enrichModuleMetaData(data, meta)
	enrichHTTPClientsDataWithMeta(data, clients, meta)
	if err := ensureDBXPackage(opts); err != nil {
		return err
	}
	repl := moduleTemplateRepl(meta)
	excludes := []string{"interface/http/CLIENT", "bootstrap/snippets"}
	skip := []string{}
	if style.IsService() {
		excludes = append(excludes, "/command/", "/query/")
		skip = append(skip, filepath.ToSlash(meta.appRel("application.go")))
	}
	if err := renderScopedFiltered("module", opts.ProjectDir, repl, data, opts.Force, "", excludes, skip...); err != nil {
		return err
	}
	if style.IsService() {
		pascal := utils.ToPascal(opts.ModuleName)
		appPath := filepath.Join(opts.ProjectDir, meta.appRel("application.go"))
		appSrc := genModuleSkeletonServiceApplication(projectModule, meta, pascal)
		if opts.Force {
			if err := os.WriteFile(appPath, []byte(appSrc), 0644); err != nil {
				return err
			}
			fmt.Printf("  created %s\n", meta.appRel("application.go"))
		} else if _, err := os.Stat(appPath); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(appPath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(appPath, []byte(appSrc), 0644); err != nil {
				return err
			}
			fmt.Printf("  created %s\n", meta.appRel("application.go"))
		}
	}
	if err := renderHTTPClientLayer("module", opts.ProjectDir, snake, projectModule, clients, data, opts.Force); err != nil {
		return err
	}
	if err := writeHTTPHandlersFromCodegen(opts, projectModule, clients, meta, style, false); err != nil {
		return err
	}
	if err := ensureGoquDeps(opts.ProjectDir); err != nil {
		return err
	}
	opts.Clients = clientNames(clients)
	if err := renderRegisterFile(opts, false, opts.Force); err != nil {
		return err
	}
	if err := ensureRepositoryPG(opts); err != nil {
		return err
	}
	return ensureModulesRegistry(opts)
}

func GenerateCRUD(opts ModuleOptions) error {
	if opts.ModuleName == "" {
		return fmt.Errorf("module name is required")
	}
	snake := utils.ToSnake(opts.ModuleName)
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	if !moduleExists(opts.ProjectDir, meta) {
		if err := GenerateModule(opts); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(os.Stderr, "info: module %q skeleton exists, skipping module template\n", snake)
	}
	if err := ensureDBXPackage(opts); err != nil {
		return err
	}
	if err := ensureGoquDeps(opts.ProjectDir); err != nil {
		return err
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	clients, err := ResolveHTTPClients(opts.Clients, opts.PublicClients)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	enrichModuleMetaData(data, meta)
	enrichHTTPClientsDataWithMeta(data, clients, meta)
	style, err := resolveAppStyleForModule(opts)
	if err != nil {
		return err
	}
	if err := renderHTTPCrudOverlay(opts.ProjectDir, snake, projectModule, clients, data, opts.Force); err != nil {
		return err
	}
	if err := writeHTTPHandlersFromCodegen(opts, projectModule, clients, meta, style, true); err != nil {
		return err
	}
	if err := ensureModuleCRUDLayer(opts); err != nil {
		return err
	}
	if err := ensureRepositoryPG(opts); err != nil {
		return err
	}
	opts.Clients = clientNames(clients)
	if err := renderRegisterFile(opts, true, opts.Force); err != nil {
		return err
	}
	if err := ensureModulesRegistry(opts); err != nil {
		return err
	}
	if opts.WithMigration {
		return writeModuleTableMigration(opts, snake)
	}
	return nil
}

func ensureRepositoryPG(opts ModuleOptions) error {
	snake := utils.ToSnake(opts.ModuleName)
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	pgPath := filepath.Join(opts.ProjectDir, persistenceRepoRel(meta, "repository_pg.go"))
	legacyPG := filepath.Join(opts.ProjectDir, "internal", "infrastructure", "persistence", "repository", snake, "repository_pg.go")
	if _, err := os.Stat(pgPath); err != nil {
		if _, err2 := os.Stat(legacyPG); err2 == nil {
			pgPath = legacyPG
		}
	}
	if _, err := os.Stat(pgPath); err == nil && !opts.Force {
		fmt.Fprintf(os.Stderr, "info: %s exists, skipping repository_pg\n", filepath.ToSlash(pgPath))
		return nil
	}
	return GenerateRepository(opts)
}

func GenerateRepository(opts ModuleOptions) error {
	if opts.ModuleName == "" {
		return fmt.Errorf("module name is required")
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	enrichModuleMetaData(data, meta)
	repl := moduleTemplateRepl(meta)

	var skip []string
	if moduleExists(opts.ProjectDir, meta) {
		fmt.Fprintf(os.Stderr, "info: module %q exists, skipping domain/repository.go (only adding PG stub)\n", meta.ModuleID)
		skip = append(skip, filepath.ToSlash(meta.domainRel("repository.go")))
	}
	return renderScoped("repository", opts.ProjectDir, repl, data, opts.Force, skip...)
}

func GenerateProto(opts ModuleOptions) error {
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	return renderScoped("proto", opts.ProjectDir, moduleTemplateRepl(meta), data, opts.Force)
}

func GenerateEvent(opts ModuleOptions) error {
	if opts.ModuleName == "" {
		return fmt.Errorf("event name is required")
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	domain := strings.TrimSpace(opts.Domain)
	if domain == "" {
		domain = strings.TrimSpace(opts.Group)
	}
	if domain == "" {
		domain = "integration"
	}
	data := toMap(BuildEventData(opts.ModuleName, projectModule))
	data["Domain"] = domain
	repl := map[string]string{
		"EVENT":  data["EventSnake"].(string),
		"DOMAIN": domain,
	}
	return renderScoped("event", opts.ProjectDir, repl, data, opts.Force)
}

func GenerateAggregate(opts ModuleOptions) error {
	fmt.Fprintf(os.Stderr, "warn: add aggregate is deprecated; use add module or add crud instead\n")
	if opts.ModuleName == "" {
		return fmt.Errorf("aggregate name is required")
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	snake := utils.ToSnake(opts.ModuleName)
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	if moduleExists(opts.ProjectDir, meta) {
		return fmt.Errorf("module %q already exists; use add module or add crud instead of add aggregate", snake)
	}
	data := toMap(BuildModuleData(opts.ModuleName, projectModule))
	return renderScoped("aggregate", opts.ProjectDir, moduleTemplateRepl(meta), data, opts.Force)
}

func moduleExists(projectDir string, meta ModuleMeta) bool {
	entity := filepath.Join(projectDir, meta.domainRel("entity.go"))
	_, err := os.Stat(entity)
	return err == nil
}

// renderScopedOrSkip 渲染模板；目标文件已存在且未 --force 时跳过（不报错）。
func renderScopedOrSkip(templateRoot, projectDir string, repl map[string]string, data map[string]any, force bool, skipTargets ...string) error {
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
		if path == "." {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.Contains(path, "register_MODULE") || strings.Contains(path, "register_DOMAIN") {
			return nil
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

func renderScoped(templateRoot, projectDir string, repl map[string]string, data map[string]any, force bool, skipTargets ...string) error {
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
		if path == "." {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.Contains(path, "register_MODULE") || strings.Contains(path, "register_DOMAIN") {
			return nil
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
