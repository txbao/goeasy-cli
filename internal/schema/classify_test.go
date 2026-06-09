package schema

import "testing"

func TestShouldExcludeMigrationsTable(t *testing.T) {
	if !ShouldExcludeTable("_sqlx_migrations", nil) {
		t.Fatal("expected exclude _sqlx_migrations")
	}
	if ShouldExcludeTable("sys_roles", nil) {
		t.Fatal("sys_roles should not be excluded")
	}
}

func TestClassifyCreateUpdateOmit(t *testing.T) {
	tbl := TableMeta{
		Name: "sys_roles",
		Columns: []ColumnMeta{
			{Name: "id", DBType: "bigint", IsPrimaryKey: true, Ordinal: 1},
			{Name: "name", DBType: "varchar", Ordinal: 2},
			{Name: "status", DBType: "smallint", Ordinal: 3},
			{Name: "created_at", DBType: "timestamp with time zone", Ordinal: 4},
			{Name: "updated_at", DBType: "timestamp with time zone", Ordinal: 5},
			{Name: "deleted_at", DBType: "timestamp with time zone", Nullable: true, Ordinal: 6},
		},
	}
	ct := Classify(tbl, "sys_roles", "sys_roles", DefaultCodegenRules())
	if len(ct.CreateCols) != 2 {
		t.Fatalf("CreateCols want 2 (name,status) got %d", len(ct.CreateCols))
	}
	for _, c := range ct.CreateCols {
		if c.Name == "id" || c.Name == "created_at" {
			t.Fatalf("unexpected create col %s", c.Name)
		}
	}
	for _, c := range ct.UpdateCols {
		if c.Name == "created_at" || c.Name == "updated_at" || c.Name == "deleted_at" {
			t.Fatalf("unexpected update col %s", c.Name)
		}
	}
}

func TestModuleNameFromPrefix(t *testing.T) {
	if ModuleNameFromPhysical("ge_sys_roles", "ge_") != "sys_roles" {
		t.Fatal("prefix strip")
	}
}
