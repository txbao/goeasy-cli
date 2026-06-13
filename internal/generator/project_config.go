package generator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type projectDBConfig struct {
	Database struct {
		Driver      string `yaml:"driver"`
		TablePrefix string `yaml:"table_prefix"`
	} `yaml:"database"`
}

func readProjectDatabaseConfig(projectDir, configPath string) (driver, tablePrefix string, err error) {
	path := resolveConfigPath(projectDir, configPath)
	b, err := os.ReadFile(path)
	if err != nil {
		return "postgres", "", nil
	}
	var cfg projectDBConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return "", "", err
	}
	driver = cfg.Database.Driver
	if driver == "" {
		driver = "postgres"
	}
	switch driver {
	case "postgres", "mysql":
		return driver, cfg.Database.TablePrefix, nil
	default:
		return "", "", fmt.Errorf("unsupported database.driver %q in config (use postgres or mysql)", driver)
	}
}

func readProjectDBDriver(projectDir string) (string, error) {
	driver, _, err := readProjectDatabaseConfig(projectDir, "")
	return driver, err
}

// tableName 与 shared/dbx.TableName 规则一致。
func tableName(prefix, module string) string {
	if prefix == "" {
		return module
	}
	return prefix + module
}
