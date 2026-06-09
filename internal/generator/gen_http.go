package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/utils"
)

// GenerateHTTPFromOpenAPI 契约驱动：从 OpenAPI 生成 HTTP 接口层与应用层桩。
func GenerateHTTPFromOpenAPI(opts GenHTTPOptions) error {
	files, err := resolveOpenAPIFiles(opts.ProjectDir, opts.OpenAPIFile, opts.OpenAPIDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := generateHTTPFromOpenAPIFile(opts, f); err != nil {
			return err
		}
	}
	return nil
}

func generateHTTPFromOpenAPIFile(opts GenHTTPOptions, openapiPath string) error {
	contract, err := parseOpenAPIContract(openapiPath)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "info: gen http from %s -> module %s\n", filepath.ToSlash(openapiPath), contract.ModuleSnake)

	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}

	cfg := readCodegenDomains(opts.ProjectDir, opts.ConfigPath)
	layout := contract.Layout
	if opts.Domain != "" || opts.Group != "" || opts.Resource != "" {
		domain := opts.Domain
		if domain == "" {
			domain = opts.Group
		}
		layout = ResolveModuleMeta(contract.ModuleSnake, domain, opts.Resource, cfg.Domains)
	} else if layout.Domain == "" {
		layout = ResolveModuleMeta(contract.ModuleSnake, "", "", cfg.Domains)
	}
	layout.ModuleID = contract.ModuleSnake
	layout.ModuleSnake = contract.ModuleSnake
	normalizeMeta(&layout)

	if opts.Force && repositoryPGExists(opts.ProjectDir, layout) && !opts.AllowOverwrite {
		return fmt.Errorf(
			"module %s has add db crud artifacts (repository_pg); --force would overwrite them. "+
				"Use add db openapi to sync contract, gen http --merge-http for incremental HTTP, or add --allow-overwrite",
			contract.ModuleSnake,
		)
	}

	clients, err := clientsFromContract(contract, opts.Clients)
	if err != nil {
		return err
	}
	clients, err = applyPublicClients(clients, opts.PublicClients)
	if err != nil {
		return err
	}

	modOpts := ModuleOptions{
		ProjectDir:    opts.ProjectDir,
		ModuleName:    contract.ModuleSnake,
		Force:         opts.Force,
		Clients:       clientNames(clients),
		PublicClients: opts.PublicClients,
		Domain:        opts.Domain,
		Group:         opts.Group,
		Resource:      opts.Resource,
		ConfigPath:    defaultConfigPath(opts.ConfigPath),
		AppStyle:      opts.AppStyle,
	}
	style, err := resolveAppStyleForModule(modOpts)
	if err != nil {
		return err
	}

	withApp := opts.WithApp && !opts.MergeHTTP
	if !moduleExists(opts.ProjectDir, layout) {
		if err := GenerateModule(modOpts); err != nil {
			return err
		}
	}

	if withApp && len(contract.CT.ReadCols) > 0 {
		if err := writeAppLayerFromContract(opts, projectModule, contract, layout, style); err != nil {
			return err
		}
	}

	if err := writeHTTPLayerFromContract(opts, projectModule, contract, layout, clients, style); err != nil {
		return err
	}
	if opts.MergeHTTP {
		if err := writeHTTPLayerExtras(opts, projectModule, contract, layout, clients); err != nil {
			return err
		}
	}

	modOpts.Clients = clientNames(clients)
	registerForce := opts.Force && !opts.MergeHTTP
	return renderRegisterFile(modOpts, true, registerForce)
}

func defaultConfigPath(p string) string {
	if strings.TrimSpace(p) == "" {
		return "configs/config.yaml"
	}
	return p
}

func writeAppLayerFromContract(opts GenHTTPOptions, projectModule string, contract OpenAPIContract, meta ModuleMeta, style AppStyle) error {
	pascal := contract.ModulePascal
	alias := utils.ToIdent(contract.ModuleSnake)
	ct := contract.CT

	files := map[string]string{
		meta.domainRel("entity.go"):    genEntity(projectModule, ct, pascal, meta),
		meta.domainRel("aggregate.go"): genAggregate(projectModule, ct, pascal, meta),
	}
	for rel, content := range buildDBAppLayerFiles(style, projectModule, ct, pascal, alias, meta) {
		files[rel] = content
	}
	skipped, _, _, err := writeDBGeneratedFiles(opts.ProjectDir, files, meta.ModuleID, opts.Force)
	if err != nil {
		return err
	}
	if skipped > 0 && !opts.Force {
		fmt.Fprintf(os.Stderr, "info: %d app file(s) skipped (use --force)\n", skipped)
	}
	return ensureModulesRegistry(ModuleOptions{ProjectDir: opts.ProjectDir, ModuleName: contract.ModuleSnake, ConfigPath: defaultConfigPath(opts.ConfigPath)})
}

func writeHTTPLayerFromContract(opts GenHTTPOptions, projectModule string, contract OpenAPIContract, meta ModuleMeta, clients []ClientSurface, style AppStyle) error {
	pascal := contract.ModulePascal
	alias := utils.ToIdent(contract.ModuleSnake)
	goeasy := currentGoEasyModule()
	ct := contract.CT

	files := map[string]string{}
	for _, cl := range clients {
		ops := contract.ClientOps[cl.Name]
		fullCRUD := clientFullCRUDFromOps(ops)
		cl.FullCRUD = fullCRUD
		files[meta.HTTPRel(cl.Name, "dto.go")] = genHTTPClientDTO(projectModule, cl.Name, ct, pascal, meta)
		if ops.Get || ops.Create {
			files[meta.HTTPRel(cl.Name, "handler.go")] = genHandler(projectModule, cl.Name, ct, pascal, alias, goeasy, fullCRUD, meta, style)
			files[meta.HTTPRel(cl.Name, "router.go")] = genRouter(meta, fullCRUD)
		}
		if ops.List || ops.Update || ops.Delete {
			files[meta.HTTPRel(cl.Name, "handler_crud.go")] = genHandlerCrud(projectModule, cl.Name, ct, pascal, goeasy, fullCRUD, meta, style)
			files[meta.HTTPRel(cl.Name, "router_crud.go")] = genRouterCrud(meta, fullCRUD)
		}
	}
	skipped, _, _, err := writeDBGeneratedFiles(opts.ProjectDir, files, meta.ModuleID, opts.Force)
	if err != nil {
		return err
	}
	if skipped > 0 && !opts.Force {
		fmt.Fprintf(os.Stderr, "info: %d http file(s) skipped (use --force)\n", skipped)
	}
	return nil
}
