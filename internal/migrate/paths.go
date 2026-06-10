package migrate

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResolveOptions loads config and resolves migrations dir to migrations/<driver>/ by default.
func ResolveOptions(opts Options) (Options, databaseYAML, error) {
	cfgPath := opts.ConfigPath
	if !filepath.IsAbs(cfgPath) {
		cfgPath = filepath.Join(opts.ProjectDir, cfgPath)
	}
	dbCfg, err := loadDatabaseConfig(cfgPath)
	if err != nil {
		return opts, databaseYAML{}, err
	}
	if err := dbCfg.validateForMigrate(); err != nil {
		return opts, databaseYAML{}, err
	}
	dir, err := resolveMigrationsDir(opts.ProjectDir, opts.MigrationsDir, dbCfg.Driver)
	if err != nil {
		return opts, databaseYAML{}, err
	}
	opts.MigrationsDir = dir
	return opts, dbCfg, nil
}

func resolveMigrationsDir(projectDir, migrationsFlag, driver string) (string, error) {
	flag := strings.TrimSpace(migrationsFlag)
	if filepath.IsAbs(flag) {
		return filepath.Clean(flag), nil
	}
	norm := filepath.ToSlash(filepath.Clean(flag))
	if flag != "" && norm != "migrations" && norm != "." {
		abs, err := filepath.Abs(filepath.Join(projectDir, flag))
		if err != nil {
			return "", err
		}
		return abs, nil
	}

	switch driver {
	case "postgres", "mysql":
		return filepath.Abs(filepath.Join(projectDir, "migrations", driver))
	default:
		return "", fmt.Errorf("unsupported database.driver %q (use postgres or mysql)", driver)
	}
}
