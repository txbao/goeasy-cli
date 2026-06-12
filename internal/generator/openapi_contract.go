package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
	"gopkg.in/yaml.v3"
)

// OpenAPIExtraEndpoint OpenAPI 中非标准 CRUD 的路径（用于 --merge-http 增量生成）。
type OpenAPIExtraEndpoint struct {
	Client      string
	Method      string
	Path        string
	OperationID string
	Summary     string
}

// OpenAPIContract 从 OpenAPI 3 YAML 解析的模块契约（契约驱动生成入口）。
type OpenAPIContract struct {
	SourceFile     string
	ModuleSnake    string
	ModulePascal   string
	Layout         HTTPRouteLayout
	ClientOps      map[string]clientOperations // client → 可用 HTTP 操作
	ExtraEndpoints []OpenAPIExtraEndpoint
	CT             schema.ClassifiedTable
}

type clientOperations struct {
	List   bool
	Get    bool
	Create bool
	Update bool
	Delete bool
}

// GenHTTPOptions 从 OpenAPI 生成 HTTP 接口层与应用层桩。
type GenHTTPOptions struct {
	ModuleOptions
	OpenAPIFile      string
	OpenAPIDir       string
	WithApp          bool // 生成/更新 app+domain 桩（默认 true）
	AllowOverwrite   bool // 允许 --force 覆盖 add db crud 产物
	MergeHTTP        bool // 仅增量 HTTP：不碰 app/domain，不覆盖已有 HTTP 文件
}

func resolveOpenAPIFiles(projectDir, single, dir string) ([]string, error) {
	if single != "" {
		if !filepath.IsAbs(single) {
			single = filepath.Join(projectDir, single)
		}
		return []string{single}, nil
	}
	root := dir
	if root == "" {
		root = filepath.Join(projectDir, APIOpenAPIDir)
	} else if !filepath.IsAbs(root) {
		root = filepath.Join(projectDir, root)
	}
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".openapi.yaml") || strings.HasSuffix(strings.ToLower(d.Name()), ".yaml") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no OpenAPI files under %s", root)
	}
	return out, nil
}

func parseOpenAPIContract(path string) (OpenAPIContract, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return OpenAPIContract{}, err
	}
	var doc map[string]any
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return OpenAPIContract{}, fmt.Errorf("parse %s: %w", path, err)
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	base = strings.TrimSuffix(base, ".openapi")

	ct := OpenAPIContract{SourceFile: path, ClientOps: map[string]clientOperations{}}
	moduleSnake := utils.ToSnake(base)

	paths, _ := doc["paths"].(map[string]any)
	for rawPath, item := range paths {
		pathItem, _ := item.(map[string]any)
		client, layout := parseHTTPPath(rawPath)
		if client == "" {
			continue
		}
		if ct.Layout.ModuleSnake == "" {
			ct.Layout = layout
			if layout.ModuleSnake != "" && moduleSnake == "" {
				moduleSnake = layout.ModuleSnake
			}
		}
		ops := ct.ClientOps[client]
		for method, opRaw := range pathItem {
			m := strings.ToLower(method)
			if m == "parameters" {
				continue
			}
			op, _ := opRaw.(map[string]any)
			if tag := firstOpenAPITag(op); tag != "" {
				moduleSnake = tag
			}
			opID, _ := op["operationId"].(string)
			summary, _ := op["summary"].(string)
			if isStandardCRUDOperation(client, m, rawPath, layout) {
				switch {
				case m == "get" && strings.Contains(rawPath, "{id}"):
					ops.Get = true
				case m == "get":
					ops.List = true
				case m == "post":
					ops.Create = true
				case m == "put" || m == "patch":
					ops.Update = true
				case m == "delete":
					ops.Delete = true
				}
			} else if client != "" && m != "parameters" {
				ct.ExtraEndpoints = append(ct.ExtraEndpoints, OpenAPIExtraEndpoint{
					Client:      client,
					Method:      m,
					Path:        rawPath,
					OperationID: opID,
					Summary:     summary,
				})
			}
		}
		ct.ClientOps[client] = ops
	}

	if moduleSnake == "" {
		moduleSnake = utils.ToSnake(base)
	}
	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	pascal := inferPascalFromSchemas(schemas, moduleSnake)
	ct.ModuleSnake = moduleSnake
	ct.ModulePascal = pascal

	readCols := openAPISchemaCols(schemas, pascal+"DTO")
	if len(readCols) == 0 {
		readCols = openAPISchemaCols(schemas, pascal)
	}
	createCols := openAPISchemaCols(schemas, "Create"+pascal+"Request")
	updateCols := openAPISchemaCols(schemas, "Update"+pascal+"Request")
	ct.Layout.ModuleSnake = moduleSnake
	ct.CT = schema.ClassifiedTable{
		ModuleName: moduleSnake,
		ReadCols:   readCols,
		CreateCols: createCols,
		UpdateCols: updateCols,
	}
	return ct, nil
}

func parseHTTPPath(raw string) (client string, layout HTTPRouteLayout) {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, "/{"); idx >= 0 {
		raw = raw[:idx]
	}
	segs := strings.Split(strings.Trim(raw, "/"), "/")
	if len(segs) < 3 || segs[0] != "api" || segs[1] != "v1" {
		return "", HTTPRouteLayout{}
	}
	client = segs[2]
	rest := segs[3:]
	switch len(rest) {
	case 0:
		return client, HTTPRouteLayout{}
	case 1:
		return client, ModuleMeta{ModuleID: rest[0], ModuleSnake: rest[0], Domain: rest[0], Resource: rest[0], Grouped: true, Group: rest[0]}
	default:
		return client, ModuleMeta{
			ModuleID: rest[1], ModuleSnake: rest[1],
			Domain: rest[0], Resource: rest[1], Grouped: true, Group: rest[0],
		}
	}
}

func firstOpenAPITag(op map[string]any) string {
	tags, _ := op["tags"].([]any)
	if len(tags) == 0 {
		return ""
	}
	s, _ := tags[0].(string)
	return strings.TrimSpace(s)
}

func inferPascalFromSchemas(schemas map[string]any, moduleSnake string) string {
	if schemas == nil {
		return utils.ToPascal(moduleSnake)
	}
	for name := range schemas {
		if strings.HasSuffix(name, "DTO") {
			return strings.TrimSuffix(name, "DTO")
		}
	}
	return utils.ToPascal(moduleSnake)
}

func openAPISchemaCols(schemas map[string]any, name string) []schema.ColumnMeta {
	if schemas == nil || name == "" {
		return nil
	}
	raw, ok := schemas[name]
	if !ok {
		return nil
	}
	obj, _ := raw.(map[string]any)
	props, _ := obj["properties"].(map[string]any)
	if len(props) == 0 {
		return nil
	}
	names := make([]string, 0, len(props))
	for k := range props {
		names = append(names, k)
	}
	sortStrings(names)
	out := make([]schema.ColumnMeta, 0, len(names))
	for _, n := range names {
		prop, _ := props[n].(map[string]any)
		out = append(out, columnMetaFromOpenAPIProp(n, prop))
	}
	return out
}

func columnMetaFromOpenAPIProp(name string, prop map[string]any) schema.ColumnMeta {
	typ, _ := prop["type"].(string)
	desc, _ := prop["description"].(string)
	dbType := openAPITypeToDB(typ, prop)
	return schema.ColumnMeta{
		Name:     name,
		DBType:   dbType,
		Nullable: true,
		Comment:  desc,
	}
}

func openAPITypeToDB(typ string, prop map[string]any) string {
	switch typ {
	case "integer":
		if format, _ := prop["format"].(string); format == "int64" {
			return "bigint"
		}
		return "integer"
	case "number":
		return "float8"
	case "boolean":
		return "bool"
	default:
		return "varchar"
	}
}

func sortStrings(ss []string) {
	for i := 0; i < len(ss); i++ {
		for j := i + 1; j < len(ss); j++ {
			if ss[j] < ss[i] {
				ss[i], ss[j] = ss[j], ss[i]
			}
		}
	}
}

func clientsFromContract(ct OpenAPIContract, override []string) ([]ClientSurface, error) {
	if len(override) > 0 {
		return NormalizeClients(override)
	}
	names := make([]string, 0, len(ct.ClientOps))
	for c := range ct.ClientOps {
		names = append(names, c)
	}
	sortStrings(names)
	if len(names) == 0 {
		return NormalizeClients(nil)
	}
	return NormalizeClients(names)
}

func clientFullCRUDFromOps(ops clientOperations) bool {
	return ops.Create || ops.Update || ops.Delete
}
