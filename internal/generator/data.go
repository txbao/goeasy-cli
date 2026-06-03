package generator

import (
	"os"
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/utils"
)

const DefaultZdgfModule = "github.com/txbao/goeasy"

type TemplateData struct {
	ProjectName   string
	ModuleName    string
	ServiceName   string
	ZdgfModule    string
	ZdgfReplace   string
	TemplateName  string
	TemplateLabel string
	ModuleSnake   string
	ModulePascal  string
	EventSnake    string
	EventPascal   string
}

func BuildProjectData(opts Options) TemplateData {
	label := templateLabel(opts.TemplateName)
	goesyReplace := opts.ZdgfReplace
	if goesyReplace == "" {
		goesyReplace = detectZdgfReplace(opts.OutputDir)
	}
	return TemplateData{
		ProjectName:   opts.ProjectName,
		ModuleName:    opts.ModuleName,
		ServiceName:   defaultString(opts.ServiceName, "service"),
		ZdgfModule:    DefaultZdgfModule,
		ZdgfReplace:   goesyReplace,
		TemplateName:  opts.TemplateName,
		TemplateLabel: label,
	}
}

func BuildModuleData(moduleName, projectModule string) TemplateData {
	snake := utils.ToSnake(moduleName)
	return TemplateData{
		ModuleName:   projectModule,
		ModuleSnake:  snake,
		ModulePascal: utils.ToPascal(moduleName),
	}
}

func BuildEventData(eventName, projectModule string) TemplateData {
	snake := utils.ToSnake(eventName)
	return TemplateData{
		ModuleName:  projectModule,
		EventSnake:  snake,
		EventPascal: utils.ToPascal(eventName),
	}
}

func templateLabel(name string) string {
	switch name {
	case "default", "project":
		return "标准微服务（DDD Lite + Gin）"
	case "monolith":
		return "单体应用（多模块路由前缀）"
	case "auth":
		return "认证服务（JWT/OAuth 配置占位，无 User/Role 实体）"
	case "system":
		return "系统服务（RBAC 集成占位）"
	case "payment":
		return "支付服务（幂等 / Outbox 目录占位）"
	default:
		return name
	}
}

func resolveTemplateRoot(name string) string {
	switch name {
	case "default":
		return "project"
	default:
		return name
	}
}

func detectZdgfReplace(outputDir string) string {
	candidates := []string{
		filepath.Join(outputDir, "..", "goesy"),
		filepath.Join(outputDir, "..", "..", "goesy"),
		filepath.Join(outputDir, "goesy"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(filepath.Join(c, "go.mod")); err == nil && !st.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}
	return ""
}

func defaultString(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func toMap(d TemplateData) map[string]any {
	return map[string]any{
		"ProjectName":   d.ProjectName,
		"ModuleName":    d.ModuleName,
		"ServiceName":   d.ServiceName,
		"ZdgfModule":    d.ZdgfModule,
		"ZdgfReplace":   d.ZdgfReplace,
		"TemplateName":  d.TemplateName,
		"TemplateLabel": d.TemplateLabel,
		"ModuleSnake":   d.ModuleSnake,
		"ModulePascal":  d.ModulePascal,
		"EventSnake":    d.EventSnake,
		"EventPascal":   d.EventPascal,
	}
}
