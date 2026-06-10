package migrate

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationsTable = "_sqlx_migrations"

type Options struct {
	ProjectDir    string
	ConfigPath    string
	MigrationsDir string
	Steps         int
}

func Up(opts Options) error {
	m, err := openMigrate(opts)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("no pending migrations")
			return nil
		}
		return err
	}
	return nil
}

func Down(opts Options) error {
	steps := opts.Steps
	if steps <= 0 {
		steps = 1
	}
	m, err := openMigrate(opts)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Steps(-steps); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("no migrations to roll back")
			return nil
		}
		return err
	}
	return nil
}

func Status(opts Options) error {
	resolved, _, err := ResolveOptions(opts)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "info: migrations dir %s\n", resolved.MigrationsDir)

	m, err := openMigrate(opts)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			version = 0
			dirty = false
		} else {
			return err
		}
	}

	files, err := listUpMigrations(resolved.MigrationsDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		id, err := migrationID(f)
		if err != nil {
			continue
		}
		state := "pending"
		if uint(id) <= version {
			state = "applied"
		}
		fmt.Printf("%s\t%s\n", migrationVersion(f), state)
	}
	fmt.Printf("current version: %d, dirty: %t\n", version, dirty)
	return nil
}

func Version(opts Options) error {
	m, err := openMigrate(opts)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			version = 0
			dirty = false
		} else {
			return err
		}
	}
	fmt.Printf("current version: %d, dirty: %t\n", version, dirty)
	return nil
}

func Goto(opts Options, target uint) error {
	m, err := openMigrate(opts)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	return m.Migrate(target)
}

func Force(opts Options, target int) error {
	m, err := openMigrate(opts)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	return m.Force(target)
}

func Create(opts Options, name string) error {
	if name == "" {
		return fmt.Errorf("migration name is required")
	}
	resolved, dbCfg, err := ResolveOptions(opts)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "info: create in migrations/%s/\n", dbCfg.Driver)
	return createMigrationFiles(resolved.MigrationsDir, name)
}

func createMigrationFiles(migrationsDir, name string) error {
	name = strings.ReplaceAll(name, " ", "_")
	ts := time.Now().Format("20060102150405")
	version := ts + "_" + name
	up := filepath.Join(migrationsDir, version+".up.sql")
	down := filepath.Join(migrationsDir, version+".down.sql")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return err
	}
	upBody := fmt.Sprintf("-- %s\n\nSELECT 1;\n", version)
	if err := os.WriteFile(up, []byte(upBody), 0644); err != nil {
		return err
	}
	downBody := fmt.Sprintf("-- rollback %s\n\nSELECT 1;\n", version)
	if err := os.WriteFile(down, []byte(downBody), 0644); err != nil {
		return err
	}
	fmt.Printf("created %s\n", up)
	fmt.Printf("created %s\n", down)
	return nil
}

func openMigrate(opts Options) (*migrate.Migrate, error) {
	resolved, dbCfg, err := ResolveOptions(opts)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "info: driver=%s migrations=%s\n", dbCfg.Driver, resolved.MigrationsDir)

	db, driverName, dbDriver, err := openSQL(dbCfg)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping: %w", err)
	}

	absDir, err := filepath.Abs(resolved.MigrationsDir)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := os.Stat(absDir); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrations directory %s: %w", absDir, err)
	}

	sourceURL := fileURL(absDir)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, driverName, dbDriver)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return m, nil
}

func openSQL(cfg databaseYAML) (*sql.DB, string, database.Driver, error) {
	switch cfg.Driver {
	case "postgres":
		db, err := sql.Open("pgx", cfg.DSN)
		if err != nil {
			return nil, "", nil, err
		}
		drv, err := postgres.WithInstance(db, &postgres.Config{MigrationsTable: migrationsTable})
		if err != nil {
			_ = db.Close()
			return nil, "", nil, err
		}
		return db, "postgres", drv, nil
	case "mysql":
		db, err := sql.Open("mysql", cfg.DSN)
		if err != nil {
			return nil, "", nil, err
		}
		drv, err := mysql.WithInstance(db, &mysql.Config{MigrationsTable: migrationsTable})
		if err != nil {
			_ = db.Close()
			return nil, "", nil, err
		}
		return db, "mysql", drv, nil
	default:
		return nil, "", nil, fmt.Errorf("unsupported driver %q", cfg.Driver)
	}
}

func closeMigrate(m *migrate.Migrate) {
	if m == nil {
		return
	}
	_, _ = m.Close()
}

func fileURL(absPath string) string {
	path := filepath.ToSlash(absPath)
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(path, "/") {
			path = strings.TrimPrefix(path, "/")
		}
		return "file://" + path
	}
	if strings.HasPrefix(path, "/") {
		return "file://" + path
	}
	return "file:///" + path
}

func listUpMigrations(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)
	return files, nil
}

func migrationVersion(upPath string) string {
	base := filepath.Base(upPath)
	return strings.TrimSuffix(base, ".up.sql")
}

func migrationID(upPath string) (uint64, error) {
	base := filepath.Base(upPath)
	name := strings.TrimSuffix(base, ".up.sql")
	parts := strings.SplitN(name, "_", 2)
	return strconv.ParseUint(parts[0], 10, 64)
}
