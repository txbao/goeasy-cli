package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

)

func writeModuleTableMigration(opts ModuleOptions, moduleSnake string) error {
	driver, prefix, err := readProjectDatabaseConfig(opts.ProjectDir, opts.ConfigPath)
	if err != nil {
		return err
	}
	fullTable := tableName(prefix, moduleSnake)
	dir := filepath.Join(opts.ProjectDir, "migrations", driver)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	ts := time.Now().Format("20060102150405")
	base := fmt.Sprintf("%s_create_%s_table", ts, moduleSnake)
	upPath := filepath.Join(dir, base+".up.sql")
	downPath := filepath.Join(dir, base+".down.sql")
	if !opts.Force {
		if _, err := os.Stat(upPath); err == nil {
			fmt.Fprintf(os.Stderr, "info: migration %s exists, skipping (use --force)\n", base+".up.sql")
			return nil
		}
	}
	upSQL, downSQL := moduleTableMigrationSQL(driver, fullTable)
	if err := os.WriteFile(upPath, []byte(upSQL), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(downPath, []byte(downSQL), 0644); err != nil {
		return err
	}
	fmt.Printf("  created %s\n", filepath.ToSlash(upPath))
	fmt.Printf("  created %s\n", filepath.ToSlash(downPath))
	return nil
}

func moduleTableMigrationSQL(driver, table string) (up, down string) {
	switch driver {
	case "mysql":
		up = fmt.Sprintf(`-- create table %s
CREATE TABLE IF NOT EXISTS %s (
  id VARCHAR(255) PRIMARY KEY,
  active TINYINT(1) NOT NULL DEFAULT 1
);
`, table, table)
	default:
		up = fmt.Sprintf(`-- create table %s
CREATE TABLE IF NOT EXISTS %s (
  id TEXT PRIMARY KEY,
  active BOOLEAN NOT NULL DEFAULT true
);
`, table, table)
	}
	down = fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", table)
	return up, down
}
