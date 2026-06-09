package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateRPCDemo_service(t *testing.T) {
	dir := t.TempDir()
	writeMinimalGoMod(t, dir, "example.com/rpctest")
	writeRPCDemoConfig(t, dir, "service")
	writeRPCDemoProtoPB(t, dir, "sys_roles")

	if err := GenerateRPCDemo(RPCDemoOptions{
		ProjectDir:     dir,
		RemoteService:  "user",
		ProtoModule:    "sys_roles",
		SkipFetchProto: true,
		Force:          true,
	}); err != nil {
		t.Fatal(err)
	}
	assertRPCDemoCommonFiles(t, dir, "user", "sys_roles_gateway.go")

	app, err := os.ReadFile(filepath.Join(dir, "internal/app/rpcdemo/application.go"))
	if err != nil {
		t.Fatal(err)
	}
	appSrc := string(app)
	if strings.Contains(appSrc, "Queries()") {
		t.Fatalf("service style application must not expose Queries()")
	}
	if !strings.Contains(appSrc, "func (a *Application) Get(") {
		t.Fatalf("service style application should have Get method")
	}
	if !strings.Contains(appSrc, "*port.RoleView") {
		t.Fatalf("Get should return port view")
	}

	handler, err := os.ReadFile(filepath.Join(dir, "internal/interface/http/admin/rpcdemo/handler.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(handler), "h.app.Get(") {
		t.Fatalf("service style handler should call h.app.Get")
	}
	if !strings.Contains(string(handler), "ToResponse(view)") {
		t.Fatalf("handler should map view to response DTO")
	}
}

func TestGenerateRPCDemo_light_cqrs(t *testing.T) {
	dir := t.TempDir()
	writeMinimalGoMod(t, dir, "example.com/rpctest")
	writeRPCDemoConfig(t, dir, "light_cqrs")
	writeRPCDemoProtoPB(t, dir, "sys_roles")

	if err := GenerateRPCDemo(RPCDemoOptions{
		ProjectDir:     dir,
		RemoteService:  "demo1",
		ProtoModule:    "sys_roles",
		AppStyle:       "light_cqrs",
		SkipFetchProto: true,
		Force:          true,
	}); err != nil {
		t.Fatal(err)
	}
	assertRPCDemoCommonFiles(t, dir, "demo1", "sys_roles_gateway.go")

	app, err := os.ReadFile(filepath.Join(dir, "internal/app/rpcdemo/application.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(app), "Queries()") {
		t.Fatalf("light_cqrs application should expose Queries()")
	}

	cmd, err := os.ReadFile(filepath.Join(dir, "internal/app/rpcdemo/command/create.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cmd), "Create(ctx context.Context, cmd CreateCommand) (*port.RoleView") {
		t.Fatalf("command create should return port view via gateway")
	}
}

func TestGenerateRPCDemo_sys_apis_proto(t *testing.T) {
	dir := t.TempDir()
	writeMinimalGoMod(t, dir, "example.com/rpctest")
	writeRPCDemoConfig(t, dir, "service")
	writeRPCDemoProtoPB(t, dir, "sys_apis")
	writeRPCDemoSysApisProto(t, dir)

	if err := GenerateRPCDemo(RPCDemoOptions{
		ProjectDir:     dir,
		RemoteService:  "demo1",
		ProtoModule:    "sys_apis",
		SkipFetchProto: true,
		Force:          true,
	}); err != nil {
		t.Fatal(err)
	}
	gw, err := os.ReadFile(filepath.Join(dir, "internal/infrastructure/rpc/demo1/sys_apis_gateway.go"))
	if err != nil {
		t.Fatal(err)
	}
	gwSrc := string(gw)
	if !strings.Contains(gwSrc, "api/proto/gen/imported/sys_apis") {
		t.Fatalf("gateway should import sys_apis pb package")
	}
	if !strings.Contains(gwSrc, "SysApisGateway") {
		t.Fatalf("gateway should be SysApisGateway")
	}
	if !strings.Contains(gwSrc, "CreateSysApis") {
		t.Fatalf("gateway should call CreateSysApis")
	}
	if !strings.Contains(gwSrc, "out.GetId()") {
		t.Fatalf("gateway should use protoc GetId() getter")
	}
	if strings.Contains(gwSrc, "out.GetID()") {
		t.Fatalf("gateway must not use GetID()")
	}
	if !strings.Contains(gwSrc, "}, nil\n}\n\nfunc (g *SysApisGateway) Create") {
		t.Fatalf("GetByID must close with } before Create:\n%s", gwSrc)
	}
	if !strings.HasSuffix(strings.TrimSpace(gwSrc), "}") {
		t.Fatalf("Create must end with closing }")
	}

	dto, err := os.ReadFile(filepath.Join(dir, "internal/interface/http/admin/rpcdemo/dto.go"))
	if err != nil {
		t.Fatal(err)
	}
	dtoSrc := string(dto)
	for _, field := range []string{`json:"is_public"`, `json:"description"`, `json:"version"`, `json:"created_at"`} {
		if !strings.Contains(dtoSrc, field) {
			t.Fatalf("dto missing field tag %s:\n%s", field, dtoSrc)
		}
	}

	portSrc, err := os.ReadFile(filepath.Join(dir, "internal/app/rpcdemo/port/port.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(portSrc), "Create(ctx context.Context, in *CreateInput)") {
		t.Fatalf("port should define Create method")
	}
}

func TestProtoGetterFromField(t *testing.T) {
	cases := map[string]string{
		"id":          "GetId",
		"name":        "GetName",
		"created_at":  "GetCreatedAt",
		"is_public":   "GetIsPublic",
		"category_id": "GetCategoryId",
	}
	for field, want := range cases {
		if got := protoGetterFromField(field); got != want {
			t.Fatalf("protoGetterFromField(%q) = %q, want %q", field, got, want)
		}
	}
}

func TestResolveRPCDemoProtoModule(t *testing.T) {
	cases := []struct {
		proto, fromURL, want string
	}{
		{"", "", defaultRPCDemoProto},
		{"", `D:/dev/demo1/api/proto/sys_apis.proto`, "sys_apis"},
		{"", "https://example.com/contracts/sys_roles.proto", "sys_roles"},
		{"sys_menus", `D:/x/sys_apis.proto`, "sys_menus"},
	}
	for _, tc := range cases {
		got := resolveRPCDemoProtoModule(tc.proto, tc.fromURL)
		if got != tc.want {
			t.Fatalf("resolveRPCDemoProtoModule(%q, %q) = %q, want %q", tc.proto, tc.fromURL, got, tc.want)
		}
	}
}

func TestResolveRPCDemoProtoMeta_from_proto_file(t *testing.T) {
	dir := t.TempDir()
	writeRPCDemoSysApisProto(t, dir)

	meta := resolveRPCDemoProtoMeta("example.com/demo", dir, "sys_apis", "demo1", "")
	if len(meta.ViewFields) < 10 {
		t.Fatalf("expected full SysApis fields, got %d", len(meta.ViewFields))
	}
	if len(meta.CreateFields) < 6 {
		t.Fatalf("expected create request fields, got %d", len(meta.CreateFields))
	}
	if meta.CreateFields[0].JSONName != "name" {
		t.Fatalf("first create field should be name, got %s", meta.CreateFields[0].JSONName)
	}
}

func TestGenerateRPCDemo_infer_proto_from_url(t *testing.T) {
	dir := t.TempDir()
	writeMinimalGoMod(t, dir, "example.com/rpctest")
	writeRPCDemoConfig(t, dir, "service")
	writeRPCDemoProtoPB(t, dir, "sys_apis")

	protoSrc := filepath.Join(dir, "remote", "sys_apis.proto")
	if err := os.MkdirAll(filepath.Dir(protoSrc), 0755); err != nil {
		t.Fatal(err)
	}
	writeRPCDemoSysApisProtoContent(t, protoSrc)
	if err := GenerateRPCDemo(RPCDemoOptions{
		ProjectDir:     dir,
		RemoteService:  "demo1",
		FromURL:        protoSrc,
		SkipFetchProto: true,
		Force:          true,
	}); err != nil {
		t.Fatal(err)
	}
	gw, err := os.ReadFile(filepath.Join(dir, "internal/infrastructure/rpc/demo1/sys_apis_gateway.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gw), "SysApisGateway") {
		t.Fatalf("should infer sys_apis from --from-url, got:\n%s", gw)
	}
}

func TestGenerateRPCDemo_missing_proto(t *testing.T) {
	dir := t.TempDir()
	writeMinimalGoMod(t, dir, "example.com/rpctest")

	err := GenerateRPCDemo(RPCDemoOptions{
		ProjectDir:     dir,
		RemoteService:  "demo1",
		ProtoModule:    "sys_roles",
		SkipFetchProto: true,
		Force:          true,
	})
	if err == nil {
		t.Fatal("expected error when proto pb is missing")
	}
	if !strings.Contains(err.Error(), "missing gRPC pb package") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertRPCDemoCommonFiles(t *testing.T, dir, remote, gatewayFile string) {
	t.Helper()
	want := []string{
		"internal/app/rpcdemo/port/port.go",
		"internal/app/rpcdemo/query/get.go",
		"internal/bootstrap/register_rpcdemo.go",
		"internal/interface/http/admin/rpcdemo/dto.go",
		filepath.Join("internal", "infrastructure", "rpc", remote, gatewayFile),
	}
	for _, rel := range want {
		if rel == "internal/app/rpcdemo/query/get.go" {
			continue // service style omits query/get.go
		}
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, "internal/app/rpcdemo/port/port.go"))
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	if !strings.Contains(src, "GetByID") {
		t.Fatalf("port should define GetByID")
	}
	if !strings.Contains(src, "Create(") {
		t.Fatalf("port should define Create")
	}
}

func writeRPCDemoConfig(t *testing.T, dir, appStyle string) {
	t.Helper()
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := "codegen:\n  app_style: " + appStyle + "\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeRPCDemoProtoPB(t *testing.T, dir, module string) {
	t.Helper()
	genDir := filepath.Join(dir, "api", "proto", "gen", "imported", module)
	if err := os.MkdirAll(genDir, 0755); err != nil {
		t.Fatal(err)
	}
	grpcFile := filepath.Join(genDir, module+"_grpc.pb.go")
	if err := os.WriteFile(grpcFile, []byte("package stub\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeRPCDemoSysApisProto(t *testing.T, dir string) {
	t.Helper()
	protoDir := filepath.Join(dir, "api", "proto", "imported")
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeRPCDemoSysApisProtoContent(t, filepath.Join(protoDir, "sys_apis.proto"))
}

func writeRPCDemoSysApisProtoContent(t *testing.T, path string) {
	t.Helper()
	content := `syntax = "proto3";
package sys_apis;

service SysApisService {
  rpc GetSysApis (GetSysApisRequest) returns (SysApis);
  rpc CreateSysApis (CreateSysApisRequest) returns (SysApis);
}

message SysApis {
  int64 id = 1;
  string created_at = 2;
  string updated_at = 3;
  string deleted_at = 4;
  string name = 5;
  string path = 6;
  string method = 7;
  string module = 8;
  string description = 9;
  int32 is_public = 10;
  int32 status = 11;
  int32 version = 12;
}

message GetSysApisRequest {
  string id = 1;
}

message CreateSysApisRequest {
  string name = 1;
  string path = 2;
  string method = 3;
  string module = 4;
  string description = 5;
  int32 is_public = 6;
  int32 status = 7;
  int32 version = 8;
}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeMinimalGoMod(t *testing.T, dir, module string) {
	t.Helper()
	content := "module " + module + "\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
