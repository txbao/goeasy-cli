package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

func GenerateDBCRUD(opts DBOptions) error {
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
		if err := generateDBCRUDForTable(opts, dsn, driver, prefix, rules, physical, module); err != nil {
			return err
		}
	}
	if err := maybeGenerateDBOpenAPI(opts); err != nil {
		return err
	}
	if opts.WithProto && !opts.SkipProto {
		return GenerateDBProto(opts)
	}
	return nil
}

func resolveDBTables(opts DBOptions, dsn, prefix, driver string) ([]string, error) {
	if opts.Table != "" {
		physical := tableName(prefix, opts.Table)
		if err := assertTableAllowed(physical, opts.Exclude); err != nil {
			return nil, err
		}
		return []string{physical}, nil
	}
	if len(opts.Tables) > 0 {
		var out []string
		for _, t := range opts.Tables {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			physical := tableName(prefix, t)
			if schema.ShouldExcludeTable(physical, opts.Exclude) {
				fmt.Fprintf(os.Stderr, "info: skip excluded table %s\n", physical)
				continue
			}
			out = append(out, physical)
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("no tables matched after exclude")
		}
		return out, nil
	}
	if opts.All {
		schemaName := opts.Schema
		if schemaName == "" {
			schemaName = "public"
		}
		names, err := listDatabaseTables(driver, dsn, schemaName)
		if err != nil {
			return nil, err
		}
		var out []string
		for _, n := range names {
			if schema.ShouldExcludeTable(n, opts.Exclude) {
				continue
			}
			if opts.IncludePrefix != "" && !strings.HasPrefix(n, opts.IncludePrefix) {
				continue
			}
			out = append(out, n)
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("no tables matched after exclude/filter")
		}
		return out, nil
	}
	return nil, fmt.Errorf("specify --table, --tables, or --all")
}

func generateDBCRUDForTable(opts DBOptions, dsn, driver, prefix string, rules schema.CodegenRules, physical, module string) error {
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
	goeasy := currentGoEasyModule()

	fmt.Fprintf(os.Stderr, "info: generate db crud for table %s -> module %s\n", physical, module)

	if err := ensureDBXPackage(opts.ModuleOptions); err != nil {
		return err
	}
	if err := ensureGoquDeps(opts.ProjectDir); err != nil {
		return err
	}
	snake := utils.ToSnake(module)
	layoutMeta := moduleMetaFromDB(opts, module)
	modOpts := ModuleOptions{
		ProjectDir:    opts.ProjectDir,
		ModuleName:    module,
		Force:         opts.Force,
		Clients:       opts.Clients,
		PublicClients: opts.PublicClients,
		Domain:        opts.Domain,
		Group:         opts.Group,
		Resource:      opts.Resource,
		ConfigPath:    opts.ConfigPath,
		AppStyle:      opts.AppStyle,
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
	clients, err := ResolveHTTPClients(opts.Clients, opts.PublicClients)
	if err != nil {
		return err
	}
	routeData := toMap(BuildModuleData(module, projectModule))
	enrichModuleMetaData(routeData, layoutMeta)
	if err := renderHTTPCrudOverlay(opts.ProjectDir, snake, projectModule, clients, routeData, opts.Force); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			return err
		}
		fmt.Fprintf(os.Stderr, "info: crud overlay skipped (files exist)\n")
	}

	files := map[string]string{
		layoutMeta.domainRel("repository.go"):              genRepositoryIface(layoutMeta.Resource),
		layoutMeta.domainRel("entity.go"):                  genEntity(projectModule, ct, pascal, layoutMeta),
		layoutMeta.domainRel("aggregate.go"):               genAggregate(projectModule, ct, pascal, layoutMeta),
		persistenceRepoRel(layoutMeta, "repository_pg.go"): genRepositoryPG(projectModule, goeasy, ct, pascal, alias, layoutMeta),
		persistenceRepoRel(layoutMeta, "repository.go"):    genMemoryRepository(projectModule, ct, pascal, layoutMeta),
	}
	for rel, content := range buildDBAppLayerFiles(style, projectModule, ct, pascal, alias, layoutMeta) {
		files[rel] = content
	}
	appendHTTPClientFiles(files, clients, projectModule, goeasy, ct, pascal, alias, layoutMeta, style)
	skipped, skippedCritical, _, err := writeDBGeneratedFiles(opts.ProjectDir, files, snake, opts.Force)
	if err != nil {
		return err
	}
	if err := writeCRUDCurlFiles(opts.ProjectDir, clients, layoutMeta, ct, opts.Force); err != nil {
		return err
	}
	warnImportCycleSkipped(snake, skippedCritical)
	if skipped > 0 && !opts.Force {
		fmt.Fprintf(os.Stderr, "info: %d file(s) skipped (existing). Use --force to overwrite with database schema.\n", skipped)
	}

	if opts.SkipRegister {
		fmt.Fprintf(os.Stderr, "info: skip register (--skip-register)\n")
		return nil
	}
	if err := renderRegisterFile(modOpts, true, opts.Force); err != nil {
		return err
	}
	return ensureModulesRegistry(modOpts)
}
