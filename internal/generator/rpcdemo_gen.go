package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeRPCDemoGeneratedFiles(opts RPCDemoOptions, projectModule string, style AppStyle, meta RPCDemoProtoMeta) error {
	cleanupStaleRPCDemoGateways(opts.ProjectDir, opts.RemoteService, filepath.Base(meta.GatewayFile))
	files := buildRPCDemoGeneratedFiles(projectModule, opts.RemoteService, style, meta)
	for rel, content := range files {
		if err := writeRPCDemoFile(opts.ProjectDir, rel, content, opts.Force); err != nil {
			return err
		}
	}
	if style.IsService() {
		removeRPCDemoLightCQRSFiles(opts.ProjectDir)
	}
	return nil
}

func cleanupStaleRPCDemoGateways(projectDir, remoteService, keepGateway string) {
	dir := filepath.Join(projectDir, "internal", "infrastructure", "rpc", remoteService)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_gateway.go") || e.Name() == keepGateway {
			continue
		}
		_ = os.Remove(filepath.Join(dir, e.Name()))
		fmt.Printf("  removed stale %s\n", filepath.ToSlash(filepath.Join("internal", "infrastructure", "rpc", remoteService, e.Name())))
	}
}

func writeRPCDemoFile(projectDir, rel, content string, force bool) error {
	target := filepath.Join(projectDir, rel)
	if !force {
		if _, err := os.Stat(target); err == nil {
			fmt.Fprintf(os.Stderr, "info: skip existing %s (use --force)\n", filepath.ToSlash(rel))
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("  created %s\n", filepath.ToSlash(rel))
	return nil
}

func removeRPCDemoLightCQRSFiles(projectDir string) {
	for _, rel := range []string{
		"internal/app/rpcdemo/command/create.go",
		"internal/app/rpcdemo/query/get.go",
	} {
		_ = os.Remove(filepath.Join(projectDir, rel))
	}
}

func buildRPCDemoGeneratedFiles(projectModule, remoteService string, style AppStyle, meta RPCDemoProtoMeta) map[string]string {
	files := map[string]string{
		"internal/app/rpcdemo/port/port.go":                genRPCDemoPort(meta),
		meta.GatewayFile:                                   genRPCDemoGateway(projectModule, remoteService, meta),
		"internal/bootstrap/register_rpcdemo.go":           genRPCDemoRegister(projectModule, remoteService, meta),
		"internal/interface/http/admin/rpcdemo/handler.go": genRPCDemoHTTPHandler(projectModule, style, meta),
		"internal/interface/http/admin/rpcdemo/dto.go":     genRPCDemoHTTPDTO(projectModule, meta),
	}
	if style.IsService() {
		files["internal/app/rpcdemo/application.go"] = genRPCDemoServiceApplication(projectModule, meta)
	} else {
		files["internal/app/rpcdemo/application.go"] = genRPCDemoLightCQRSApplication(projectModule, meta)
		files["internal/app/rpcdemo/command/create.go"] = genRPCDemoCommandCreate(projectModule, meta)
		files["internal/app/rpcdemo/query/get.go"] = genRPCDemoQueryGet(projectModule, meta)
	}
	return files
}

func rpcDemoGWParam(pascal string) string {
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func genRPCDemoStructFields(fields []rpcViewField, tagFn func(rpcViewField) string) string {
	var b strings.Builder
	for _, f := range fields {
		tag := tagFn(f)
		if tag != "" {
			b.WriteString(fmt.Sprintf("\t%s %s %s\n", f.GoName, f.GoType, tag))
		} else {
			b.WriteString(fmt.Sprintf("\t%s %s\n", f.GoName, f.GoType))
		}
	}
	return b.String()
}

func genRPCDemoPort(meta RPCDemoProtoMeta) string {
	var b strings.Builder
	b.WriteString("package port\n\nimport \"context\"\n\n")
	b.WriteString(fmt.Sprintf("// %s 调用对端 gRPC %s（由 infrastructure/rpc 实现）。\n", meta.GatewayName, meta.Pascal+"Service"))
	b.WriteString(fmt.Sprintf("type %s interface {\n", meta.GatewayName))
	b.WriteString(fmt.Sprintf("\tGetByID(ctx context.Context, id string) (*%s, error)\n", meta.ViewName))
	b.WriteString("\tCreate(ctx context.Context, in *CreateInput) (*" + meta.ViewName + ", error)\n")
	b.WriteString("}\n\n")
	b.WriteString(fmt.Sprintf("// %s 对端快照（应用层 DTO，不依赖 pb）。\n", meta.ViewName))
	b.WriteString(fmt.Sprintf("type %s struct {\n", meta.ViewName))
	b.WriteString(genRPCDemoStructFields(meta.ViewFields, func(f rpcViewField) string { return "" }))
	b.WriteString("}\n\n")
	b.WriteString("// CreateInput 对端 Create 请求字段（与 proto Create" + meta.Pascal + "Request 一致）。\n")
	b.WriteString("type CreateInput struct {\n")
	b.WriteString(genRPCDemoStructFields(meta.CreateFields, func(f rpcViewField) string { return "" }))
	b.WriteString("}\n")
	return b.String()
}

func genRPCDemoMapPbToView(meta RPCDemoProtoMeta, pbVar string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\treturn &rpcdemoport.%s{\n", meta.ViewName))
	for _, f := range meta.ViewFields {
		b.WriteString(fmt.Sprintf("\t\t%s: %s.%s(),\n", f.GoName, pbVar, f.PbGetter))
	}
	b.WriteString("\t}, nil\n")
	return b.String()
}

func genRPCDemoGateway(projectModule, remoteService string, meta RPCDemoProtoMeta) string {
	var b strings.Builder
	remotePkg := remoteService
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\n", remotePkg))
	b.WriteString(fmt.Sprintf("\t%s \"%s\"\n", meta.PbAlias, meta.ImportPath))
	b.WriteString("\t\"github.com/txbao/goeasy/grpcx\"\n\n")
	b.WriteString(fmt.Sprintf("\trpcdemoport \"%s/internal/app/rpcdemo/port\"\n)\n\n", projectModule))
	b.WriteString(fmt.Sprintf("// %s 调用 %s 的 %s（ACL）。\n", meta.GatewayName, remotePkg, meta.Pascal+"Service"))
	b.WriteString(fmt.Sprintf("type %s struct {\n\tcli  *grpcx.Client\n\tstub %s.%sServiceClient\n}\n\n", meta.GatewayName, meta.PbAlias, meta.Pascal))
	b.WriteString(fmt.Sprintf("func New%s(cli *grpcx.Client) *%s {\n", meta.GatewayName, meta.GatewayName))
	b.WriteString(fmt.Sprintf("\treturn &%s{\n\t\tcli:  cli,\n\t\tstub: %s.New%sServiceClient(cli.Conn()),\n\t}\n}\n\n", meta.GatewayName, meta.PbAlias, meta.Pascal))

	// GetByID
	b.WriteString(fmt.Sprintf("func (g *%s) GetByID(ctx context.Context, id string) (*rpcdemoport.%s, error) {\n", meta.GatewayName, meta.ViewName))
	b.WriteString(fmt.Sprintf("\tvar out *%s.%s\n", meta.PbAlias, meta.Pascal))
	b.WriteString("\terr := g.cli.Invoke(ctx, func(callCtx context.Context) error {\n\t\tvar callErr error\n")
	b.WriteString(fmt.Sprintf("\t\tout, callErr = g.stub.Get%s(callCtx, &%s.Get%sRequest{Id: id})\n", meta.Pascal, meta.PbAlias, meta.Pascal))
	b.WriteString("\t\treturn callErr\n\t})\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"Get%s id=%%s: %%w\", id, err)\n\t}\n", meta.Pascal))
	b.WriteString(genRPCDemoMapPbToView(meta, "out"))
	b.WriteString("}\n")

	// Create
	b.WriteString(fmt.Sprintf("\nfunc (g *%s) Create(ctx context.Context, in *rpcdemoport.CreateInput) (*rpcdemoport.%s, error) {\n", meta.GatewayName, meta.ViewName))
	b.WriteString(fmt.Sprintf("\tvar out *%s.%s\n", meta.PbAlias, meta.Pascal))
	b.WriteString("\terr := g.cli.Invoke(ctx, func(callCtx context.Context) error {\n\t\tvar callErr error\n")
	b.WriteString(fmt.Sprintf("\t\tout, callErr = g.stub.Create%s(callCtx, &%s.Create%sRequest{\n", meta.Pascal, meta.PbAlias, meta.Pascal))
	for _, f := range meta.CreateFields {
		b.WriteString(fmt.Sprintf("\t\t\t%s: in.%s,\n", f.GoName, f.GoName))
	}
	b.WriteString("\t\t})\n\t\treturn callErr\n\t})\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"Create%s: %%w\", err)\n\t}\n", meta.Pascal))
	b.WriteString(genRPCDemoMapPbToView(meta, "out"))
	b.WriteString("}\n")
	return b.String()
}

func genRPCDemoRegister(projectModule, remoteService string, meta RPCDemoProtoMeta) string {
	remotePkg := remoteService
	gwVar := rpcDemoGWParam(meta.Pascal) + "GW"
	return fmt.Sprintf(`package bootstrap

import (
	"fmt"

	"github.com/gin-gonic/gin"

	goeasyapp "github.com/txbao/goeasy/app"

	rpcdemoapp "%s/internal/app/rpcdemo"
	rpcdemoinfra "%s/internal/infrastructure/persistence/repository/rpcdemo"
	remote "%s/internal/infrastructure/rpc/%s"
	adminrpcdemo "%s/internal/interface/http/admin/rpcdemo"
	"%s/internal/interface/http/middleware"
)

// RegisterRPCDemo 装配并注册 rpcdemo 模块（跨服务 gRPC Gateway 示范）。
func RegisterRPCDemo(engine *gin.Engine, infra goeasyapp.HTTPInfra) error {
	cli, err := RPCClientLazy(infra, "%s")
	if err != nil {
		return fmt.Errorf("rpcdemo: %%w", err)
	}
	repo := rpcdemoinfra.NewRepository()
	%s := remote.New%s(cli)
	application := rpcdemoapp.NewApplication(repo, %s)

	apiAdmin := engine.Group("/api/v1/admin", middleware.AdminAuth(infra))
	hAdmin := adminrpcdemo.NewHandler(application)
	adminrpcdemo.RegisterRoutes(apiAdmin, hAdmin)
	return nil
}
`, projectModule, projectModule, projectModule, remotePkg, projectModule, projectModule, remoteService, gwVar, meta.GatewayName, gwVar)
}

func genRPCDemoCreateCommandStruct(meta RPCDemoProtoMeta) string {
	var b strings.Builder
	b.WriteString("type CreateCommand struct {\n")
	b.WriteString(genRPCDemoStructFields(meta.CreateFields, func(f rpcViewField) string {
		return fmt.Sprintf("`json:\"%s\"`", f.JSONName)
	}))
	b.WriteString("}\n")
	return b.String()
}

func genRPCDemoServiceApplication(projectModule string, meta RPCDemoProtoMeta) string {
	gwParam := rpcDemoGWParam(meta.Pascal)
	var b strings.Builder
	b.WriteString("package rpcdemo\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\n")
	b.WriteString(fmt.Sprintf("\t\"%s/internal/app/rpcdemo/port\"\n", projectModule))
	b.WriteString(fmt.Sprintf("\tdomain \"%s/internal/domain/rpcdemo\"\n)\n\n", projectModule))
	b.WriteString("type Application struct {\n")
	b.WriteString(fmt.Sprintf("\trepo      domain.Repository\n\t%s port.%s\n", gwParam, meta.GatewayName))
	b.WriteString("\tdomainSvc *domain.DomainService\n}\n\n")
	b.WriteString(fmt.Sprintf("func NewApplication(repo domain.Repository, %s port.%s) *Application {\n", gwParam, meta.GatewayName))
	b.WriteString(fmt.Sprintf("\treturn &Application{\n\t\trepo: repo,\n\t\t%s: %s,\n\t\tdomainSvc: domain.NewDomainService(),\n\t}\n}\n\n", gwParam, gwParam))
	b.WriteString(genRPCDemoCreateCommandStruct(meta))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("func (a *Application) Get(ctx context.Context, id string) (*port.%s, error) {\n", meta.ViewName))
	b.WriteString(fmt.Sprintf("\treturn a.%s.GetByID(ctx, id)\n", gwParam))
	b.WriteString("}\n\n")
	b.WriteString("func (a *Application) Create(ctx context.Context, cmd CreateCommand) (*port." + meta.ViewName + ", error) {\n")
	b.WriteString("\tin := &port.CreateInput{\n")
	for _, f := range meta.CreateFields {
		b.WriteString(fmt.Sprintf("\t\t%s: cmd.%s,\n", f.GoName, f.GoName))
	}
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\tview, err := a.%s.Create(ctx, in)\n", gwParam))
	b.WriteString("\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\t_ = domain.NewRpcdemoCreated(fmt.Sprintf(\"%d\", view.ID))\n")
	b.WriteString("\treturn view, nil\n")
	b.WriteString("}\n")
	return b.String()
}

func genRPCDemoLightCQRSApplication(projectModule string, meta RPCDemoProtoMeta) string {
	gwParam := rpcDemoGWParam(meta.Pascal)
	return fmt.Sprintf(`package rpcdemo

import (
	"%s/internal/app/rpcdemo/command"
	"%s/internal/app/rpcdemo/port"
	"%s/internal/app/rpcdemo/query"
	domain "%s/internal/domain/rpcdemo"
)

type Application struct {
	queries  *query.Handler
	commands *command.Handler
}

func NewApplication(repo domain.Repository, %s port.%s) *Application {
	return &Application{
		queries:  query.NewHandler(%s),
		commands: command.NewHandler(%s),
	}
}

func (a *Application) Queries() *query.Handler    { return a.queries }
func (a *Application) Commands() *command.Handler { return a.commands }
`, projectModule, projectModule, projectModule, projectModule, gwParam, meta.GatewayName, gwParam, gwParam)
}

func genRPCDemoCommandCreate(projectModule string, meta RPCDemoProtoMeta) string {
	gwParam := rpcDemoGWParam(meta.Pascal)
	var b strings.Builder
	b.WriteString("package command\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\n")
	b.WriteString(fmt.Sprintf("\t\"%s/internal/app/rpcdemo/port\"\n", projectModule))
	b.WriteString(fmt.Sprintf("\tdomain \"%s/internal/domain/rpcdemo\"\n)\n\n", projectModule))
	b.WriteString(fmt.Sprintf("type Handler struct {\n\t%s port.%s\n}\n\n", gwParam, meta.GatewayName))
	b.WriteString(fmt.Sprintf("func NewHandler(%s port.%s) *Handler {\n", gwParam, meta.GatewayName))
	b.WriteString(fmt.Sprintf("\treturn &Handler{%s: %s}\n}\n\n", gwParam, gwParam))
	b.WriteString(genRPCDemoCreateCommandStruct(meta))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("func (h *Handler) Create(ctx context.Context, cmd CreateCommand) (*port.%s, error) {\n", meta.ViewName))
	b.WriteString("\tin := &port.CreateInput{\n")
	for _, f := range meta.CreateFields {
		b.WriteString(fmt.Sprintf("\t\t%s: cmd.%s,\n", f.GoName, f.GoName))
	}
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\tview, err := h.%s.Create(ctx, in)\n", gwParam))
	b.WriteString("\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\t_ = domain.NewRpcdemoCreated(fmt.Sprintf(\"%d\", view.ID))\n")
	b.WriteString("\treturn view, nil\n")
	b.WriteString("}\n")
	return b.String()
}

func genRPCDemoQueryGet(projectModule string, meta RPCDemoProtoMeta) string {
	gwParam := rpcDemoGWParam(meta.Pascal)
	return fmt.Sprintf(`package query

import (
	"context"

	"%s/internal/app/rpcdemo/port"
)

type Handler struct {
	%s port.%s
}

func NewHandler(%s port.%s) *Handler {
	return &Handler{%s: %s}
}

func (h *Handler) Get(ctx context.Context, id string) (*port.%s, error) {
	return h.%s.GetByID(ctx, id)
}
`, projectModule, gwParam, meta.GatewayName, gwParam, meta.GatewayName, gwParam, gwParam, meta.ViewName, gwParam)
}

func genRPCDemoHTTPDTO(projectModule string, meta RPCDemoProtoMeta) string {
	var b strings.Builder
	b.WriteString("package rpcdemo\n\nimport (\n")
	b.WriteString(fmt.Sprintf("\trpcdemoport \"%s/internal/app/rpcdemo/port\"\n", projectModule))
	b.WriteString(")\n\n")
	b.WriteString("// ResponseDTO HTTP 响应，字段与 proto 实体消息一致。\n")
	b.WriteString("type ResponseDTO struct {\n")
	b.WriteString(genRPCDemoStructFields(meta.ViewFields, func(f rpcViewField) string {
		return fmt.Sprintf("`json:\"%s\"`", f.JSONName)
	}))
	b.WriteString("}\n\n")
	b.WriteString(fmt.Sprintf("func ToResponse(v *rpcdemoport.%s) ResponseDTO {\n", meta.ViewName))
	b.WriteString("\tif v == nil {\n\t\treturn ResponseDTO{}\n\t}\n")
	b.WriteString("\treturn ResponseDTO{\n")
	for _, f := range meta.ViewFields {
		b.WriteString(fmt.Sprintf("\t\t%s: v.%s,\n", f.GoName, f.GoName))
	}
	b.WriteString("\t}\n}\n")
	return b.String()
}

func genRPCDemoHTTPHandler(projectModule string, style AppStyle, meta RPCDemoProtoMeta) string {
	var b strings.Builder
	b.WriteString("package rpcdemo\n\nimport (\n\t\"net/http\"\n\n\t\"github.com/gin-gonic/gin\"\n\n")
	b.WriteString(fmt.Sprintf("\tapp \"%s/internal/app/rpcdemo\"\n", projectModule))
	if style.IsLightCQRS() {
		b.WriteString(fmt.Sprintf("\t\"%s/internal/app/rpcdemo/command\"\n", projectModule))
	}
	b.WriteString("\tzresp \"github.com/txbao/goeasy/response\"\n")
	b.WriteString("\tzvalid \"github.com/txbao/goeasy/validator\"\n")
	b.WriteString(")\n\n")
	b.WriteString("type Handler struct {\n\tapp *app.Application\n}\n\n")
	b.WriteString("func NewHandler(application *app.Application) *Handler {\n\treturn &Handler{app: application}\n}\n\n")
	b.WriteString("func (h *Handler) Get(c *gin.Context) {\n\tid := c.Param(\"id\")\n")
	if style.IsService() {
		b.WriteString("\tview, err := h.app.Get(c.Request.Context(), id)\n")
	} else {
		b.WriteString("\tview, err := h.app.Queries().Get(c.Request.Context(), id)\n")
	}
	b.WriteString("\tif err != nil {\n\t\tzresp.Fail(c, http.StatusInternalServerError, err.Error())\n\t\treturn\n\t}\n")
	b.WriteString("\tzresp.Success(c, ToResponse(view))\n}\n\n")
	b.WriteString("func (h *Handler) Create(c *gin.Context) {\n")
	if style.IsService() {
		b.WriteString("\tvar cmd app.CreateCommand\n")
	} else {
		b.WriteString("\tvar cmd command.CreateCommand\n")
	}
	b.WriteString("\tif err := c.ShouldBindJSON(&cmd); err != nil {\n\t\tzresp.Fail(c, http.StatusBadRequest, err.Error())\n\t\treturn\n\t}\n")
	b.WriteString("\tif err := zvalid.Validate(&cmd); err != nil {\n\t\tzresp.Fail(c, http.StatusBadRequest, zvalid.Format(err))\n\t\treturn\n\t}\n")
	if style.IsService() {
		b.WriteString("\tview, err := h.app.Create(c.Request.Context(), cmd)\n")
	} else {
		b.WriteString("\tview, err := h.app.Commands().Create(c.Request.Context(), cmd)\n")
	}
	b.WriteString("\tif err != nil {\n\t\tzresp.Fail(c, http.StatusInternalServerError, err.Error())\n\t\treturn\n\t}\n")
	b.WriteString("\tzresp.Success(c, ToResponse(view))\n}\n")
	return b.String()
}
