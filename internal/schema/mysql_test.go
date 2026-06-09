package schema

import "testing"

func TestMySQLDatabaseFromDSN(t *testing.T) {
	dsn := "root:secret@tcp(127.0.0.1:3306)/app_db?parseTime=true&loc=Local"
	if got := MySQLDatabaseFromDSN(dsn); got != "app_db" {
		t.Fatalf("got %q", got)
	}
	if MySQLDatabaseFromDSN("invalid") != "" {
		t.Fatal("expected empty")
	}
}
