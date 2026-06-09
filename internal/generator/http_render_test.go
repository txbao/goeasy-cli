package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderHTTPClientLayerAdmin(t *testing.T) {
	dir := t.TempDir()
	projectModule := "example.com/demo"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+projectModule+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	clients, err := NormalizeClients([]string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	data := toMap(BuildModuleData("foomod", projectModule))
	layout := ResolveHTTPRoute("foomod", "", "", nil)
	enrichHTTPRouteData(data, layout)
	enrichHTTPClientsDataWithLayout(data, clients, layout)

	if err := renderHTTPClientLayer("module", dir, "foomod", projectModule, clients, data, true); err != nil {
		t.Fatal(err)
	}
	meta := metaFromData(data)
	if err := writeHTTPHandlersFromCodegen(ModuleOptions{ProjectDir: dir, ModuleName: "foomod"}, projectModule, clients, meta, AppStyleService, false); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"handler.go", "router.go", "dto.go"} {
		p := filepath.Join(dir, "internal", "interface", "http", "admin", "foomod", "foomod", name)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected %s: %v", filepath.ToSlash(p), err)
		}
	}
}
