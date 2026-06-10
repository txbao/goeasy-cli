package migrate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMigrationVersion(t *testing.T) {
	got := migrationVersion(`migrations/postgres/000001_init.up.sql`)
	if got != "000001_init" {
		t.Fatalf("got %q", got)
	}
}

func TestMigrationID(t *testing.T) {
	id, err := migrationID(`migrations/postgres/000001_init.up.sql`)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("got %d", id)
	}
}

func TestFileURL(t *testing.T) {
	if runtime.GOOS == "windows" {
		got := fileURL(`C:\dev\go\src\goeasy\goeasy-cli\migrations\postgres`)
		if got != "file://C:/dev/go/src/goeasy/goeasy-cli/migrations/postgres" {
			t.Fatalf("got %q", got)
		}
	} else {
		got := fileURL(`/home/user/goeasy/goeasy-cli/migrations/postgres`)
		if got != "file:///home/user/goeasy/goeasy-cli/migrations/postgres" {
			t.Fatalf("got %q", got)
		}
	}
}

func TestResolveMigrationsDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "migrations", "postgres"), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := resolveMigrationsDir(dir, "migrations", "postgres")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "migrations", "postgres")
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("default: got %q want %q", got, want)
	}

	custom := filepath.Join(dir, "migrations", "mysql")
	if err := os.MkdirAll(custom, 0755); err != nil {
		t.Fatal(err)
	}
	got, err = resolveMigrationsDir(dir, "migrations/mysql", "postgres")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(got) != filepath.Clean(custom) {
		t.Fatalf("override: got %q want %q", got, custom)
	}

	_, err = resolveMigrationsDir(dir, "migrations", "sqlite")
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}

	abs := filepath.Join(dir, "migrations", "postgres")
	got, err = resolveMigrationsDir(dir, abs, "postgres")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(got) != filepath.Clean(abs) {
		t.Fatalf("absolute path: got %q want %q", got, abs)
	}
}

func TestResolveOptionsIdempotent(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "migrations", "postgres"), 0755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, "cfg.yaml")
	content := `database:
  enabled: true
  driver: postgres
  dsn: postgres://localhost/db
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	opts := Options{ProjectDir: dir, ConfigPath: cfgPath, MigrationsDir: "migrations"}
	first, _, err := ResolveOptions(opts)
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := ResolveOptions(first)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(first.MigrationsDir) != filepath.Clean(second.MigrationsDir) {
		t.Fatalf("idempotent resolve: first=%q second=%q", first.MigrationsDir, second.MigrationsDir)
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

func TestLoadDatabaseConfigMySQL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	content := `database:
  enabled: true
  driver: mysql
  dsn: user:pass@tcp(127.0.0.1:3306)/db
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabaseConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if db.Driver != "mysql" {
		t.Fatalf("driver=%q", db.Driver)
	}
	if err := db.validateForMigrate(); err != nil {
		t.Fatal(err)
	}
}
