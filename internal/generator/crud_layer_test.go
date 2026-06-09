package generator

import (
	"strings"
	"testing"
)

func TestGenModuleDomainRepositoryIncludesList(t *testing.T) {
	out := genModuleDomainRepository("foomod")
	if !strings.Contains(out, "List(ctx context.Context") {
		t.Fatal("expected List on domain repository")
	}
}
