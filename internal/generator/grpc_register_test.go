package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGRPCBootstrapRegistryAppendsCall(t *testing.T) {
	dir := t.TempDir()
	bootstrapDir := filepath.Join(dir, "internal", "bootstrap")
	if err := os.MkdirAll(bootstrapDir, 0755); err != nil {
		t.Fatal(err)
	}
	initial := `package bootstrap

func RegisterGRPCServers(s interface{}, infra interface{}) {
	// grpc bootstrap modules (goeasy add db proto appends below)
}
`
	if err := os.WriteFile(filepath.Join(bootstrapDir, "grpc.go"), []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}
	opts := DBOptions{ModuleOptions: ModuleOptions{ProjectDir: dir}}
	if err := ensureGRPCBootstrapRegistry(opts, "SysRoles"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(bootstrapDir, "grpc.go"))
	if err != nil {
		t.Fatal(err)
	}
	out := string(b)
	if !strings.Contains(out, "RegisterSysRolesGRPC(s, infra)") {
		t.Fatalf("missing bootstrap call: %s", out)
	}
}
