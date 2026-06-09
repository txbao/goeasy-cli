package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/txbao/goeasy-cli/internal/utils"
)

func parseDomainModuleIDs(content string) []string {
	var ids []string
	seen := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "goeasy-module:") {
			continue
		}
		i := strings.Index(line, "goeasy-module:")
		id := strings.TrimSpace(line[i+len("goeasy-module:"):])
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

func collectDomainModuleIDs(projectDir, registerPath, currentModuleID string) []string {
	seen := map[string]bool{}
	var ids []string
	if b, err := os.ReadFile(registerPath); err == nil {
		for _, id := range parseDomainModuleIDs(string(b)) {
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	if !seen[currentModuleID] {
		ids = append(ids, currentModuleID)
	}
	sort.Strings(ids)
	return ids
}

func discoverModuleHTTPClients(projectDir string, meta ModuleMeta) []string {
	var clients []string
	for _, name := range []string{"admin", "h5", "app"} {
		dir := filepath.Join(projectDir, meta.HTTPDir(name))
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			clients = append(clients, name)
		}
	}
	if len(clients) == 0 {
		return []string{"admin"}
	}
	return clients
}

func moduleWantsCRUD(projectDir string, meta ModuleMeta) bool {
	rel := meta.HTTPRel("admin", "router_crud.go")
	if _, err := os.Stat(filepath.Join(projectDir, rel)); err == nil {
		return true
	}
	for _, c := range []string{"h5", "app"} {
		rel = meta.HTTPRel(c, "router_crud.go")
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err == nil {
			return true
		}
	}
	return false
}

func buildDomainRegisterData(projectDir, domain string, moduleIDs []string, current ModuleOptions, withCRUD bool) (map[string]any, error) {
	projectModule, err := readModulePath(projectDir)
	if err != nil {
		return nil, err
	}
	cfg := readCodegenDomains(projectDir, current.ConfigPath)

	type modEntry struct {
		data           map[string]any
		clientSurfaces []ClientSurface
	}

	entries := make([]modEntry, 0, len(moduleIDs))
	clientUnion := map[string]ClientSurface{}
	anyPG := false
	needMiddleware := false

	for _, moduleID := range moduleIDs {
		modOpts := current
		modOpts.ModuleName = moduleID
		if moduleID != current.ModuleName {
			modOpts.Clients = discoverModuleHTTPClients(projectDir, resolveModuleMetaForModule(modOpts, current.ConfigPath))
			modOpts.PublicClients = nil
		}
		meta := resolveModuleMetaForModule(modOpts, current.ConfigPath)
		clients, err := ResolveHTTPClients(modOpts.Clients, modOpts.PublicClients)
		if err != nil {
			return nil, err
		}
		modCRUD := withCRUD && moduleID == current.ModuleName
		if moduleID != current.ModuleName {
			modCRUD = moduleWantsCRUD(projectDir, meta)
		}
		withPG := repositoryPGExists(projectDir, meta)
		if withPG {
			anyPG = true
		}
		if clientsNeedHTTPMiddleware(clients) {
			needMiddleware = true
		}
		for i := range clients {
			clients[i].ImportAlias = clientImportAlias(clients[i].Name, meta.ModuleID)
			clientUnion[clients[i].Name] = clients[i]
		}

		data := toMap(BuildModuleData(moduleID, projectModule))
		enrichModuleData(data, moduleID, modCRUD, withPG)
		enrichModuleMetaData(data, meta)
		data["AppImport"] = meta.AppImportPath(projectModule)
		data["DomainImport"] = meta.DomainImportPath(projectModule)
		data["PersistenceImport"] = meta.PersistenceImportPath(projectModule)
		if v, ok := data["ModulePascal"].(string); !ok || v == "" {
			data["ModulePascal"] = utils.ToPascal(moduleID)
		}

		surfaces := make([]map[string]any, 0, len(clients))
		for _, c := range clients {
			surfaces = append(surfaces, map[string]any{
				"Name":           c.Name,
				"Pascal":         c.Pascal,
				"Middleware":     c.Middleware,
				"FullCRUD":       c.FullCRUD,
				"UseAuth":        c.UseAuth,
				"ImportAlias":    c.ImportAlias,
				"HttpImportPath": meta.HTTPImportSuffix(c.Name),
			})
		}
		data["ClientSurfaces"] = surfaces
		data["WithCRUD"] = modCRUD

		entries = append(entries, modEntry{data: data, clientSurfaces: clients})
	}

	shared := make([]ClientSurface, 0, len(clientUnion))
	for _, name := range []string{"admin", "h5", "app"} {
		if c, ok := clientUnion[name]; ok {
			shared = append(shared, c)
		}
	}
	sharedMaps := make([]map[string]any, 0, len(shared))
	for _, c := range shared {
		sharedMaps = append(sharedMaps, map[string]any{
			"Name":       c.Name,
			"Pascal":     c.Pascal,
			"Middleware": c.Middleware,
			"UseAuth":    c.UseAuth,
		})
	}

	modules := make([]map[string]any, len(entries))
	for i, e := range entries {
		modules[i] = e.data
	}

	_ = cfg
	return map[string]any{
		"Domain":             domain,
		"DomainPascal":       utils.ToPascal(domain),
		"ProjectModule":      projectModule,
		"GoEasyModule":       currentGoEasyModule(),
		"Modules":            modules,
		"SharedClients":      sharedMaps,
		"AnyWithPG":          anyPG,
		"NeedHTTPMiddleware": needMiddleware,
	}, nil
}

func renderDomainRegisterFile(projectDir, domain string, data map[string]any) error {
	sub, err := fsSub("module")
	if err != nil {
		return err
	}
	tplPath := "internal/bootstrap/register_DOMAIN.go.tmpl"
	content, err := fs.ReadFile(sub, tplPath)
	if err != nil {
		return fmt.Errorf("read register domain template: %w", err)
	}
	out, err := executeTemplate(filepath.Base(tplPath), content, data)
	if err != nil {
		return err
	}
	targetRel := filepath.ToSlash(filepath.Join("internal", "bootstrap", "register_"+domain+".go"))
	targetPath := filepath.Join(projectDir, targetRel)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}
	verb := "created"
	if _, err := os.Stat(targetPath); err == nil {
		verb = "updated"
	}
	if err := os.WriteFile(targetPath, out, 0644); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", verb, targetRel)
	return nil
}
