package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProtoFilesSingle(t *testing.T) {
	dir := t.TempDir()
	protoDir := filepath.Join(dir, "api", "proto")
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(protoDir, "sys_roles.proto")
	if err := os.WriteFile(f, []byte("syntax = \"proto3\";"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := resolveProtoFiles(dir, "api/proto/sys_roles.proto")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "api/proto/sys_roles.proto" {
		t.Fatalf("got %v", got)
	}
}

func TestResolveProtoFilesAll(t *testing.T) {
	dir := t.TempDir()
	protoDir := filepath.Join(dir, "api", "proto")
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.proto", "b.proto"} {
		if err := os.WriteFile(filepath.Join(protoDir, name), []byte("syntax = \"proto3\";"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := resolveProtoFiles(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}
