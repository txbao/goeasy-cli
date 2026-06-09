package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func TestGenGRPCHandlersServiceStyle(t *testing.T) {
	meta := metaForTest("sys_roles", "system", "roles")
	cols := []GRPCCol{{Name: "name", Pascal: "Name", ProtoField: "Name"}}
	out := genGRPCHandlers(
		"github.com/org/demo1",
		"github.com/txbao/goeasy",
		"SysRoles",
		"sysroles",
		meta,
		"sys_roles",
		AppStyleService,
		cols,
		cols,
		cols,
	)
	if strings.Contains(out, "/command\"") {
		t.Fatal("service style must not import command subpackage")
	}
	if !strings.Contains(out, "s.app.Get(ctx,") {
		t.Fatal("service style must call s.app.Get")
	}
	if !strings.Contains(out, "s.app.Create(ctx, cmd)") {
		t.Fatal("service style must call s.app.Create")
	}
	if strings.Contains(out, "Commands()") || strings.Contains(out, "Queries()") {
		t.Fatal("service style must not use CQRS facade")
	}
	if !strings.Contains(out, "sysrolesapp.CreateCommand") {
		t.Fatal("service style must use app package CreateCommand")
	}
}

func TestGenGRPCHandlersLightCQRSStyle(t *testing.T) {
	meta := metaForTest("sys_roles", "system", "roles")
	cols := []GRPCCol{{Name: "name", Pascal: "Name", ProtoField: "Name"}}
	out := genGRPCHandlers(
		"github.com/org/demo1",
		"github.com/txbao/goeasy",
		"SysRoles",
		"sysroles",
		meta,
		"sys_roles",
		AppStyleLightCQRS,
		cols,
		cols,
		cols,
	)
	if !strings.Contains(out, "/command\"") {
		t.Fatal("light_cqrs style must import command subpackage")
	}
	if !strings.Contains(out, "s.app.Queries().Get") {
		t.Fatal("light_cqrs style must use Queries().Get")
	}
	if !strings.Contains(out, "s.app.Commands().Create") {
		t.Fatal("light_cqrs style must use Commands().Create")
	}
	if strings.Contains(out, "sysrolesapp.CreateCommand") {
		t.Fatal("light_cqrs style must use command.CreateCommand")
	}
}

func TestRenderGRPCModuleServiceStyleHandlers(t *testing.T) {
	dir := t.TempDir()
	projectModule := "example.com/demo"
	snake := "sys_roles"
	layoutMeta := metaForTest(snake, "system", "roles")
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(layoutMeta.appRel("application.go"))), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+projectModule+"\n"), 0644); err != nil {
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
  app_style: service
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
		CreateCols: []schema.ColumnMeta{{Name: "name", DBType: "varchar(64)"}},
		UpdateCols: []schema.ColumnMeta{{Name: "name", DBType: "varchar(64)"}},
	}
	opts := DBOptions{
		ModuleOptions: ModuleOptions{
			ProjectDir: dir,
			ConfigPath: filepath.Join(dir, "configs", "config.yaml"),
			Force:      true,
		},
	}
	if err := renderGRPCModule(opts, projectModule, currentGoEasyModule(), ct); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, layoutMeta.grpcRel("handlers.go")))
	if err != nil {
		t.Fatal(err)
	}
	hs := string(body)
	if strings.Contains(hs, "/command\"") {
		t.Fatalf("handlers must match service app_style:\n%s", hs)
	}
	if !strings.Contains(hs, "s.app.Get(ctx,") {
		t.Fatalf("handlers must use s.app.Get:\n%s", hs)
	}
}
