package migrate

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type projectConfig struct {
	Database databaseYAML `yaml:"database"`
}

type databaseYAML struct {
	Enabled bool   `yaml:"enabled"`
	ORM     string `yaml:"orm"`
	Driver  string `yaml:"driver"`
	DSN     string `yaml:"dsn"`
}

func loadDatabaseConfig(configPath string) (databaseYAML, error) {
	b, err := os.ReadFile(configPath)
	if err != nil {
		return databaseYAML{}, err
	}
	var cfg projectConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return databaseYAML{}, err
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "postgres"
	}
	if cfg.Database.ORM == "" {
		cfg.Database.ORM = "sqlx"
	}
	return cfg.Database, nil
}

func (d databaseYAML) validateForMigrate() error {
	if !d.Enabled {
		return fmt.Errorf("database.enabled is false in config")
	}
	if d.DSN == "" {
		return fmt.Errorf("database.dsn is empty")
	}
	switch d.Driver {
	case "postgres", "mysql":
		return nil
	default:
		return fmt.Errorf("migrate supports database.driver postgres or mysql, got %q", d.Driver)
	}
}
