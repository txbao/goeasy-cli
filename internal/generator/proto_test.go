package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateProtoWritesModuleIDPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := GenerateProto(ModuleOptions{
		ProjectDir: dir,
		ModuleName: "sys_roles",
	}); err != nil {
		t.Fatal(err)
	}
	protoPath := filepath.Join(dir, "api", "proto", "sys_roles.proto")
	if _, err := os.Stat(protoPath); err != nil {
		t.Fatalf("expected sys_roles.proto, not MODULE.proto: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "api", "proto", "MODULE.proto")); err == nil {
		t.Fatal("must not create api/proto/MODULE.proto")
	}
	body, err := os.ReadFile(protoPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, "package sys_roles;") || !strings.Contains(s, "service SysRolesService") {
		t.Fatalf("unexpected proto content:\n%s", s)
	}
}

func TestGenerateProtoDomainLayoutStillUsesModuleIDInProto(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := `codegen:
  domains:
    system:
      table_prefix: sys_
      modules:
        sys_roles:
          resource: roles
`
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	if err := GenerateProto(ModuleOptions{
		ProjectDir: dir,
		ModuleName: "sys_roles",
		ConfigPath: filepath.Join(dir, "configs", "config.yaml"),
	}); err != nil {
		t.Fatal(err)
	}
	protoPath := filepath.Join(dir, "api", "proto", "sys_roles.proto")
	if _, err := os.Stat(protoPath); err != nil {
		t.Fatal(err)
	}
}
