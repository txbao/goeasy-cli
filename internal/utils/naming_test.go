package utils

import "testing"

func TestToPascal(t *testing.T) {
	if got := ToPascal("user_order"); got != "UserOrder" {
		t.Fatalf("got %s", got)
	}
}

func TestToSnake(t *testing.T) {
	if got := ToSnake("UserOrder"); got != "user_order" {
		t.Fatalf("got %s", got)
	}
}

func TestToIdent(t *testing.T) {
	if got := ToIdent("sys_roles"); got != "sysroles" {
		t.Fatalf("got %s", got)
	}
}
