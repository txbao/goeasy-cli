package schema

import "testing"

func TestEntityJSONTag(t *testing.T) {
	nullable := ColumnMeta{Name: "deleted_at", Nullable: true}
	if got := EntityJSONTag(nullable, "*time.Time"); got != `json:"deleted_at,omitempty"` {
		t.Fatalf("nullable: got %q", got)
	}
	required := ColumnMeta{Name: "name", Nullable: false}
	if got := EntityJSONTag(required, "string"); got != `json:"name"` {
		t.Fatalf("required: got %q", got)
	}
}
