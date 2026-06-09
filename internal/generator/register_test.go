package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateModuleCreatesRegisterAndModules(t *testing.T) {
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

	if err := GenerateModule(ModuleOptions{ProjectDir: dir, ModuleName: "sys_roles"}); err != nil {
		t.Fatal(err)
	}
	reg := filepath.Join(dir, "internal", "bootstrap", "register_sys_roles.go")
	if _, err := os.Stat(reg); err != nil {
		t.Fatalf("register file: %v", err)
	}
	b, _ := os.ReadFile(reg)
	if strings.Contains(string(b), "RegisterCRUDRoutes") {
		t.Fatal("module-only register should not include CRUD routes")
	}
	modules := filepath.Join(dir, "internal", "bootstrap", "modules.go")
	mb, err := os.ReadFile(modules)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mb), "RegisterSysRoles(engine, infra)") {
		t.Fatal("modules.go should register domain bootstrap")
	}
}

func TestGenerateCRUDPublicH5Register(t *testing.T) {
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
	if err := GenerateCRUD(ModuleOptions{
		ProjectDir:    dir,
		ModuleName:    "products",
		Clients:       []string{"admin", "h5"},
		PublicClients: []string{"h5"},
	}); err != nil {
		t.Fatal(err)
	}
	reg := filepath.Join(dir, "internal", "bootstrap", "register_products.go")
	b, err := os.ReadFile(reg)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `engine.Group("/api/v1/admin", middleware.AdminAuth(infra))`) {
		t.Fatal("admin group must use AdminAuth")
	}
	if strings.Contains(s, `engine.Group("/api/v1/h5", middleware.MemberAuth(infra))`) {
		t.Fatal("public h5 must not use MemberAuth")
	}
	if !strings.Contains(s, `engine.Group("/api/v1/h5")`) {
		t.Fatal("public h5 group must have no auth middleware")
	}
}

func TestGenerateCRUDUpdatesRegisterWithCRUD(t *testing.T) {
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

	if err := GenerateCRUD(ModuleOptions{ProjectDir: dir, ModuleName: "orders"}); err != nil {
		t.Fatal(err)
	}
	reg := filepath.Join(dir, "internal", "bootstrap", "register_orders.go")
	b, err := os.ReadFile(reg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "RegisterCRUDRoutes") {
		t.Fatal("crud register should include RegisterCRUDRoutes")
	}
}
