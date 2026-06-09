package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func TestRequireAppLayerForGRPC(t *testing.T) {
	dir := t.TempDir()
	opts := DBOptions{ModuleOptions: ModuleOptions{ProjectDir: dir}}
	if err := requireAppLayerForGRPC(dir, "sys_roles", opts); err == nil {
		t.Fatal("expected error without app layer")
	}
	meta := metaForTest("sys_roles", "sys_roles", "sys_roles")
	appDir := filepath.Join(dir, filepath.Dir(meta.appRel("application.go")))
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, meta.appRel("application.go")), []byte("package sys_roles\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := requireAppLayerForGRPC(dir, "sys_roles", opts); err != nil {
		t.Fatal(err)
	}
}

func TestRenderGRPCModuleCreatesStubsWhenProtoExists(t *testing.T) {
	dir := t.TempDir()
	projectModule := "example.com/demo12"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+projectModule+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	snake := "sys_roles"
	layoutMeta := metaForTest(snake, snake, snake)
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(layoutMeta.appRel("application.go"))), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, layoutMeta.appRel("application.go")), []byte("package sys_roles\n"), 0644); err != nil {
		t.Fatal(err)
	}
	protoDir := filepath.Join(dir, "api", "proto")
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(protoDir, snake+".proto"), []byte("syntax = \"proto3\";\n"), 0644); err != nil {
		t.Fatal(err)
	}
	bootstrapDir := filepath.Join(dir, "internal", "bootstrap")
	if err := os.MkdirAll(bootstrapDir, 0755); err != nil {
		t.Fatal(err)
	}
	grpcBootstrap := `package bootstrap

func RegisterGRPCServers(s interface{}, infra interface{}) {
	// grpc bootstrap modules (goeasy add db proto appends below)
}
`
	if err := os.WriteFile(filepath.Join(bootstrapDir, "grpc.go"), []byte(grpcBootstrap), 0644); err != nil {
		t.Fatal(err)
	}

	ct := schema.ClassifiedTable{
		ModuleName: snake,
		ReadCols: []schema.ColumnMeta{
			{Name: "id", DBType: "bigint"},
			{Name: "name", DBType: "varchar(64)"},
		},
		CreateCols: []schema.ColumnMeta{{Name: "name", DBType: "varchar(64)"}},
		UpdateCols: []schema.ColumnMeta{{Name: "name", DBType: "varchar(64)"}},
	}
	opts := DBOptions{ModuleOptions: ModuleOptions{ProjectDir: dir}}
	if err := renderGRPCModule(opts, projectModule, currentGoEasyModule(), ct); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{
		layoutMeta.grpcRel("server.go"),
		layoutMeta.grpcRel("handlers.go"),
		layoutMeta.grpcRel("convert.go"),
		filepath.Join("internal", "bootstrap", "register_"+snake+"_grpc.go"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
	b, err := os.ReadFile(filepath.Join(bootstrapDir, "grpc.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "RegisterSysRolesGRPC(s, infra)") {
		t.Fatalf("grpc.go missing bootstrap call: %s", string(b))
	}
}

func TestRenderGRPCModuleDomainLayoutRegisterImport(t *testing.T) {
	dir := t.TempDir()
	projectModule := "example.com/demo"
	snake := "sys_roles"
	layoutMeta := metaForTest(snake, "system", "roles")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+projectModule+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(layoutMeta.appRel("application.go"))), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, layoutMeta.appRel("application.go")), []byte("package roles\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "internal", "bootstrap"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal", "bootstrap", "grpc.go"), []byte(grpcBootstrapFixture()), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := `codegen:
  layout: domain
  domains:
    system:
      table_prefix: sys_
      modules:
        sys_roles:
          resource: roles
`
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	ct := schema.ClassifiedTable{
		ModuleName: snake,
		ReadCols:   []schema.ColumnMeta{{Name: "id", DBType: "bigint"}},
	}
	opts := DBOptions{
		ModuleOptions: ModuleOptions{
			ProjectDir: dir,
			ConfigPath: filepath.Join(dir, "configs", "config.yaml"),
		},
	}
	if err := renderGRPCModule(opts, projectModule, currentGoEasyModule(), ct); err != nil {
		t.Fatal(err)
	}
	regPath := filepath.Join(dir, "internal", "bootstrap", "register_"+snake+"_grpc.go")
	regBody, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}
	rs := string(regBody)
	wantImport := projectModule + "/internal/interface/grpc/system/roles"
	if !strings.Contains(rs, wantImport) {
		t.Fatalf("register import want %q in:\n%s", wantImport, rs)
	}
	serverBody, err := os.ReadFile(filepath.Join(dir, layoutMeta.grpcRel("server.go")))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(serverBody), "package roles") {
		t.Fatal("grpc server package must be resource name")
	}
}

func TestRenderGRPCModuleSkipsExistingServerWithoutForce(t *testing.T) {
	dir := t.TempDir()
	projectModule := "example.com/demo12"
	snake := "sys_roles"
	layoutMeta := metaForTest(snake, snake, snake)
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(layoutMeta.appRel("application.go"))), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, layoutMeta.appRel("application.go")), []byte("package sys_roles\n"), 0644); err != nil {
		t.Fatal(err)
	}
	serverDir := filepath.Join(dir, filepath.Dir(layoutMeta.grpcRel("server.go")))
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := "// keep me\n"
	if err := os.WriteFile(filepath.Join(serverDir, "server.go"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}
	bootstrapDir := filepath.Join(dir, "internal", "bootstrap")
	if err := os.MkdirAll(bootstrapDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bootstrapDir, "grpc.go"), []byte(grpcBootstrapFixture()), 0644); err != nil {
		t.Fatal(err)
	}

	ct := schema.ClassifiedTable{
		ModuleName: snake,
		ReadCols:   []schema.ColumnMeta{{Name: "id", DBType: "bigint"}},
	}
	opts := DBOptions{ModuleOptions: ModuleOptions{ProjectDir: dir}}
	if err := renderGRPCModule(opts, projectModule, currentGoEasyModule(), ct); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(serverDir, "server.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != existing {
		t.Fatalf("server.go overwritten without --force: %q", string(b))
	}
}

func grpcBootstrapFixture() string {
	return `package bootstrap

func RegisterGRPCServers(s interface{}, infra interface{}) {
	// grpc bootstrap modules (goeasy add db proto appends below)
}
`
}
