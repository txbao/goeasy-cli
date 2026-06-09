package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModuleExists(t *testing.T) {
	dir := t.TempDir()
	meta := metaForTest("sys_roles", "system", "roles")
	if moduleExists(dir, meta) {
		t.Fatal("expected false for missing module")
	}
	domainDir := filepath.Join(dir, filepath.Dir(meta.domainRel("entity.go")))
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, meta.domainRel("entity.go")), []byte("package roles\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if !moduleExists(dir, meta) {
		t.Fatal("expected true when entity.go exists")
	}
}

func TestRenderScopedSkipTarget(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	meta := metaForTest("sys_roles", "system", "roles")
	domainDir := filepath.Join(dir, filepath.Dir(meta.domainRel("entity.go")))
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, meta.domainRel("entity.go")), []byte("package roles\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, meta.domainRel("repository.go")), []byte("package roles\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := GenerateRepository(ModuleOptions{
		ProjectDir: dir,
		ModuleName: "sys_roles",
		Domain:     "system",
		Resource:   "roles",
	})
	if err != nil {
		t.Fatal(err)
	}
	pg := filepath.Join(dir, persistenceRepoRel(meta, "repository_pg.go"))
	if _, err := os.Stat(pg); err != nil {
		t.Fatalf("expected repository_pg.go: %v", err)
	}
}
