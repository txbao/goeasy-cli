package schema

import "testing"

func TestGoTypeBigint(t *testing.T) {
	f := GoFieldFromColumn(ColumnMeta{Name: "id", DBType: "bigint"})
	if f.GoType != "int64" {
		t.Fatalf("got %s", f.GoType)
	}
}

func TestToPascalColID(t *testing.T) {
	if got := GoFieldFromColumn(ColumnMeta{Name: "id", DBType: "bigint"}).Name; got != "ID" {
		t.Fatalf("got %q want ID", got)
	}
	if got := GoFieldFromColumn(ColumnMeta{Name: "category_id", DBType: "int"}).Name; got != "CategoryID" {
		t.Fatalf("got %q want CategoryID", got)
	}
}
