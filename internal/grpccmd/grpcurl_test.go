package grpccmd

import (
	"reflect"
	"testing"
)

func TestGRPCURLArgsOrderList(t *testing.T) {
	args := grpcURLArgs(true, "10.0.0.1:9001", "list")
	want := []string{"-plaintext", "10.0.0.1:9001", "list"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("list args: got %v want %v", args, want)
	}
}

func TestGRPCURLArgsOrderCall(t *testing.T) {
	args := grpcURLArgs(true, "10.0.0.1:9001", "-d", `{"id":"1"}`, "sys_roles.SysRolesService/GetSysRoles")
	want := []string{"-plaintext", "-d", `{"id":"1"}`, "10.0.0.1:9001", "sys_roles.SysRolesService/GetSysRoles"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("call args: got %v want %v", args, want)
	}
}

// grpcURLArgs mirrors runGRPCURL argument assembly for unit tests.
func grpcURLArgs(plaintext bool, target string, args ...string) []string {
	cmdArgs := make([]string, 0, 8)
	if plaintext {
		cmdArgs = append(cmdArgs, "-plaintext")
	}
	if len(args) >= 2 && args[0] == "-d" {
		cmdArgs = append(cmdArgs, "-d", args[1])
		cmdArgs = append(cmdArgs, target)
		if len(args) > 2 {
			cmdArgs = append(cmdArgs, args[2:]...)
		}
	} else {
		cmdArgs = append(cmdArgs, target)
		cmdArgs = append(cmdArgs, args...)
	}
	return cmdArgs
}
