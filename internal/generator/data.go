package generator

import (
	"os"
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/utils"
)

type TemplateData struct {
	ProjectName   string
	ModuleName    string
	ServiceName   string
	GoEasyModule  string
	GoeasyReplace string
	TemplateName  string
	TemplateLabel string
	ModuleSnake   string
	ModulePascal  string
	ModuleAlias   string
	EventSnake    string
	EventPascal   string
	RemoteService string // add rpcdemo：对端逻辑服务名（app_name）
	ProtoModule   string // add rpcdemo：对端 proto 模块名
	ProtoPascal   string
	ProtoPbAlias  string
	ProtoImportPath string
	GatewayName   string
	ViewName      string
}

func resolveGoEasyModule(opts Options) string {
	if opts.GoEasyModule != "" {
		return opts.GoEasyModule
	}
	if v := os.Getenv("GOEASY_MODULE"); v != "" {
		return v
	}
	return DefaultGoEasyModule
}

func currentGoEasyModule() string {
	if v := os.Getenv("GOEASY_MODULE"); v != "" {
		return v
	}
	return DefaultGoEasyModule
}

func BuildProjectData(opts Options) TemplateData {
	label := templateLabel(opts.TemplateName)
	goeasyReplace := opts.GoeasyReplace
	if goeasyReplace == "" {
		goeasyReplace = detectGoeasyReplace(opts.OutputDir)
	}
	return TemplateData{
		ProjectName:   opts.ProjectName,
		ModuleName:    opts.ModuleName,
		ServiceName:   defaultString(opts.ServiceName, "service"),
		GoEasyModule:  resolveGoEasyModule(opts),
		GoeasyReplace: goeasyReplace,
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
		ModuleAlias:  utils.ToIdent(moduleName),
		GoEasyModule: currentGoEasyModule(),
	}
}

func BuildEventData(eventName, projectModule string) TemplateData {
	snake := utils.ToSnake(eventName)
	return TemplateData{
		ModuleName:   projectModule,
		EventSnake:   snake,
		EventPascal:  utils.ToPascal(eventName),
		GoEasyModule: currentGoEasyModule(),
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

func detectGoeasyReplace(outputDir string) string {
	candidates := []string{
		filepath.Join(outputDir, "..", "goeasy"),
		filepath.Join(outputDir, "..", "..", "goeasy"),
		filepath.Join(outputDir, "goeasy"),
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
		"GoEasyModule":  d.GoEasyModule,
		"GoeasyReplace": d.GoeasyReplace,
		"TemplateName":  d.TemplateName,
		"TemplateLabel": d.TemplateLabel,
		"ModuleSnake":   d.ModuleSnake,
		"ModulePascal":  d.ModulePascal,
		"ModuleAlias":   d.ModuleAlias,
		"EventSnake":    d.EventSnake,
		"EventPascal":   d.EventPascal,
		"RemoteService":   d.RemoteService,
		"ProtoModule":     d.ProtoModule,
		"ProtoPascal":     d.ProtoPascal,
		"ProtoPbAlias":    d.ProtoPbAlias,
		"ProtoImportPath": d.ProtoImportPath,
		"GatewayName":     d.GatewayName,
		"ViewName":        d.ViewName,
	}
}
