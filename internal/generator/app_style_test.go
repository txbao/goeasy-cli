package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAppStyleAliases(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", "service"},
		{"light", "light_cqrs"},
		{"light_cqrs", "light_cqrs"},
		{"full", "full_cqrs"},
	} {
		got, err := ParseAppStyle(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if string(got) != tc.want {
			t.Fatalf("%q: got %q", tc.in, got)
		}
	}
}

func TestFullCQRSRejected(t *testing.T) {
	_, err := ResolveAppStyle("", "", "full_cqrs", "sys_roles", ModuleMeta{Domain: "system"}, AppStyleService)
	if err == nil || !strings.Contains(err.Error(), "full_cqrs") {
		t.Fatalf("expected rejection, got %v", err)
	}
}

func TestResolveAppStyleFromConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := `codegen:
  app_style: light_cqrs
  domains:
    system:
      app_style: service
      modules:
        sys_roles:
          app_style: light_cqrs
`
	_ = os.MkdirAll(filepath.Join(dir, "configs"), 0755)
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	meta := ModuleMeta{Domain: "system", ModuleID: "sys_roles"}
	style, err := ResolveAppStyle(dir, filepath.Join(dir, "configs", "config.yaml"), "", "sys_roles", meta, AppStyleService)
	if err != nil {
		t.Fatal(err)
	}
	if style != AppStyleLightCQRS {
		t.Fatalf("got %q", style)
	}
}

func TestGenServiceApplication(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genServiceApplication("github.com/demo/app", ct, "SysRoles", "sysRoles", meta)
	if strings.Contains(out, "Queries()") || strings.Contains(out, "Commands()") {
		t.Fatal("service application must not use CQRS handlers")
	}
	if !strings.Contains(out, "func (a *Application) List(") {
		t.Fatal("missing List")
	}
}
