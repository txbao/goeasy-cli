package generator

import (
	"strings"
	"testing"
)

func TestGenGRPCConvertGo(t *testing.T) {
	cols := []GRPCCol{{Pascal: "ID", ProtoField: "Id", GoType: "int64", ProtoType: "int64"}}
	meta := metaForTest("sys_roles", "system", "roles")
	out := genGRPCConvertGo(meta, "github.com/demo/app", "sysroles", "SysRoles", "github.com/demo/app/api/proto/gen/sys_roles", cols, nil, nil)
	if !strings.Contains(out, "dtoFieldID") {
		t.Fatal("expected dtoFieldID")
	}
	if !strings.Contains(out, "package roles") {
		t.Fatal("expected package")
	}
}

func TestProtoStructField(t *testing.T) {
	if protoStructField("ID") != "Id" {
		t.Fatal("expected Id for proto struct field")
	}
}
