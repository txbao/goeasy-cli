package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateEventDomainLayout(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := GenerateEvent(ModuleOptions{
		ProjectDir: dir,
		ModuleName: "apis-paid",
		Domain:     "iam",
	}); err != nil {
		t.Fatal(err)
	}
	eventPath := filepath.Join(dir, "internal", "domain", "iam", "event", "apis_paid", "event.go")
	if _, err := os.Stat(eventPath); err != nil {
		t.Fatalf("event.go: %v", err)
	}
	b, _ := os.ReadFile(eventPath)
	if !strings.Contains(string(b), "func (e ApisPaid) Name()") {
		t.Fatal("Name() missing on event")
	}
	pubPath := filepath.Join(dir, "internal", "infrastructure", "iam", "event", "apis_paid", "publisher.go")
	if _, err := os.Stat(pubPath); err != nil {
		t.Fatalf("publisher.go: %v", err)
	}
}
