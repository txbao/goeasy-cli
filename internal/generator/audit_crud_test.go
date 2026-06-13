package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCRUDWithAuditServiceStyle(t *testing.T) {
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

	meta := metaForTest("orders", "orders", "orders")
	if err := GenerateCRUD(ModuleOptions{
		ProjectDir: dir,
		ModuleName: "orders",
		WithAudit:  true,
		AppStyle:   "service",
	}); err != nil {
		t.Fatal(err)
	}

	appPath := filepath.Join(dir, meta.appRel("application.go"))
	body, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, "recorder audit.Recorder") {
		t.Fatalf("expected recorder field:\n%s", s)
	}
	if !strings.Contains(s, "a.recorder.Record(ctx, contextx.OperatorFrom(ctx), audit.Entry") {
		t.Fatalf("expected audit stub in application:\n%s", s)
	}

	regPath := filepath.Join(dir, "internal", "bootstrap", "register_orders.go")
	reg, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}
	rs := string(reg)
	if !strings.Contains(rs, "infra.AuditRecorder") {
		t.Fatalf("expected AuditRecorder in register:\n%s", rs)
	}
	if !strings.Contains(rs, "InjectOperatorContext") {
		t.Fatalf("expected InjectOperatorContext in register:\n%s", rs)
	}
}

func TestGenerateCRUDWithAuditLightCQRS(t *testing.T) {
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

	meta := metaForTest("products", "products", "products")
	if err := GenerateCRUD(ModuleOptions{
		ProjectDir: dir,
		ModuleName: "products",
		WithAudit:  true,
		AppStyle:   "light_cqrs",
	}); err != nil {
		t.Fatal(err)
	}

	createPath := filepath.Join(dir, meta.appRel("command/create.go"))
	body, err := os.ReadFile(createPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, "h.recorder.Record(ctx, contextx.OperatorFrom(ctx), audit.Entry") {
		t.Fatalf("expected audit stub in command/create:\n%s", s)
	}
}

func TestModuleHasAudit(t *testing.T) {
	dir := t.TempDir()
	meta := metaForTest("foo", "foo", "foo")
	appPath := filepath.Join(dir, meta.appRel("application.go"))
	if err := os.MkdirAll(filepath.Dir(appPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(appPath, []byte("type Application struct {\n\trecorder audit.Recorder\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if !moduleHasAudit(dir, meta) {
		t.Fatal("expected moduleHasAudit true")
	}
}
