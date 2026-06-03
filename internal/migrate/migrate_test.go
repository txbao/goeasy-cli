package migrate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMigrationVersion(t *testing.T) {
	got := migrationVersion(`migrations/000001_init.up.sql`)
	if got != "000001_init" {
		t.Fatalf("got %q", got)
	}
}

func TestMigrationID(t *testing.T) {
	id, err := migrationID(`migrations/000001_init.up.sql`)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("got %d", id)
	}
}

func TestFileURL(t *testing.T) {
	if runtime.GOOS == "windows" {
		got := fileURL(`C:\dev\go\src\qianruan\goeasy\goeasy-cli\migrations`)
		if got != "file://C:/dev/go/src/goeasy/goeasy-cli/migrations" {
			t.Fatalf("got %q", got)
		}
	} else {
		got := fileURL(`/home/user/goeasy/goeasy-cli/migrations`)
		if got != "file:///home/user/goeasy/goeasy-cli/migrations" {
			t.Fatalf("got %q", got)
		}
	}
}

func TestLoadDatabaseConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	content := `database:
  enabled: true
  dsn: postgres://localhost/db
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabaseConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if db.ORM != "sqlx" || db.Driver != "postgres" {
		t.Fatalf("defaults: orm=%q driver=%q", db.ORM, db.Driver)
	}
}
