package schema

import (
	"strings"
	"testing"
)

func TestFieldTagsCreateRequired(t *testing.T) {
	c := ColumnMeta{Name: "name", DBType: "varchar(64)", Nullable: false}
	got := FieldTags(c, "create")
	if !strings.Contains(got, `json:"name"`) || !strings.Contains(got, "required") || !strings.Contains(got, "max=64") {
		t.Fatalf("got %q", got)
	}
}

func TestFieldTagsUpdateOmitempty(t *testing.T) {
	c := ColumnMeta{Name: "name", DBType: "varchar(64)", Nullable: false}
	got := FieldTags(c, "update")
	if !strings.Contains(got, "omitempty") || strings.Contains(got, "required") {
		t.Fatalf("got %q", got)
	}
}
