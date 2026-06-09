package grpccmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDirect(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	cfg := `app_name: order
discovery:
  mode: direct
  services:
    user: "10.10.10.12:28021"
`
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Resolve(ResolveOptions{CommonOptions: CommonOptions{
		Dir: dir, ConfigPath: "configs/config.yaml", Service: "user",
	}}); err != nil {
		t.Fatal(err)
	}
}
