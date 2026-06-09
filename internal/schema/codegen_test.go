package schema

import "testing"

func TestUniqueColumnsByName(t *testing.T) {
	create := []ColumnMeta{{Name: "name"}, {Name: "code"}}
	update := []ColumnMeta{{Name: "name"}, {Name: "status"}}
	got := UniqueColumnsByName(create, update)
	if len(got) != 3 {
		t.Fatalf("got %d cols", len(got))
	}
	if got[0].Name != "name" || got[1].Name != "code" || got[2].Name != "status" {
		t.Fatalf("order wrong: %+v", got)
	}
}

func TestColsNeedTimeImport(t *testing.T) {
	cols := []ColumnMeta{{Name: "created_at", DBType: "timestamptz", Nullable: false}}
	if !ColsNeedTimeImport(cols) {
		t.Fatal("expected time import")
	}
	if ColsNeedTimeImport([]ColumnMeta{{Name: "name", DBType: "varchar"}}) {
		t.Fatal("unexpected time import")
	}
}
