package generator

import "testing"

func TestTableName(t *testing.T) {
	if tableName("", "sys_roles") != "sys_roles" {
		t.Fatal("empty prefix")
	}
	if tableName("ge_", "sys_roles") != "ge_sys_roles" {
		t.Fatal("with prefix")
	}
}
