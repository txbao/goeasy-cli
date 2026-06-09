package generator

import (
	"path/filepath"
	"testing"
)

func TestNormalizeClientsDefault(t *testing.T) {
	got, err := NormalizeClients(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "admin" {
		t.Fatalf("got %+v", got)
	}
	if !got[0].UseAuth {
		t.Fatal("admin should use auth by default")
	}
}

func TestResolveHTTPClientsPublicH5(t *testing.T) {
	got, err := ResolveHTTPClients([]string{"admin", "h5"}, []string{"h5"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d clients", len(got))
	}
	if !got[0].UseAuth || got[0].Name != "admin" {
		t.Fatalf("admin: %+v", got[0])
	}
	if got[1].UseAuth || got[1].Name != "h5" {
		t.Fatalf("h5 should be public: %+v", got[1])
	}
}

func TestResolveHTTPClientsRejectsPublicAdmin(t *testing.T) {
	_, err := ResolveHTTPClients([]string{"admin"}, []string{"admin"})
	if err == nil {
		t.Fatal("expected error for --public admin")
	}
}

func TestResolveHTTPClientsPublicRequiresClient(t *testing.T) {
	_, err := ResolveHTTPClients([]string{"admin"}, []string{"h5"})
	if err == nil {
		t.Fatal("expected error when --public h5 without --client h5")
	}
}

func TestHTTPModuleRel(t *testing.T) {
	meta := metaForTest("sys_roles", "sys_roles", "sys_roles")
	rel := httpModuleRel("admin", meta, "handler.go")
	want := "internal/interface/http/admin/sys_roles/sys_roles/handler.go"
	if filepath.ToSlash(rel) != want {
		t.Fatalf("got %s", rel)
	}
}

func TestHTTPModuleRelGrouped(t *testing.T) {
	layout := ResolveHTTPRoute("sys_roles", "", "", map[string]string{"sys_": "system"})
	rel := layout.HTTPRel("admin", "handler.go")
	want := "internal/interface/http/admin/system/roles/handler.go"
	if filepath.ToSlash(rel) != want {
		t.Fatalf("got %s want %s", rel, want)
	}
}
