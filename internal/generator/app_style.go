package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/utils"

	"gopkg.in/yaml.v3"
)

// AppStyle 应用层生成风格。
type AppStyle string

const (
	AppStyleService    AppStyle = "service"
	AppStyleLightCQRS  AppStyle = "light_cqrs"
	AppStyleFullCQRS   AppStyle = "full_cqrs"
)

const fullCQRSGuide = "see goeasy-cli/docs/guide/18-app-style.md"

// ParseAppStyle 解析 CLI / 配置取值；支持别名 light→light_cqrs、full→full_cqrs。
func ParseAppStyle(raw string) (AppStyle, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return AppStyleService, nil
	}
	switch s {
	case "service":
		return AppStyleService, nil
	case "light", "light_cqrs":
		return AppStyleLightCQRS, nil
	case "full", "full_cqrs":
		return AppStyleFullCQRS, nil
	default:
		return "", fmt.Errorf("invalid app_style %q (use service, light_cqrs, or full_cqrs)", raw)
	}
}

// ValidateForGenerate 校验是否可由 CLI 生成代码。
func (s AppStyle) ValidateForGenerate() error {
	if s == AppStyleFullCQRS {
		return fmt.Errorf("app_style full_cqrs is not supported by CLI code generation (%s)", fullCQRSGuide)
	}
	return nil
}

func (s AppStyle) IsService() bool   { return s == AppStyleService }
func (s AppStyle) IsLightCQRS() bool { return s == AppStyleLightCQRS }

type moduleAppStyleEntry struct {
	AppStyle string `yaml:"app_style"`
	Resource string `yaml:"resource"`
}

type domainAppStyleConfig struct {
	AppStyle string                         `yaml:"app_style"`
	Modules  map[string]moduleAppStyleEntry `yaml:"modules"`
}

type codegenAppStyleConfig struct {
	AppStyle string                          `yaml:"app_style"`
	Domains  map[string]domainAppStyleConfig `yaml:"domains"`
}

// ResolveAppStyle 按 CLI > 模块 > 域 > 项目默认 > 命令默认 解析。
func ResolveAppStyle(projectDir, configPath, cliFlag, moduleID string, meta ModuleMeta, commandDefault AppStyle) (AppStyle, error) {
	if strings.TrimSpace(cliFlag) != "" {
		style, err := ParseAppStyle(cliFlag)
		if err != nil {
			return "", err
		}
		return style, style.ValidateForGenerate()
	}

	cfg := readCodegenAppStyleConfig(projectDir, configPath)
	if meta.Domain != "" {
		if dc, ok := cfg.Domains[meta.Domain]; ok {
			if ent, ok := dc.Modules[moduleID]; ok && strings.TrimSpace(ent.AppStyle) != "" {
				style, err := ParseAppStyle(ent.AppStyle)
				if err != nil {
					return "", err
				}
				return style, style.ValidateForGenerate()
			}
			if strings.TrimSpace(dc.AppStyle) != "" {
				style, err := ParseAppStyle(dc.AppStyle)
				if err != nil {
					return "", err
				}
				return style, style.ValidateForGenerate()
			}
		}
	}
	if strings.TrimSpace(cfg.AppStyle) != "" {
		style, err := ParseAppStyle(cfg.AppStyle)
		if err != nil {
			return "", err
		}
		return style, style.ValidateForGenerate()
	}
	if commandDefault == "" {
		commandDefault = AppStyleService
	}
	return commandDefault, commandDefault.ValidateForGenerate()
}

func readCodegenAppStyleConfig(projectDir, configPath string) codegenAppStyleConfig {
	cfg := codegenAppStyleConfig{}
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(projectDir, configPath)
	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}
	var raw struct {
		Codegen codegenAppStyleConfig `yaml:"codegen"`
	}
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return cfg
	}
	return raw.Codegen
}

func resolveAppStyleForModule(opts ModuleOptions) (AppStyle, error) {
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	return ResolveAppStyle(opts.ProjectDir, opts.ConfigPath, opts.AppStyle, utils.ToSnake(opts.ModuleName), meta, AppStyleService)
}
