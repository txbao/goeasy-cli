package configpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePriority(t *testing.T) {
	dir := t.TempDir()
	flag := filepath.Join(dir, "custom.yaml")
	if got := Resolve(dir, flag); got != flag {
		t.Fatalf("flag: got %q", got)
	}
	t.Setenv(EnvVar, filepath.Join(dir, "from-env.yaml"))
	if got := Resolve(dir, ""); !filepath.IsAbs(got) || filepath.Base(got) != "from-env.yaml" {
		t.Fatalf("env: got %q", got)
	}
	t.Setenv(EnvVar, "")
	if got := Resolve(dir, ""); got != filepath.Join(dir, DefaultRel) {
		t.Fatalf("default: got %q", got)
	}
}

func TestResolveRelativeFlag(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "configs"), 0755)
	want := filepath.Join(dir, "configs", "config.yaml")
	if got := Resolve(dir, "configs/config.yaml"); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
