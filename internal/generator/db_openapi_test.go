package generator

import (
	"strings"
	"testing"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func TestGenOpenAPIUsesTableComment(t *testing.T) {
	ct := schema.Classify(schema.TableMeta{
		Name:         "sys_roles",
		TableComment: "系统角色表",
		Columns: []schema.ColumnMeta{
			{Name: "id", DBType: "bigint", IsPrimaryKey: true, Ordinal: 1, Comment: "主键"},
			{Name: "name", DBType: "varchar", Ordinal: 2, Comment: "角色名"},
		},
	}, "sys_roles", "sys_roles", schema.DefaultCodegenRules())
	layout := ResolveHTTPRoute("sys_roles", "", "", map[string]string{"sys_": "system"})
	out := genOpenAPIFile("github.com/demo/app", ct, "SysRoles", "sys_roles", "admin", layout)
	if !strings.Contains(out, "/api/v1/admin/system/roles") {
		t.Fatal("expected admin API path prefix")
	}
	if !strings.Contains(out, "系统角色表") {
		t.Fatal("expected table comment in description")
	}
	if !strings.Contains(out, "角色名") {
		t.Fatal("expected column comment in schema")
	}
	if strings.Contains(out, "requestBody") && strings.Contains(out, "operationId: list") {
		// list GET 不应带 requestBody
		idx := strings.Index(out, "operationId: list")
		chunk := out[idx : idx+400]
		if strings.Contains(chunk, "requestBody") {
			t.Fatal("list GET should not have requestBody")
		}
	}
}

func TestGenProtoIncludesListRPC(t *testing.T) {
	ct := rolesLikeClassified()
	out := genProtoFile("github.com/demo/app", ct, "SysRoles", "sys_roles")
	if !strings.Contains(out, "rpc ListSysRoles") {
		t.Fatal("expected List RPC")
	}
	if !strings.Contains(out, "message ListSysRolesResponse") {
		t.Fatal("expected ListSysRolesResponse")
	}
}
