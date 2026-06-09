package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileHasQueryListImportCycle(t *testing.T) {
	old := []byte(`package query
import app "github.com/demo/app/internal/app/sys_roles"
func (h *Handler) List() (app.ListResult, error) { return app.ListResult{}, nil }
`)
	if !fileHasQueryListImportCycle(old, "sys_roles") {
		t.Fatal("expected cycle detection")
	}
	newOK := []byte(`package query
import domain "github.com/demo/app/internal/domain/sys_roles"
func (h *Handler) List() ([]*domain.Aggregate, int64, error) { return nil, 0, nil }
`)
	if fileHasQueryListImportCycle(newOK, "sys_roles") {
		t.Fatal("expected no cycle")
	}
}

func TestWriteDBGeneratedFileFixesQueryListCycle(t *testing.T) {
	dir := t.TempDir()
	snake := "sys_roles"
	meta := metaForTest(snake, snake, snake)
	rel := relQueryList(meta)
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	cycleBody := `package query
import app "github.com/demo/app/internal/app/sys_roles"
func (h *Handler) List(ctx context.Context, page, pageSize int) (app.ListResult, error) {
	return app.ListResult{}, nil
}
`
	if err := os.WriteFile(path, []byte(cycleBody), 0644); err != nil {
		t.Fatal(err)
	}
	fixed := `package query
import domain "github.com/demo/app/internal/domain/sys_roles"
func (h *Handler) List(ctx context.Context, page, pageSize int) ([]*domain.Aggregate, int64, error) {
	return nil, 0, nil
}
`
	skipped, cycleFixed, err := writeDBGeneratedFile(dir, rel, fixed, snake, false)
	if err != nil || skipped || !cycleFixed {
		t.Fatalf("skipped=%v cycleFixed=%v err=%v", skipped, cycleFixed, err)
	}
	data, _ := os.ReadFile(path)
	if fileHasQueryListImportCycle(data, snake) {
		t.Fatal("file still has cycle import")
	}
}

func TestGenAppListNoQueryImport(t *testing.T) {
	out := genAppList(metaForTest("sys_roles", "sys_roles", "sys_roles"), "SysRoles")
	if strings.Contains(out, "import") && strings.Contains(out, "/query") {
		t.Fatal("app list.go must not import query subpackage")
	}
	if !strings.Contains(out, "a.queries.List") {
		t.Fatal("expected queries.List call in same package")
	}
}
