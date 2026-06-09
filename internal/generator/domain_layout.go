package generator

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/txbao/goeasy-cli/internal/utils"
	"gopkg.in/yaml.v3"
)

// ModuleMeta 领域边界布局元数据（layout: domain）。
type ModuleMeta struct {
	ModuleID    string // 表/缓存/proto 逻辑名，如 sys_roles
	ModuleSnake string // 与 ModuleID 同义（兼容旧字段）
	Domain      string // 限界上下文，如 system
	Resource    string // Go 包名与 URL 资源段，如 roles
	Grouped     bool   // 兼容：domain 布局恒为 true
	Group       string // 兼容：等同 Domain
}

func normalizeMeta(m *ModuleMeta) {
	if m.ModuleID == "" && m.ModuleSnake != "" {
		m.ModuleID = m.ModuleSnake
	}
	if m.ModuleSnake == "" {
		m.ModuleSnake = m.ModuleID
	}
	if m.Domain == "" && m.Group != "" {
		m.Domain = m.Group
	}
	if m.Group == "" {
		m.Group = m.Domain
	}
	if m.Domain != "" && m.Resource != "" {
		m.Grouped = true
	}
}

// PackageName 领域/应用/基础设施包名。
func (m ModuleMeta) PackageName() string {
	return m.Resource
}

// RoutePrefix 挂在 /api/v1/{client} 后的路径前缀。
func (m ModuleMeta) RoutePrefix() string {
	return "/" + m.Domain + "/" + m.Resource
}

func (m ModuleMeta) domainRel(file string) string {
	return filepath.Join("internal", "domain", m.Domain, m.Resource, file)
}

func (m ModuleMeta) appRel(file string) string {
	return filepath.Join("internal", "app", m.Domain, m.Resource, file)
}

func (m ModuleMeta) persistenceRel(file string) string {
	return filepath.Join("internal", "infrastructure", m.Domain, "persistence", m.Resource, file)
}

func (m ModuleMeta) grpcRel(file string) string {
	return filepath.Join("internal", "interface", "grpc", m.Domain, m.Resource, file)
}

func (m ModuleMeta) HTTPDir(client string) string {
	normalizeMeta(&m)
	return filepath.Join("internal", "interface", "http", client, m.Domain, m.Resource)
}

func (m ModuleMeta) HTTPRel(client, file string) string {
	return filepath.Join(m.HTTPDir(client), file)
}

func (m ModuleMeta) HTTPImportSuffix(client string) string {
	return filepath.ToSlash(m.HTTPDir(client))
}

func (m ModuleMeta) GRPCImportSuffix() string {
	return filepath.ToSlash(filepath.Join("internal", "interface", "grpc", m.Domain, m.Resource))
}

func (m ModuleMeta) DomainImportPath(projectModule string) string {
	return projectModule + "/" + filepath.ToSlash(filepath.Join("internal", "domain", m.Domain, m.Resource))
}

func (m ModuleMeta) AppImportPath(projectModule string) string {
	return projectModule + "/" + filepath.ToSlash(filepath.Join("internal", "app", m.Domain, m.Resource))
}

func (m ModuleMeta) PersistenceImportPath(projectModule string) string {
	return projectModule + "/" + filepath.ToSlash(filepath.Join("internal", "infrastructure", m.Domain, "persistence", m.Resource))
}

type domainModuleEntry struct {
	Resource string `yaml:"resource"`
}

type domainConfig struct {
	TablePrefix string                       `yaml:"table_prefix"`
	Modules     map[string]domainModuleEntry `yaml:"modules"`
}

type codegenDomainsConfig struct {
	Layout  string                  `yaml:"layout"`
	Domains map[string]domainConfig `yaml:"domains"`
}

func readCodegenDomains(projectDir, configPath string) codegenDomainsConfig {
	cfg := codegenDomainsConfig{Layout: "domain"}
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(projectDir, configPath)
	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}
	var raw struct {
		Codegen codegenDomainsConfig `yaml:"codegen"`
	}
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return cfg
	}
	if raw.Codegen.Layout != "" {
		cfg.Layout = raw.Codegen.Layout
	}
	if len(raw.Codegen.Domains) > 0 {
		cfg.Domains = raw.Codegen.Domains
	}
	return cfg
}

// ResolveModuleMeta 解析模块领域布局；explicitDomain/resource 优先，其次 config domains。
func ResolveModuleMeta(moduleID, explicitDomain, explicitResource string, domains map[string]domainConfig) ModuleMeta {
	moduleID = strings.TrimSpace(moduleID)
	meta := ModuleMeta{ModuleID: moduleID}

	domain := strings.TrimSpace(explicitDomain)
	resource := strings.TrimSpace(explicitResource)

	if domain == "" && len(domains) > 0 {
		for d, dc := range domains {
			if dc.Modules != nil {
				if ent, ok := dc.Modules[moduleID]; ok {
					domain = strings.TrimSpace(d)
					if ent.Resource != "" {
						resource = strings.TrimSpace(ent.Resource)
					}
					break
				}
			}
		}
	}

	if domain == "" && len(domains) > 0 {
		type prefixEntry struct {
			domain string
			prefix string
		}
		var entries []prefixEntry
		for d, dc := range domains {
			p := strings.TrimSpace(dc.TablePrefix)
			if p != "" {
				entries = append(entries, prefixEntry{domain: strings.TrimSpace(d), prefix: p})
			}
		}
		sort.Slice(entries, func(i, j int) bool { return len(entries[i].prefix) > len(entries[j].prefix) })
		for _, e := range entries {
			if strings.HasPrefix(moduleID, e.prefix) {
				domain = e.domain
				if resource == "" {
					resource = strings.Trim(strings.TrimPrefix(moduleID, e.prefix), "_")
				}
				break
			}
		}
	}

	if domain == "" {
		domain = moduleID
	}
	if resource == "" {
		resource = moduleID
	}

	meta.Domain = domain
	meta.Resource = resource
	meta.ModuleSnake = meta.ModuleID
	meta.Group = meta.Domain
	meta.Grouped = true
	return meta
}

func resolveModuleMetaForModule(opts ModuleOptions, configPath string) ModuleMeta {
	snake := strings.TrimSpace(opts.ModuleName)
	if snake == "" {
		return ModuleMeta{}
	}
	cfg := readCodegenDomains(opts.ProjectDir, configPath)
	domain := opts.Domain
	if domain == "" {
		domain = opts.Group // 兼容旧 --group
	}
	return ResolveModuleMeta(snake, domain, opts.Resource, cfg.Domains)
}

func enrichModuleMetaData(data map[string]any, meta ModuleMeta) {
	data["ModuleID"] = meta.ModuleID
	data["Domain"] = meta.Domain
	data["Resource"] = meta.Resource
	data["DomainPascal"] = utils.ToPascal(meta.Domain)
	if v, ok := data["ModulePascal"]; !ok || v == "" {
		data["ModulePascal"] = utils.ToPascal(meta.Resource)
	}
	data["Package"] = meta.PackageName()
	data["RoutePrefix"] = meta.RoutePrefix()
	data["HTTPPackage"] = meta.PackageName()
	data["HTTPRoutePrefix"] = meta.RoutePrefix()
	data["GRPCImportPath"] = meta.GRPCImportSuffix()
	// 兼容旧模板键
	data["ModuleSnake"] = meta.ModuleID
	data["HTTPGrouped"] = true
	data["HTTPGroup"] = meta.Domain
	data["HTTPResource"] = meta.Resource
}

func enrichHTTPClientsDataWithMeta(data map[string]any, clients []ClientSurface, meta ModuleMeta) {
	surfaces := make([]map[string]any, 0, len(clients))
	for _, c := range clients {
		s := c
		s.ImportAlias = clientImportAlias(s.Name, meta.ModuleID)
		surfaces = append(surfaces, map[string]any{
			"Name":           s.Name,
			"Pascal":         s.Pascal,
			"Middleware":     s.Middleware,
			"FullCRUD":       s.FullCRUD,
			"UseAuth":        s.UseAuth,
			"ImportAlias":    s.ImportAlias,
			"HttpImportPath": meta.HTTPImportSuffix(s.Name),
		})
	}
	data["Clients"] = surfaces
	data["NeedHTTPMiddleware"] = clientsNeedHTTPMiddleware(clients)
	if len(surfaces) > 0 {
		data["HTTPClient"] = surfaces[0]["Name"]
	}
}

func metaFromData(data map[string]any) ModuleMeta {
	meta := ModuleMeta{}
	if v, ok := data["ModuleID"].(string); ok {
		meta.ModuleID = v
	} else if v, ok := data["ModuleSnake"].(string); ok {
		meta.ModuleID = v
	}
	if v, ok := data["Domain"].(string); ok {
		meta.Domain = v
	} else if v, ok := data["HTTPGroup"].(string); ok {
		meta.Domain = v
	}
	if v, ok := data["Resource"].(string); ok {
		meta.Resource = v
	} else if v, ok := data["HTTPResource"].(string); ok {
		meta.Resource = v
	}
	if meta.Domain == "" && meta.ModuleID != "" {
		cfg := codegenDomainsConfig{}
		meta = ResolveModuleMeta(meta.ModuleID, "", "", cfg.Domains)
	}
	return meta
}

func moduleMetaByID(moduleID string) ModuleMeta {
	return ModuleMeta{ModuleID: moduleID, Domain: moduleID, Resource: moduleID}
}

func moduleMetaFromDB(opts DBOptions, module string) ModuleMeta {
	return resolveModuleMetaForModule(ModuleOptions{
		ProjectDir: opts.ProjectDir,
		ModuleName: module,
		Domain:     opts.Domain,
		Group:      opts.Group,
		Resource:   opts.Resource,
		ConfigPath: opts.ConfigPath,
	}, opts.ConfigPath)
}

func moduleTemplateRepl(meta ModuleMeta) map[string]string {
	return map[string]string{
		"DOMAIN":    meta.Domain,
		"RESOURCE":  meta.Resource,
		"MODULE":    meta.ModuleID,
		"MODULE_ID": meta.ModuleID,
	}
}

func httpTemplateRepl(client string, meta ModuleMeta) (repl map[string]string, includeSubstr string) {
	repl = moduleTemplateRepl(meta)
	repl["CLIENT"] = client
	return repl, "interface/http/CLIENT/DOMAIN/RESOURCE"
}

// API 契约目录约定。
const (
	APIContractsOpenAPI = "api/contracts/openapi"
	APIGeneratedOpenAPI = "api/generated/openapi"
	APIProtoDir         = "api/proto"
	APIProtoGenDir      = "api/proto/gen"
	APIExamplesDir      = "api/examples"
)
