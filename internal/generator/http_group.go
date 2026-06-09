package generator

import "strings"

// HTTPRouteLayout 已合并为 ModuleMeta（layout: domain）。
type HTTPRouteLayout = ModuleMeta

func enrichHTTPRouteData(data map[string]any, meta ModuleMeta) {
	enrichModuleMetaData(data, meta)
}

func enrichHTTPClientsDataWithLayout(data map[string]any, clients []ClientSurface, meta ModuleMeta) {
	enrichHTTPClientsDataWithMeta(data, clients, meta)
}

func resolveHTTPRouteForModule(opts ModuleOptions, configPath string) ModuleMeta {
	return resolveModuleMetaForModule(opts, configPath)
}

// ResolveHTTPRoute 兼容旧 group_prefixes 测试；新配置请用 ResolveModuleMeta + domains。
func ResolveHTTPRoute(moduleID, explicitDomain, explicitResource string, prefixToGroup map[string]string) ModuleMeta {
	domains := make(map[string]domainConfig, len(prefixToGroup))
	for prefix, domain := range prefixToGroup {
		domain = strings.TrimSpace(domain)
		prefix = strings.TrimSpace(prefix)
		if domain == "" || prefix == "" {
			continue
		}
		dc := domains[domain]
		dc.TablePrefix = prefix
		domains[domain] = dc
	}
	return ResolveModuleMeta(moduleID, explicitDomain, explicitResource, domains)
}

func readCodegenGroupPrefixes(projectDir, configPath string) map[string]string {
	cfg := readCodegenDomains(projectDir, configPath)
	out := make(map[string]string)
	for domain, dc := range cfg.Domains {
		if p := strings.TrimSpace(dc.TablePrefix); p != "" {
			out[p] = domain
		}
	}
	return out
}
