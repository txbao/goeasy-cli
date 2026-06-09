package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/schema"

	"gopkg.in/yaml.v3"
)

type projectCodegenYAML struct {
	Database struct {
		Enabled     bool   `yaml:"enabled"`
		DSN         string `yaml:"dsn"`
		Driver      string `yaml:"driver"`
		TablePrefix string `yaml:"table_prefix"`
	} `yaml:"database"`
	Redis struct {
		Enabled bool   `yaml:"enabled"`
		Addr    string `yaml:"addr"`
	} `yaml:"redis"`
	Codegen struct {
		CreateOmit             []string `yaml:"create_omit"`
		UpdateOmit             []string `yaml:"update_omit"`
		SoftDeleteColumn       string   `yaml:"soft_delete_column"`
		TouchCreatedAtOnInsert bool     `yaml:"touch_created_at_on_insert"`
		TouchUpdatedAtOnUpdate bool     `yaml:"touch_updated_at_on_update"`
	} `yaml:"codegen"`
}

func resolveConfigPath(projectDir, configPath string) string {
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(projectDir, configPath)
	}
	return configPath
}

func loadProjectCodegenYAML(projectDir, configPath string) (projectCodegenYAML, string, error) {
	configPath = resolveConfigPath(projectDir, configPath)
	b, err := os.ReadFile(configPath)
	if err != nil {
		return projectCodegenYAML{}, configPath, err
	}
	var cfg projectCodegenYAML
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return projectCodegenYAML{}, configPath, err
	}
	return cfg, configPath, nil
}

// validateProjectDBRedisForCodegen add db * 执行前检查 database + redis 已配置（不连库）。
func validateProjectDBRedisForCodegen(projectDir, configPath string) error {
	cfg, path, err := loadProjectCodegenYAML(projectDir, configPath)
	if err != nil {
		return fmt.Errorf("read config for add db: %w", err)
	}
	if !cfg.Database.Enabled {
		return fmt.Errorf("database.enabled is false in %s (add db requires enabled database)", path)
	}
	if cfg.Database.DSN == "" {
		return fmt.Errorf("database.dsn is empty in %s", path)
	}
	driver := cfg.Database.Driver
	if driver == "" {
		driver = "postgres"
	}
	if driver != "postgres" && driver != "mysql" {
		return fmt.Errorf("add db supports postgres or mysql, got %q in %s", driver, path)
	}
	if !cfg.Redis.Enabled {
		return fmt.Errorf("redis.enabled is false in %s (add db crud generates cache-aware repository_pg)", path)
	}
	if cfg.Redis.Addr == "" {
		return fmt.Errorf("redis.addr is empty in %s", path)
	}
	return nil
}

func readProjectCodegen(projectDir, configPath string) (dsn, driver, prefix string, rules schema.CodegenRules, err error) {
	if err := validateProjectDBRedisForCodegen(projectDir, configPath); err != nil {
		return "", "", "", rules, err
	}
	cfg, _, err := loadProjectCodegenYAML(projectDir, configPath)
	if err != nil {
		return "", "", "", rules, err
	}
	configPath = resolveConfigPath(projectDir, configPath)
	driver = cfg.Database.Driver
	if driver == "" {
		driver = "postgres"
	}
	if driver != "postgres" && driver != "mysql" {
		return "", "", "", rules, fmt.Errorf("add db supports postgres or mysql, got %q", driver)
	}
	rules = schema.DefaultCodegenRules()
	if len(cfg.Codegen.CreateOmit) > 0 {
		rules.CreateOmit = cfg.Codegen.CreateOmit
	}
	if len(cfg.Codegen.UpdateOmit) > 0 {
		rules.UpdateOmit = cfg.Codegen.UpdateOmit
	}
	if cfg.Codegen.SoftDeleteColumn != "" {
		rules.SoftDeleteColumn = cfg.Codegen.SoftDeleteColumn
	}
	if !cfg.Codegen.TouchCreatedAtOnInsert {
		rules.TouchCreatedAtOnInsert = false
	}
	if !cfg.Codegen.TouchUpdatedAtOnUpdate {
		rules.TouchUpdatedAtOnUpdate = false
	}
	return cfg.Database.DSN, driver, cfg.Database.TablePrefix, rules, nil
}
