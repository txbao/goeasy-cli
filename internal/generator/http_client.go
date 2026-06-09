package generator

import (
	"fmt"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

// ClientSurface HTTP 客户端（路径前缀 + 鉴权 + 生成范围）。
type ClientSurface struct {
	Name        string // admin, h5, app
	Pascal      string
	Middleware  string // AdminAuth, MemberAuth
	FullCRUD    bool
	ImportAlias string
	UseAuth     bool // false 时路由组不挂鉴权中间件（--public）
}

var allowedHTTPClients = map[string]ClientSurface{
	"admin": {Name: "admin", Pascal: "Admin", Middleware: "AdminAuth", FullCRUD: true, UseAuth: true},
	"h5":    {Name: "h5", Pascal: "H5", Middleware: "MemberAuth", FullCRUD: false, UseAuth: true},
	"app":   {Name: "app", Pascal: "App", Middleware: "MemberAuth", FullCRUD: false, UseAuth: true},
}

// NormalizeClients 去重、校验；空则默认 admin。
func NormalizeClients(in []string) ([]ClientSurface, error) {
	if len(in) == 0 {
		in = []string{"admin"}
	}
	seen := make(map[string]bool, len(in))
	out := make([]ClientSurface, 0, len(in))
	for _, raw := range in {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		if seen[name] {
			continue
		}
		base, ok := allowedHTTPClients[name]
		if !ok {
			return nil, fmt.Errorf("unsupported --client %q (allowed: admin, h5, app)", raw)
		}
		seen[name] = true
		out = append(out, base)
	}
	if len(out) == 0 {
		return []ClientSurface{allowedHTTPClients["admin"]}, nil
	}
	return out, nil
}

// ResolveHTTPClients 解析 --client 与 --public（公开端不挂鉴权中间件）。
func ResolveHTTPClients(names, public []string) ([]ClientSurface, error) {
	clients, err := NormalizeClients(names)
	if err != nil {
		return nil, err
	}
	if len(public) == 0 {
		return clients, nil
	}
	return applyPublicClients(clients, public)
}

func applyPublicClients(clients []ClientSurface, public []string) ([]ClientSurface, error) {
	inClients := make(map[string]bool, len(clients))
	for _, c := range clients {
		inClients[c.Name] = true
	}
	pub := make(map[string]bool)
	for _, raw := range public {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		if name == "admin" {
			return nil, fmt.Errorf("--public admin is not allowed: admin routes must use AdminAuth")
		}
		if _, ok := allowedHTTPClients[name]; !ok {
			return nil, fmt.Errorf("unsupported --public %q (allowed: h5, app)", raw)
		}
		if !inClients[name] {
			return nil, fmt.Errorf("--public %q requires the same client in --client", raw)
		}
		pub[name] = true
	}
	out := make([]ClientSurface, len(clients))
	for i, c := range clients {
		out[i] = c
		if pub[c.Name] {
			out[i].UseAuth = false
		}
	}
	return out, nil
}

func clientsNeedHTTPMiddleware(clients []ClientSurface) bool {
	for _, c := range clients {
		if c.UseAuth {
			return true
		}
	}
	return false
}

func clientImportAlias(client, moduleID string) string {
	return utils.ToIdent(client + "_" + moduleID)
}

func httpModuleRel(client string, meta ModuleMeta, file string) string {
	return meta.HTTPRel(client, file)
}

func enrichHTTPClientsData(data map[string]any, clients []ClientSurface, moduleID string) {
	meta := metaFromData(data)
	if meta.ModuleID == "" {
		meta.ModuleID = moduleID
		meta.ModuleSnake = moduleID
	}
	normalizeMeta(&meta)
	enrichHTTPClientsDataWithMeta(data, clients, meta)
}

// readColsForClient admin 用全列；h5/app 省略敏感列。
func readColsForClient(client string, cols []schema.ColumnMeta) []schema.ColumnMeta {
	if client == "admin" {
		return cols
	}
	omit := map[string]bool{
		"deleted_at": true,
		"version":  true,
	}
	out := make([]schema.ColumnMeta, 0, len(cols))
	for _, c := range cols {
		if omit[strings.ToLower(c.Name)] {
			continue
		}
		out = append(out, c)
	}
	if len(out) == 0 {
		return cols
	}
	return out
}
