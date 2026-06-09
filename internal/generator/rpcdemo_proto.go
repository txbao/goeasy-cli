package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

const defaultRPCDemoProto = "sys_roles"

// resolveRPCDemoProtoModule 解析对端 proto 模块名：显式 --proto > --from-url 文件名推断 > sys_roles。
func resolveRPCDemoProtoModule(protoModule, fromURL string) string {
	if strings.TrimSpace(protoModule) != "" {
		return utils.ToSnake(protoModule)
	}
	if inferred := inferProtoModuleFromURL(fromURL); inferred != "" {
		return inferred
	}
	return defaultRPCDemoProto
}

func inferProtoModuleFromURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	name := filepath.Base(rawURL)
	if idx := strings.Index(name, "?"); idx >= 0 {
		name = name[:idx]
	}
	name = strings.TrimSuffix(strings.ToLower(name), ".proto")
	if name == "" || name == "." || name == "remote" {
		return ""
	}
	return utils.ToSnake(name)
}

// RPCDemoProtoMeta 跨服务 gRPC 契约元数据（由 proto 文件或模块名推导）。
type RPCDemoProtoMeta struct {
	Module       string // sys_roles
	Pascal       string // SysRoles
	PbAlias      string // sys_rolespb
	ImportPath   string // {module}/api/proto/gen/imported/sys_roles
	GatewayName  string // SysRolesGateway
	ViewName     string // RoleView
	ViewFields   []rpcViewField
	CreateFields []rpcViewField
	ProtoRel     string // api/proto/imported/sys_roles.proto
	GenDir       string // api/proto/gen/imported/sys_roles
	GatewayFile  string // internal/infrastructure/rpc/{remote}/sys_roles_gateway.go
}

type rpcViewField struct {
	GoName    string // ID, CreatedAt
	JSONName  string // id, created_at
	PbGetter  string // GetId, GetCreatedAt
	ProtoType string // int64, int32, string, bool
	GoType    string // int64, int32, string, bool
}

var rpcDemoProtoPresets = map[string]struct {
	viewName string
	fields   []rpcViewField
}{
	"sys_roles": {
		viewName: "RoleView",
		fields: []rpcViewField{
			{"ID", "id", "GetId", "int64", "int64"},
			{"Name", "name", "GetName", "string", "string"},
			{"Code", "code", "GetCode", "string", "string"},
		},
	},
	"sys_apis": {
		viewName: "ApiView",
		fields: []rpcViewField{
			{"ID", "id", "GetId", "int64", "int64"},
			{"CreatedAt", "created_at", "GetCreatedAt", "string", "string"},
			{"UpdatedAt", "updated_at", "GetUpdatedAt", "string", "string"},
			{"DeletedAt", "deleted_at", "GetDeletedAt", "string", "string"},
			{"Name", "name", "GetName", "string", "string"},
			{"Path", "path", "GetPath", "string", "string"},
			{"Method", "method", "GetMethod", "string", "string"},
			{"Module", "module", "GetModule", "string", "string"},
			{"Description", "description", "GetDescription", "string", "string"},
			{"IsPublic", "is_public", "GetIsPublic", "int32", "int32"},
			{"Status", "status", "GetStatus", "int32", "int32"},
			{"Version", "version", "GetVersion", "int32", "int32"},
		},
	},
}

func resolveRPCDemoProtoMeta(projectModule, projectDir, protoModule, remoteService, fromURL string) RPCDemoProtoMeta {
	snake := utils.ToSnake(protoModule)
	pascal := utils.ToPascal(snake)
	viewName := pascal + "View"
	viewFields := []rpcViewField{
		{"ID", "id", "GetId", "int64", "int64"},
		{"Name", "name", "GetName", "string", "string"},
	}
	createFields := []rpcViewField{
		{"Name", "name", "GetName", "string", "string"},
	}

	if preset, ok := rpcDemoProtoPresets[snake]; ok {
		viewName = preset.viewName
		viewFields = append([]rpcViewField(nil), preset.fields...)
	}

	if contract, ok := loadRPCDemoProtoContract(projectDir, snake, remoteService, fromURL); ok {
		pascal = contract.ModulePascal
		if len(contract.CT.ReadCols) > 0 {
			viewFields = columnsToRPCViewFields(contract.CT.ReadCols)
		}
		if len(contract.CT.CreateCols) > 0 {
			createFields = columnsToRPCViewFields(contract.CT.CreateCols)
		}
	} else if preset, ok := rpcDemoProtoPresets[snake]; ok {
		createFields = presetCreateFields(preset.fields)
	}

	return RPCDemoProtoMeta{
		Module:       snake,
		Pascal:       pascal,
		PbAlias:      strings.ReplaceAll(snake, "-", "_") + "pb",
		ImportPath:   projectModule + "/api/proto/gen/imported/" + snake,
		GatewayName:  pascal + "Gateway",
		ViewName:     viewName,
		ViewFields:   viewFields,
		CreateFields: createFields,
		ProtoRel:     filepath.ToSlash(filepath.Join("api", "proto", "imported", snake+".proto")),
		GenDir:       filepath.ToSlash(filepath.Join("api", "proto", "gen", "imported", snake)),
		GatewayFile:  filepath.ToSlash(filepath.Join("internal", "infrastructure", "rpc", remoteService, snake+"_gateway.go")),
	}
}

func presetCreateFields(viewFields []rpcViewField) []rpcViewField {
	skip := map[string]bool{"id": true, "created_at": true, "updated_at": true, "deleted_at": true}
	var out []rpcViewField
	for _, f := range viewFields {
		if skip[f.JSONName] {
			continue
		}
		out = append(out, f)
	}
	return out
}

func loadRPCDemoProtoContract(projectDir, protoModule, remoteService, fromURL string) (ProtoContract, bool) {
	for _, path := range rpcDemoProtoSourceCandidates(projectDir, protoModule, remoteService, fromURL) {
		if st, err := os.Stat(path); err != nil || st.IsDir() {
			continue
		}
		contract, err := parseProtoContract(path)
		if err != nil {
			continue
		}
		return contract, true
	}
	return ProtoContract{}, false
}

func rpcDemoProtoSourceCandidates(projectDir, protoModule, remoteService, fromURL string) []string {
	var out []string
	if fromURL = strings.TrimSpace(fromURL); fromURL != "" {
		out = append(out, fromURL)
	}
	out = append(out,
		filepath.Join(projectDir, "api", "proto", "imported", protoModule+".proto"),
		filepath.Join(projectDir, "api", "proto", protoModule+".proto"),
	)
	if discovered := discoverRPCDemoProtoURL(projectDir, remoteService, protoModule); discovered != "" {
		out = append(out, discovered)
	}
	return out
}

func columnsToRPCViewFields(cols []schema.ColumnMeta) []rpcViewField {
	out := make([]rpcViewField, 0, len(cols))
	for _, c := range cols {
		out = append(out, columnToRPCViewField(c))
	}
	return out
}

func columnToRPCViewField(c schema.ColumnMeta) rpcViewField {
	goField := schema.GoFieldFromColumn(c)
	protoTyp := protoScalarType(c.DBType)
	return rpcViewField{
		GoName:    goField.Name,
		JSONName:  c.Name,
		PbGetter:  protoGetterFromField(c.Name),
		ProtoType: protoTyp,
		GoType:    protoTypeToGoType(protoTyp),
	}
}

// protoGetterFromField 对齐 protoc-gen-go 的 Getter 命名（如 id → GetId，is_public → GetIsPublic）。
func protoGetterFromField(protoField string) string {
	parts := strings.Split(protoField, "_")
	var b strings.Builder
	b.WriteString("Get")
	for _, p := range parts {
		if p == "" {
			continue
		}
		if strings.EqualFold(p, "id") {
			b.WriteString("Id")
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + strings.ToLower(p[1:]))
	}
	return b.String()
}

func protoScalarType(dbType string) string {
	t := strings.ToLower(dbType)
	switch {
	case strings.Contains(t, "bigint"), strings.Contains(t, "int8"):
		return "int64"
	case strings.Contains(t, "integer"), strings.Contains(t, "int4"), strings.Contains(t, "int32"):
		return "int32"
	case strings.Contains(t, "bool"):
		return "bool"
	default:
		return "string"
	}
}

func protoTypeToGoType(protoTyp string) string {
	switch protoTyp {
	case "int64":
		return "int64"
	case "int32":
		return "int32"
	case "bool":
		return "bool"
	default:
		return "string"
	}
}

func rpcDemoProtoPBAvailable(projectDir string, meta RPCDemoProtoMeta) bool {
	dir := filepath.Join(projectDir, filepath.FromSlash(meta.GenDir))
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return false
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_grpc.pb.go") {
			return true
		}
	}
	return false
}

func ensureRPCDemoProto(opts RPCDemoOptions, projectModule string, meta RPCDemoProtoMeta) error {
	if rpcDemoProtoPBAvailable(opts.ProjectDir, meta) {
		return nil
	}
	protoPath := filepath.Join(opts.ProjectDir, filepath.FromSlash(meta.ProtoRel))
	if _, err := os.Stat(protoPath); err == nil {
		maybeRunGenProtoAfterDB(opts.ProjectDir, []string{meta.ProtoRel}, opts.SkipFetchProto)
		if rpcDemoProtoPBAvailable(opts.ProjectDir, meta) {
			return nil
		}
	}
	fromURL := strings.TrimSpace(opts.FromURL)
	if fromURL == "" {
		fromURL = discoverRPCDemoProtoURL(opts.ProjectDir, opts.RemoteService, meta.Module)
	}
	if fromURL != "" && !opts.SkipFetchProto {
		if err := GenerateProtoGo(GenProtoOptions{
			ProjectDir: opts.ProjectDir,
			FromURL:    fromURL,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "warn: auto fetch/gen proto from %s: %v\n", fromURL, err)
		} else if rpcDemoProtoPBAvailable(opts.ProjectDir, meta) {
			return nil
		}
	}
	if opts.SkipFetchProto {
		fmt.Fprintf(os.Stderr, "info: next: goeasy-cli gen proto --file %s\n", meta.ProtoRel)
	}
	return fmt.Errorf("missing gRPC pb package %s; fetch remote proto first, e.g.:\n  goeasy-cli gen proto --from-url <path-to/%s.proto>\n  goeasy-cli add rpcdemo --remote %s --proto %s --force",
		meta.ImportPath, meta.Module, opts.RemoteService, meta.Module)
}

func discoverRPCDemoProtoURL(projectDir, remoteService, protoModule string) string {
	candidates := []string{
		filepath.Join(projectDir, "..", remoteService, "api", "proto", protoModule+".proto"),
		filepath.Join(projectDir, "..", remoteService, "api", "proto", "imported", protoModule+".proto"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, err := filepath.Abs(c)
			if err != nil {
				return c
			}
			return abs
		}
	}
	return ""
}

func enrichRPCDemoData(data map[string]any, meta RPCDemoProtoMeta) {
	data["ProtoModule"] = meta.Module
	data["ProtoPascal"] = meta.Pascal
	data["ProtoPbAlias"] = meta.PbAlias
	data["ProtoImportPath"] = meta.ImportPath
	data["GatewayName"] = meta.GatewayName
	data["ViewName"] = meta.ViewName
}
