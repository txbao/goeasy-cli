package generator

import (
	"path/filepath"
	"testing"
)

func TestResolveHTTPRouteFromPrefix(t *testing.T) {
	layout := ResolveHTTPRoute("sys_roles", "", "", map[string]string{"sys_": "system"})
	if !layout.Grouped || layout.Domain != "system" || layout.Resource != "roles" {
		t.Fatalf("unexpected layout: %+v", layout)
	}
	if layout.RoutePrefix() != "/system/roles" {
		t.Fatalf("route prefix: %s", layout.RoutePrefix())
	}
	if filepath.ToSlash(layout.HTTPRel("admin", "handler.go")) != "internal/interface/http/admin/system/roles/handler.go" {
		t.Fatalf("http rel: %s", layout.HTTPRel("admin", "handler.go"))
	}
}

func TestResolveHTTPRouteExplicit(t *testing.T) {
	layout := ResolveHTTPRoute("sys_roles", "order", "items", nil)
	if layout.Domain != "order" || layout.Resource != "items" {
		t.Fatalf("unexpected: %+v", layout)
	}
}

func TestResolveHTTPRouteDefaultDomain(t *testing.T) {
	layout := ResolveHTTPRoute("health", "", "", nil)
	if layout.Domain != "health" || layout.Resource != "health" {
		t.Fatalf("expected default domain layout: %+v", layout)
	}
	if layout.RoutePrefix() != "/health/health" {
		t.Fatalf("route: %s", layout.RoutePrefix())
	}
}
