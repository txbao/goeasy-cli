package generator

import (
	"strings"
	"testing"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func TestGenCRUDHttpFileContainsRoutes(t *testing.T) {
	meta := ModuleMeta{ModuleID: "sys_roles", Domain: "system", Resource: "roles"}
	cl := ClientSurface{Name: "admin", FullCRUD: true, UseAuth: true}
	ct := schema.ClassifiedTable{
		CreateCols: []schema.ColumnMeta{{Name: "name", DBType: "varchar"}},
	}
	out := genCRUDHttpFile(cl, meta, ct)
	if !strings.Contains(out, "/api/v1/admin/system/roles") {
		t.Fatal("missing route prefix")
	}
	if !strings.Contains(out, "POST {{baseUrl}}/api/v1/admin/system/roles") {
		t.Fatal("missing create request")
	}
	if !strings.Contains(out, "Authorization: Bearer {{token}}") {
		t.Fatal("missing auth header")
	}
}
