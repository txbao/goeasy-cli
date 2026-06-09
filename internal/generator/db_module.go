package generator

import (
	"fmt"
	"os"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

// GenerateDBModule 按表生成模块骨架与持久化层（无 CRUD HTTP 覆盖）。
func GenerateDBModule(opts DBOptions) error {
	dsn, driver, prefix, rules, err := readProjectCodegen(opts.ProjectDir, opts.ConfigPath)
	if err != nil {
		return err
	}
	tables, err := resolveDBTables(opts, dsn, prefix, driver)
	if err != nil {
		return err
	}
	for _, physical := range tables {
		module := resolveModuleName(opts, physical, prefix)
		if err := generateDBModuleForTable(opts, dsn, driver, prefix, rules, physical, module); err != nil {
			return err
		}
	}
	return nil
}

func generateDBModuleForTable(opts DBOptions, dsn, driver, prefix string, rules schema.CodegenRules, physical, module string) error {
	tableMeta, err := loadTableMeta(driver, dsn, opts.Schema, physical)
	if err != nil {
		return err
	}
	ct := schema.Classify(tableMeta, module, physical, rules)
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	pascal := utils.ToPascal(module)
	alias := utils.ToIdent(module)
	layoutMeta := moduleMetaFromDB(opts, module)

	fmt.Fprintf(os.Stderr, "info: generate db module for table %s -> module %s\n", physical, module)

	if err := ensureDBXPackage(opts.ModuleOptions); err != nil {
		return err
	}
	if err := ensureGoquDeps(opts.ProjectDir); err != nil {
		return err
	}
	modOpts := ModuleOptions{
		ProjectDir: opts.ProjectDir,
		ModuleName: module,
		Force:      opts.Force,
		Domain:     opts.Domain,
		Group:      opts.Group,
		Resource:   opts.Resource,
		ConfigPath: opts.ConfigPath,
		AppStyle:   opts.AppStyle,
	}
	style, err := resolveAppStyleForModule(modOpts)
	if err != nil {
		return err
	}
	if !moduleExists(opts.ProjectDir, layoutMeta) {
		if err := GenerateModule(modOpts); err != nil {
			return err
		}
	}

	files := map[string]string{
		layoutMeta.domainRel("repository.go"):              genRepositoryIface(layoutMeta.Resource),
		layoutMeta.domainRel("entity.go"):                  genEntity(projectModule, ct, pascal, layoutMeta),
		layoutMeta.domainRel("aggregate.go"):               genAggregate(projectModule, ct, pascal, layoutMeta),
		persistenceRepoRel(layoutMeta, "repository_pg.go"): genRepositoryPG(projectModule, currentGoEasyModule(), ct, pascal, alias, layoutMeta),
		persistenceRepoRel(layoutMeta, "repository.go"):    genMemoryRepository(projectModule, ct, pascal, layoutMeta),
	}
	for rel, content := range buildDBAppLayerFiles(style, projectModule, ct, pascal, alias, layoutMeta) {
		files[rel] = content
	}
	skipped, skippedCritical, _, err := writeDBGeneratedFiles(opts.ProjectDir, files, layoutMeta.ModuleID, opts.Force)
	if err != nil {
		return err
	}
	warnImportCycleSkipped(layoutMeta.ModuleID, skippedCritical)
	if skipped > 0 && !opts.Force {
		fmt.Fprintf(os.Stderr, "warn: %d file(s) skipped (existing). Re-run with --force to overwrite with database schema.\n", skipped)
	}

	if opts.SkipRegister {
		fmt.Fprintf(os.Stderr, "info: skip register (--skip-register)\n")
		return nil
	}
	if err := renderRegisterFile(modOpts, false, opts.Force); err != nil {
		return err
	}
	return ensureModulesRegistry(modOpts)
}
