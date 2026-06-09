package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func loadTableMeta(driver, dsn, schemaName, table string) (schema.TableMeta, error) {
	switch driver {
	case "mysql":
		return schema.LoadMySQLTable(dsn, effectiveMySQLDatabase(schemaName, dsn), table)
	default:
		return schema.LoadPostgresTable(dsn, schemaName, table)
	}
}

func listDatabaseTables(driver, dsn, schemaName string) ([]string, error) {
	switch driver {
	case "mysql":
		return schema.ListMySQLTables(dsn, effectiveMySQLDatabase(schemaName, dsn))
	default:
		return schema.ListPostgresTables(dsn, schemaName)
	}
}

func effectiveMySQLDatabase(schemaName, dsn string) string {
	if schemaName != "" && !strings.EqualFold(schemaName, "public") {
		return schemaName
	}
	return schema.MySQLDatabaseFromDSN(dsn)
}

func resolveModuleName(opts DBOptions, physical, prefix string) string {
	if opts.ModuleName != "" {
		return opts.ModuleName
	}
	return schema.ModuleNameFromPhysical(physical, prefix)
}

func assertTableAllowed(physical string, userExclude []string) error {
	if schema.ShouldExcludeTable(physical, userExclude) {
		if schema.ShouldExcludeTable(physical, nil) {
			return fmt.Errorf("table %q is reserved and cannot be generated", physical)
		}
		return fmt.Errorf("table %q is excluded", physical)
	}
	return nil
}

func writeProjectFileOrSkip(projectDir, rel, content string, force bool) (skipped bool, err error) {
	path := filepath.Join(projectDir, rel)
	if !force {
		if _, err := os.Stat(path); err == nil {
			return true, nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, err
	}
	return false, os.WriteFile(path, []byte(content), 0644)
}
