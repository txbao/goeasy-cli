package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateModuleOmitsListQuery(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := GenerateModule(ModuleOptions{ProjectDir: dir, ModuleName: "foomod", Force: true, AppStyle: "light_cqrs"}); err != nil {
		t.Fatal(err)
	}
	meta := metaForTest("foomod", "foomod", "foomod")
	listPath := filepath.Join(dir, relQueryList(meta))
	if _, err := os.Stat(listPath); err == nil {
		t.Fatal("add module must not generate query/list.go")
	}
	getPath := filepath.Join(dir, meta.appRel("query/get.go"))
	if _, err := os.Stat(getPath); err != nil {
		t.Fatalf("expected query/get.go: %v", err)
	}
}

func TestGenerateModuleServiceStyle(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := GenerateModule(ModuleOptions{ProjectDir: dir, ModuleName: "foomod", Force: true}); err != nil {
		t.Fatal(err)
	}
	meta := metaForTest("foomod", "foomod", "foomod")
	for _, rel := range []string{
		meta.appRel("command/create.go"),
		meta.appRel("query/get.go"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Fatalf("service style must not generate %s", rel)
		}
	}
	appPath := filepath.Join(dir, meta.appRel("application.go"))
	b, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "func (a *Application) Get(") {
		t.Fatal("service application must expose Get")
	}
	if !strings.Contains(string(b), "func (a *Application) Create(ctx context.Context, cmd CreateCommand) (string, error)") {
		t.Fatal("service application Create must return (string, error)")
	}
	if !strings.Contains(string(b), "func (a *Application) Update(") {
		t.Fatal("service application must expose Update")
	}
	if !strings.Contains(string(b), "func (a *Application) Delete(") {
		t.Fatal("service application must expose Delete")
	}
	as := string(b)
	if strings.Contains(as, "func (a *Application) List(") || strings.Contains(as, "ListResult") {
		t.Fatal("add module must not generate List or ListResult in application")
	}
	handler, err := os.ReadFile(filepath.Join(dir, meta.HTTPRel("admin", "handler.go")))
	if err != nil {
		t.Fatal(err)
	}
	hs := string(handler)
	if strings.Contains(hs, "Queries()") || strings.Contains(hs, "Commands()") {
		t.Fatal("service handler must not use CQRS facade")
	}
	if !strings.Contains(hs, "h.app.Get(") {
		t.Fatal("service handler must call h.app.Get")
	}
}

func TestGenerateCRUDServiceHTTPWithoutForce(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := `codegen:
  layout: domain
  app_style: service
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
	handlerPath := filepath.Join(dir, meta.HTTPRel("admin", "handler.go"))
	crudPath := filepath.Join(dir, meta.HTTPRel("admin", "handler_crud.go"))
	handler, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("handler.go: %v", err)
	}
	crud, err := os.ReadFile(crudPath)
	if err != nil {
		t.Fatalf("handler_crud.go: %v", err)
	}
	hs, cs := string(handler), string(crud)
	if strings.Contains(hs, "Queries()") || strings.Contains(cs, "Commands()") {
		t.Fatal("default service add crud must not emit light_cqrs HTTP handlers")
	}
	if !strings.Contains(hs, "h.app.Get(") || !strings.Contains(hs, "h.app.Create(") {
		t.Fatal("handler.go must use service Application methods")
	}
	if !strings.Contains(cs, "h.app.Update(") || !strings.Contains(cs, "h.app.Delete(") {
		t.Fatal("handler_crud.go must use service Application methods")
	}
	app, err := os.ReadFile(filepath.Join(dir, meta.appRel("application.go")))
	if err != nil {
		t.Fatal(err)
	}
	as := string(app)
	if !strings.Contains(as, "func (a *Application) Create(ctx context.Context, cmd CreateCommand) (string, error)") {
		t.Fatal("application.go Create must return (string, error)")
	}
	if !strings.Contains(as, "type UpdateCommand struct") || !strings.Contains(as, "func (a *Application) Update(") {
		t.Fatal("application.go must define UpdateCommand and Update")
	}
}

func TestGenerateCRUDLightCQRSApplication(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := GenerateCRUD(ModuleOptions{
		ProjectDir: dir,
		ModuleName: "orders",
		Force:      true,
		AppStyle:   "light_cqrs",
	}); err != nil {
		t.Fatal(err)
	}
	meta := metaForTest("orders", "orders", "orders")
	handler, err := os.ReadFile(filepath.Join(dir, meta.HTTPRel("admin", "handler.go")))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(handler), "Queries().Get") {
		t.Fatal("light_cqrs handler must use Queries().Get")
	}
	create, err := os.ReadFile(filepath.Join(dir, meta.appRel("command/create.go")))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(create), "func (h *Handler) Create(ctx context.Context, cmd CreateCommand) (string, error)") {
		t.Fatal("light_cqrs command Create must return (string, error)")
	}
}

func TestEnsureModuleCRUDLayerAddsList(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ensureModuleCRUDLayer(ModuleOptions{ProjectDir: dir, ModuleName: "orders", Force: true, AppStyle: "light_cqrs"}); err != nil {
		t.Fatal(err)
	}
	meta := metaForTest("orders", "orders", "orders")
	for _, rel := range []string{
		relQueryList(meta),
		relAppList(meta),
		meta.domainRel("repository.go"),
		persistenceRepoRel(meta, "repository.go"),
		meta.appRel("application.go"),
		meta.appRel("command/create.go"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	domainRepo, err := os.ReadFile(filepath.Join(dir, meta.domainRel("repository.go")))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(domainRepo), "List(ctx context.Context") || !strings.Contains(string(domainRepo), "Save(ctx context.Context") {
		t.Fatal("domain repository must include List and Save")
	}
}

func TestEnsureModuleCRUDLayerServiceStyle(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ensureModuleCRUDLayer(ModuleOptions{ProjectDir: dir, ModuleName: "orders", Force: true}); err != nil {
		t.Fatal(err)
	}
	meta := metaForTest("orders", "orders", "orders")
	if _, err := os.Stat(filepath.Join(dir, relQueryList(meta))); err == nil {
		t.Fatal("service style must not generate query/list.go")
	}
	b, err := os.ReadFile(filepath.Join(dir, meta.appRel("application.go")))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "func (a *Application) List(") {
		t.Fatal("service application must include List")
	}
	if !strings.Contains(string(b), "func (a *Application) Update(") || !strings.Contains(string(b), "func (a *Application) Delete(") {
		t.Fatal("service application must include Update and Delete")
	}
}
