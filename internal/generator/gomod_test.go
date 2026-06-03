package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitGoModule(t *testing.T) {
	dir := t.TempDir()
	const mod = "github.com/example/demo"
	if err := initGoModule(dir, mod, DefaultGoEasyModule, ""); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if !strings.HasPrefix(content, "module "+mod) {
		t.Fatalf("go.mod module line: got %q", content)
	}
}

func TestInitGoModuleWithReplace(t *testing.T) {
	dir := t.TempDir()
	replacePath := filepath.Join(dir, "local-goeasy")
	if err := os.MkdirAll(replacePath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := initGoModule(dir, "github.com/example/demo", DefaultGoEasyModule, replacePath); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	want := "replace " + DefaultGoEasyModule + " => " + replacePath
	if !strings.Contains(content, want) {
		t.Fatalf("go.mod missing replace: %q", content)
	}
}

func TestResolveGoEasyModule(t *testing.T) {
	t.Setenv("GOEASY_MODULE", "")
	got := resolveGoEasyModule(Options{})
	if got != DefaultGoEasyModule {
		t.Fatalf("default: got %q", got)
	}
	got = resolveGoEasyModule(Options{GoEasyModule: "github.com/custom/goeasy"})
	if got != "github.com/custom/goeasy" {
		t.Fatalf("flag: got %q", got)
	}
	t.Setenv("GOEASY_MODULE", "github.com/env/goeasy")
	got = resolveGoEasyModule(Options{})
	if got != "github.com/env/goeasy" {
		t.Fatalf("env: got %q", got)
	}
}
