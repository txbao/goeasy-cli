package generator

import (
	"path/filepath"
	"testing"
)

func TestParseOpenAPIContractUser(t *testing.T) {
	path := filepath.Join("..", "..", "..", "user", "api", "openapi", "sys_roles.openapi.yaml")
	ct, err := parseOpenAPIContract(path)
	if err != nil {
		t.Fatal(err)
	}
	if ct.ModuleSnake != "sys_roles" {
		t.Fatalf("module %q", ct.ModuleSnake)
	}
	if ct.ModulePascal != "SysRoles" {
		t.Fatalf("pascal %q", ct.ModulePascal)
	}
	if !ct.ClientOps["h5"].List || !ct.ClientOps["h5"].Get {
		t.Fatalf("h5 ops: %+v", ct.ClientOps["h5"])
	}
	if len(ct.CT.ReadCols) < 3 {
		t.Fatalf("read cols: %d", len(ct.CT.ReadCols))
	}
}

func TestParseHTTPPathGrouped(t *testing.T) {
	client, layout := parseHTTPPath("/api/v1/admin/system/roles")
	if client != "admin" || !layout.Grouped || layout.Domain != "system" || layout.Resource != "roles" {
		t.Fatalf("got %s %+v", client, layout)
	}
}

func TestParseHTTPPathSingleSegment(t *testing.T) {
	client, layout := parseHTTPPath("/api/v1/admin/sys_roles/{id}")
	if client != "admin" || layout.ModuleID != "sys_roles" || layout.Domain != "sys_roles" || layout.Resource != "sys_roles" {
		t.Fatalf("got %s %+v", client, layout)
	}
}

func TestParseProtoContractUser(t *testing.T) {
	path := filepath.Join("..", "..", "..", "user", "api", "proto", "sys_roles.proto")
	ct, err := parseProtoContract(path)
	if err != nil {
		t.Fatal(err)
	}
	if ct.ModuleSnake != "sys_roles" || ct.ModulePascal != "SysRoles" {
		t.Fatalf("%+v", ct)
	}
	if len(ct.CT.ReadCols) < 5 {
		t.Fatalf("read cols %d", len(ct.CT.ReadCols))
	}
}
