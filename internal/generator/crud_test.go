package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCRUDIncludesRepositoryPG(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	cfg := `database:
  driver: postgres
`
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}

	err := GenerateCRUD(ModuleOptions{
		ProjectDir: dir,
		ModuleName: "sys_roles",
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := metaForTest("sys_roles", "sys_roles", "sys_roles")
	pg := filepath.Join(dir, persistenceRepoRel(meta, "repository_pg.go"))
	if _, err := os.Stat(pg); err != nil {
		t.Fatalf("expected repository_pg.go: %v", err)
	}
}

func TestGenerateCRUDRepositoryPGDomainLayout(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
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
	opts := ModuleOptions{
		ProjectDir: dir,
		ModuleName: "sys_roles",
		ConfigPath: filepath.Join(dir, "configs", "config.yaml"),
	}
	if err := GenerateCRUD(opts); err != nil {
		t.Fatal(err)
	}
	meta := metaForTest("sys_roles", "system", "roles")
	pgPath := filepath.Join(dir, persistenceRepoRel(meta, "repository_pg.go"))
	body, err := os.ReadFile(pgPath)
	if err != nil {
		t.Fatalf("repository_pg.go: %v", err)
	}
	s := string(body)
	if strings.Contains(s, "<no value>") {
		t.Fatalf("repository_pg must not contain unresolved template values:\n%s", s)
	}
	wantImport := `domain "example.com/demo/internal/domain/system/roles"`
	if !strings.Contains(s, wantImport) {
		t.Fatalf("expected import %q in repository_pg.go", wantImport)
	}
	if !strings.Contains(s, "package roles") {
		t.Fatal("repository_pg package must be resource name roles")
	}
	if !strings.Contains(s, "func NewPGRepository(sqlxDB *sqlx.DB, driver, table string, c zcache.Cache") {
		t.Fatal("repository_pg must use cache-aware NewPGRepository signature")
	}
}

func TestGenerateCRUDWithMigration(t *testing.T) {
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

	err := GenerateCRUD(ModuleOptions{
		ProjectDir:    dir,
		ModuleName:    "orders",
		WithMigration: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	migDir := filepath.Join(dir, "migrations", "postgres")
	entries, err := os.ReadDir(migDir)
	if err != nil {
		t.Fatal(err)
	}
	var foundUp bool
	for _, e := range entries {
		name := e.Name()
		if strings.Contains(name, "create_orders_table") && strings.HasSuffix(name, ".up.sql") {
			foundUp = true
		}
	}
	if !foundUp {
		t.Fatal("expected create_orders_table up migration in migrations/postgres")
	}
}

func TestReadProjectDBDriverDefault(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte("database: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	driver, err := readProjectDBDriver(dir)
	if err != nil || driver != "postgres" {
		t.Fatalf("driver=%q err=%v", driver, err)
	}
}
