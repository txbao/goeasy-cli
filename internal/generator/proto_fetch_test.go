package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchProtoFromFileURL(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/org/order\n"), 0644); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "remote.proto")
	if err := os.WriteFile(src, []byte(`syntax = "proto3";
package sys_roles;
option go_package = "github.com/org/user/api/proto/sys_roles;sys_rolespb";
`), 0644); err != nil {
		t.Fatal(err)
	}
	rel, err := FetchProtoFromURL(dir, src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rel, "api/proto/imported/") {
		t.Fatalf("unexpected rel: %s", rel)
	}
	out, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `github.com/org/order/api/proto/gen/imported/remote`) {
		t.Fatalf("go_package not rewritten: %s", out)
	}
}
