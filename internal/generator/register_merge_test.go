package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeRegisterDomainSecondModule(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := `database:
  driver: postgres
codegen:
  layout: domain
  domains:
    system:
      table_prefix: sys_
      modules:
        sys_roles:
          resource: roles
        sys_menus:
          resource: menus
`
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}

	opts := func(module string) ModuleOptions {
		return ModuleOptions{
			ProjectDir: dir,
			ModuleName: module,
			Clients:    []string{"admin"},
			ConfigPath: "configs/config.yaml",
		}
	}
	if err := GenerateCRUD(opts("sys_roles")); err != nil {
		t.Fatal(err)
	}
	reg := filepath.Join(dir, "internal", "bootstrap", "register_system.go")
	b1, err := os.ReadFile(reg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b1), "func registerSystemSysRoles(") {
		t.Fatal("expected per-module register helper for sys_roles")
	}

	if err := GenerateCRUD(opts("sys_menus")); err != nil {
		t.Fatalf("second module should merge: %v", err)
	}
	b2, err := os.ReadFile(reg)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b2)
	if !strings.Contains(s, "goeasy-module: sys_roles") {
		t.Fatal("sys_roles marker missing")
	}
	if !strings.Contains(s, "goeasy-module: sys_menus") {
		t.Fatal("sys_menus marker missing")
	}
	if strings.Count(s, "func RegisterSystem(") != 1 {
		t.Fatal("RegisterSystem should appear once")
	}
	if strings.Count(s, `engine.Group("/api/v1/admin"`) != 1 {
		t.Fatal("admin router group should be created once per domain")
	}
	if !strings.Contains(s, "func registerSystemSysMenus(") {
		t.Fatal("expected per-module register helper for sys_menus")
	}
	if !strings.Contains(s, `example.com/demo/internal/interface/http/admin/system/roles`) {
		t.Fatal("http import must include project module path")
	}
	if strings.Contains(s, " := .NewHandler(") || strings.Contains(s, "\n\t.NewHandler(") {
		t.Fatalf("ImportAlias must not be empty (got bare .NewHandler): %s", s)
	}
	if !strings.Contains(s, "adminsysroles.NewHandler(") {
		t.Fatal("expected adminsysroles import alias on NewHandler")
	}
	if !strings.Contains(s, "adminsysmenus.NewHandler(") {
		t.Fatal("expected adminsysmenus import alias on NewHandler")
	}
}

func TestGenerateCRUDIdempotentNoForce(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte("database:\n  driver: postgres\n"), 0644); err != nil {
		t.Fatal(err)
	}
	opts := ModuleOptions{ProjectDir: dir, ModuleName: "orders", Clients: []string{"admin"}}
	if err := GenerateCRUD(opts); err != nil {
		t.Fatal(err)
	}
	if err := GenerateCRUD(opts); err != nil {
		t.Fatalf("second run without --force should not error: %v", err)
	}
}

func TestGenerateEventIdempotentNoForce(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	opts := ModuleOptions{ProjectDir: dir, ModuleName: "order-paid", Domain: "order"}
	if err := GenerateEvent(opts); err != nil {
		t.Fatal(err)
	}
	if err := GenerateEvent(opts); err != nil {
		t.Fatalf("second add event should skip, not error: %v", err)
	}
}
