package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateRPCClient_all_methods(t *testing.T) {
	dir := t.TempDir()
	writeMinimalGoMod(t, dir, "example.com/rpctest")
	writeRPCDemoProtoPB(t, dir, "sys_roles")
	writeSysRolesProto(t, dir)

	if err := GenerateRPCClient(RPCClientOptions{
		ProjectDir:     dir,
		RemoteService:  "demo1",
		ProtoModule:    "sys_roles",
		SkipFetchProto: true,
		Methods:        "all",
		Force:          true,
	}); err != nil {
		t.Fatal(err)
	}

	portPath := filepath.Join(dir, "internal/infrastructure/rpc/demo1/port/sys_roles.go")
	portSrc, err := os.ReadFile(portPath)
	if err != nil {
		t.Fatal(err)
	}
	ps := string(portSrc)
	for _, m := range []string{"GetByID", "Create(", "Update(", "Delete(", "List("} {
		if !strings.Contains(ps, m) {
			t.Fatalf("shared port missing %s:\n%s", m, ps)
		}
	}
	if !strings.Contains(ps, "type UpdateInput struct") {
		t.Fatalf("shared port should define UpdateInput")
	}

	gwPath := filepath.Join(dir, "internal/infrastructure/rpc/demo1/sys_roles_gateway.go")
	gwSrc, err := os.ReadFile(gwPath)
	if err != nil {
		t.Fatal(err)
	}
	gs := string(gwSrc)
	if strings.Contains(gs, "app/rpcdemo/port") {
		t.Fatal("gateway must import shared port, not rpcdemo port")
	}
	if !strings.Contains(gs, "infrastructure/rpc/demo1/port") {
		t.Fatalf("gateway should import shared port package")
	}
	for _, m := range []string{"ListSysRoles", "UpdateSysRoles", "DeleteSysRoles"} {
		if !strings.Contains(gs, m) {
			t.Fatalf("gateway missing %s", m)
		}
	}
}

func TestGenerateRPCClientBind_wire(t *testing.T) {
	dir := t.TempDir()
	writeMinimalGoMod(t, dir, "example.com/rpctest")
	writeRPCDemoProtoPB(t, dir, "sys_roles")
	writeSysRolesProto(t, dir)
	writeEtcddemoRegister(t, dir)

	if err := GenerateRPCClient(RPCClientOptions{
		ProjectDir:     dir,
		RemoteService:  "demo1",
		ProtoModule:    "sys_roles",
		SkipFetchProto: true,
		Force:          true,
	}); err != nil {
		t.Fatal(err)
	}

	if err := GenerateRPCClientBind(RPCClientBindOptions{
		ProjectDir:  dir,
		ProtoModule: "sys_roles",
		Consumer:    "etcddemo",
		Wire:        true,
		Force:       true,
	}); err != nil {
		t.Fatal(err)
	}

	aliasPath := filepath.Join(dir, "internal/app/etcddemo/etcddemo/port/sys_roles.go")
	aliasSrc, err := os.ReadFile(aliasPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(aliasSrc), "demo1port") {
		t.Fatalf("consumer port should alias shared port")
	}

	regPath := filepath.Join(dir, "internal/bootstrap/register_etcddemo.go")
	regSrc, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}
	rs := string(regSrc)
	if !strings.Contains(rs, "// goeasy-rpc-wire: sys_roles") {
		t.Fatalf("register should contain wire marker")
	}
	if !strings.Contains(rs, "RPCClientLazy(infra, \"demo1\")") {
		t.Fatalf("register should dial demo1")
	}
	if !strings.Contains(rs, "NewSysRolesGateway(cli)") {
		t.Fatalf("register should construct gateway")
	}
}

func TestGenerateRPCClientBind_idempotent_wire(t *testing.T) {
	dir := t.TempDir()
	writeMinimalGoMod(t, dir, "example.com/rpctest")
	writeRPCDemoProtoPB(t, dir, "sys_roles")
	writeSysRolesProto(t, dir)
	writeEtcddemoRegister(t, dir)

	opts := RPCClientOptions{
		ProjectDir: dir, RemoteService: "demo1", ProtoModule: "sys_roles", SkipFetchProto: true, Force: true,
	}
	if err := GenerateRPCClient(opts); err != nil {
		t.Fatal(err)
	}
	bind := RPCClientBindOptions{ProjectDir: dir, ProtoModule: "sys_roles", Consumer: "etcddemo", Wire: true, Force: true}
	if err := GenerateRPCClientBind(bind); err != nil {
		t.Fatal(err)
	}
	b1, _ := os.ReadFile(filepath.Join(dir, "internal/bootstrap/register_etcddemo.go"))
	if err := GenerateRPCClientBind(bind); err != nil {
		t.Fatal(err)
	}
	b2, _ := os.ReadFile(filepath.Join(dir, "internal/bootstrap/register_etcddemo.go"))
	if strings.Count(string(b2), "// goeasy-rpc-wire: sys_roles") != strings.Count(string(b1), "// goeasy-rpc-wire: sys_roles") {
		t.Fatal("second bind should not duplicate wire marker")
	}
}

func writeSysRolesProto(t *testing.T, dir string) {
	t.Helper()
	protoDir := filepath.Join(dir, "api", "proto", "imported")
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `syntax = "proto3";
package sys_roles;

service SysRolesService {
  rpc ListSysRoles (ListSysRolesRequest) returns (ListSysRolesResponse);
  rpc GetSysRoles (GetSysRolesRequest) returns (SysRoles);
  rpc CreateSysRoles (CreateSysRolesRequest) returns (SysRoles);
  rpc UpdateSysRoles (UpdateSysRolesRequest) returns (SysRoles);
  rpc DeleteSysRoles (DeleteSysRolesRequest) returns (DeleteSysRolesResponse);
}

message SysRoles {
  int64 id = 1;
  string name = 2;
  string code = 3;
}

message ListSysRolesRequest { int32 page = 1; int32 page_size = 2; }
message ListSysRolesResponse { repeated SysRoles list = 1; int64 total = 2; int32 page = 3; int32 page_size = 4; int32 total_pages = 5; }
message GetSysRolesRequest { string id = 1; }
message CreateSysRolesRequest { string name = 1; string code = 2; }
message UpdateSysRolesRequest { string id = 1; string name = 2; string code = 3; }
message DeleteSysRolesRequest { string id = 1; }
message DeleteSysRolesResponse { bool ok = 1; }
`
	if err := os.WriteFile(filepath.Join(protoDir, "sys_roles.proto"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeEtcddemoRegister(t *testing.T, dir string) {
	t.Helper()
	bootstrapDir := filepath.Join(dir, "internal", "bootstrap")
	if err := os.MkdirAll(bootstrapDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `package bootstrap

import (
	"github.com/gin-gonic/gin"
	goeasyapp "github.com/txbao/goeasy/app"
	etcddemoapp "example.com/rpctest/internal/app/etcddemo/etcddemo"
	etcddemodomain "example.com/rpctest/internal/domain/etcddemo/etcddemo"
	etcddemoinfra "example.com/rpctest/internal/infrastructure/etcddemo/persistence/etcddemo"
	adminetcddemo "example.com/rpctest/internal/interface/http/admin/etcddemo/etcddemo"
)

func RegisterEtcddemo(engine *gin.Engine, infra goeasyapp.HTTPInfra) error {
	apiAdmin := engine.Group("/api/v1/admin")
	return registerEtcddemoEtcddemo(apiAdmin, infra)
}

// goeasy-module: etcddemo
func registerEtcddemoEtcddemo(apiAdmin *gin.RouterGroup, infra goeasyapp.HTTPInfra) error {
	repo := etcddemoinfra.NewRepository()
	dom := etcddemodomain.NewDomainService()
	application := etcddemoapp.NewApplication(repo, dom)
	hAdmin := adminetcddemo.NewHandler(application)
	adminetcddemo.RegisterRoutes(apiAdmin, hAdmin)
	return nil
}
`
	if err := os.WriteFile(filepath.Join(bootstrapDir, "register_etcddemo.go"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
