package generator

import (
	"fmt"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func buildDBAppLayerFiles(style AppStyle, projectModule string, ct schema.ClassifiedTable, pascal, alias string, meta ModuleMeta) map[string]string {
	files := map[string]string{
		meta.appRel("dto.go"): genDTO(projectModule, ct, pascal, meta),
	}
	switch style {
	case AppStyleService:
		files[meta.appRel("application.go")] = genServiceApplication(projectModule, ct, pascal, alias, meta)
	default:
		files[meta.appRel("list.go")] = genAppList(meta, pascal)
		files[meta.appRel("command/create.go")] = genCreateCommand(projectModule, ct, pascal, alias, meta)
		files[meta.appRel("command/update.go")] = genUpdateCommand(projectModule, ct, pascal, alias, meta)
		files[meta.appRel("command/delete.go")] = genDeleteCommand()
		files[meta.appRel("query/get.go")] = genGetQuery(projectModule, ct, pascal, meta)
		files[meta.appRel("query/list.go")] = genListQuery(projectModule, ct, pascal, meta)
	}
	return files
}

// genModuleSkeletonServiceApplication add module 用：Get/Create/Update/Delete，无 List。
func genModuleSkeletonServiceApplication(projectModule string, meta ModuleMeta, pascal string, withAudit bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"context\"\n\n", meta.Resource))
	b.WriteString(auditServiceImports("", withAudit))
	b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n)\n\n", meta.DomainImportPath(projectModule)))
	b.WriteString("type Application struct {\n\trepo      domain.Repository\n\tdomainSvc *domain.DomainService\n")
	b.WriteString(auditServiceStructField(withAudit))
	b.WriteString("}\n\n")
	b.WriteString("func NewApplication(repo domain.Repository, domainSvc *domain.DomainService" + auditServiceCtorParams(withAudit) + ") *Application {\n")
	b.WriteString(auditServiceCtorBody(withAudit))
	b.WriteString("\treturn &Application{repo: repo, domainSvc: domainSvc" + auditServiceCtorAssign(withAudit) + "}\n}\n\n")
	b.WriteString("type CreateCommand struct {\n\tID string `json:\"id\" validate:\"required\"`\n}\n\n")
	b.WriteString("type UpdateCommand struct {\n\tID string `json:\"-\" validate:\"-\"`\n}\n\n")
	b.WriteString(fmt.Sprintf("func (a *Application) Get(ctx context.Context, id string) (*domain.Aggregate, error) {\n"))
	b.WriteString("\tagg, err := a.repo.FindByID(ctx, id)\n")
	b.WriteString("\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\tif err := a.domainSvc.Validate(ctx, agg); err != nil {\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\treturn agg, nil\n}\n\n")
	b.WriteString("func (a *Application) Create(ctx context.Context, cmd CreateCommand) (string, error) {\n")
	b.WriteString("\tagg := domain.NewAggregate(cmd.ID)\n")
	b.WriteString("\tif err := a.domainSvc.Validate(ctx, agg); err != nil {\n\t\treturn \"\", err\n\t}\n")
	b.WriteString("\t_ = domain.New" + pascal + "Created(cmd.ID)\n")
	b.WriteString("\tif err := a.repo.Save(ctx, agg); err != nil {\n\t\treturn \"\", err\n\t}\n")
	if withAudit {
		b.WriteString(auditRecordStub(meta.ModuleID, "create", meta.Resource, "cmd.ID"))
	}
	b.WriteString("\n\treturn cmd.ID, nil\n}\n\n")
	b.WriteString("func (a *Application) Update(ctx context.Context, cmd UpdateCommand) error {\n")
	b.WriteString("\tagg, err := a.repo.FindByID(ctx, cmd.ID)\n")
	b.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\tif err := agg.Enable(); err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\tif err := a.repo.Save(ctx, agg); err != nil {\n\t\treturn err\n\t}\n")
	if withAudit {
		b.WriteString(auditRecordStub(meta.ModuleID, "update", meta.Resource, "cmd.ID"))
	}
	b.WriteString("\n\treturn nil\n}\n\n")
	b.WriteString("func (a *Application) Delete(ctx context.Context, id string) error {\n")
	b.WriteString("\tif err := a.repo.Delete(ctx, id); err != nil {\n\t\treturn err\n\t}\n")
	if withAudit {
		b.WriteString(auditRecordStub(meta.ModuleID, "delete", meta.Resource, "id"))
	}
	b.WriteString("\n\treturn nil\n}\n")
	return b.String()
}

// genModuleServiceApplication add crud 用：在骨架基础上补 List。
func genModuleServiceApplication(projectModule string, meta ModuleMeta, pascal string, withAudit bool) string {
	var b strings.Builder
	b.WriteString(genModuleSkeletonServiceApplication(projectModule, meta, pascal, withAudit))
	b.WriteString("\n\n")
	b.WriteString(genModuleServiceListMethod(pascal))
	return b.String()
}

func genModuleServiceListMethod(pascal string) string {
	var b strings.Builder
	b.WriteString("func (a *Application) List(ctx context.Context, page, pageSize int) (ListResult, error) {\n")
	b.WriteString("\taggs, total, err := a.repo.List(ctx, page, pageSize)\n")
	b.WriteString("\tif err != nil {\n\t\treturn ListResult{}, err\n\t}\n")
	b.WriteString("\tfor _, agg := range aggs {\n")
	b.WriteString("\t\tif err := a.domainSvc.Validate(ctx, agg); err != nil {\n")
	b.WriteString("\t\t\treturn ListResult{}, err\n\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\tlist := make([]%sDTO, 0, len(aggs))\n", pascal))
	b.WriteString("\tfor _, agg := range aggs {\n")
	b.WriteString("\t\tlist = append(list, ToDTO(agg))\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn ListResult{List: list, Total: total}, nil\n}\n")
	return b.String()
}

func genServiceApplication(projectModule string, ct schema.ClassifiedTable, pascal, alias string, meta ModuleMeta) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"context\"\n\n", meta.Resource))
	b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n)\n\n", meta.DomainImportPath(projectModule)))
	b.WriteString("type Application struct {\n\trepo      domain.Repository\n\tdomainSvc *domain.DomainService\n}\n\n")
	b.WriteString("func NewApplication(repo domain.Repository, domainSvc *domain.DomainService) *Application {\n")
	b.WriteString("\treturn &Application{repo: repo, domainSvc: domainSvc}\n}\n\n")

	b.WriteString("type CreateCommand struct {\n")
	for _, c := range ct.CreateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t%s %s `%s`\n", f.Name, f.GoType, schema.FieldTags(c, "create")))
	}
	b.WriteString("}\n\n")

	b.WriteString("type UpdateCommand struct {\n\tID string `json:\"-\" validate:\"-\"`\n")
	for _, c := range ct.UpdateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t%s %s `%s`\n", f.Name, f.GoType, schema.FieldTags(c, "update")))
	}
	b.WriteString("}\n\n")

	b.WriteString("func (a *Application) Create(ctx context.Context, cmd CreateCommand) (string, error) {\n")
	b.WriteString(fmt.Sprintf("\troot := domain.NewEmpty%s()\n", pascal))
	for _, c := range ct.CreateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\troot.Set%s(cmd.%s)\n", f.Name, f.Name))
	}
	b.WriteString("\tagg := domain.NewAggregate(root)\n")
	b.WriteString("\tif err := a.domainSvc.Validate(ctx, agg); err != nil {\n\t\treturn \"\", err\n\t}\n")
	b.WriteString("\treturn a.repo.Create(ctx, agg)\n}\n\n")

	b.WriteString("func (a *Application) Update(ctx context.Context, cmd UpdateCommand) error {\n")
	b.WriteString("\tagg, err := a.repo.FindByID(ctx, cmd.ID)\n")
	b.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\troot := agg.Root()\n")
	for _, c := range ct.UpdateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\troot.Set%s(cmd.%s)\n", f.Name, f.Name))
	}
	b.WriteString("\treturn a.repo.Update(ctx, agg)\n}\n\n")

	b.WriteString("func (a *Application) Delete(ctx context.Context, id string) error {\n")
	b.WriteString("\treturn a.repo.Delete(ctx, id)\n}\n\n")

	b.WriteString("func (a *Application) Get(ctx context.Context, id string) (*domain.Aggregate, error) {\n")
	b.WriteString("\tagg, err := a.repo.FindByID(ctx, id)\n")
	b.WriteString("\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\tif err := a.domainSvc.Validate(ctx, agg); err != nil {\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\treturn agg, nil\n}\n\n")

	b.WriteString("func (a *Application) List(ctx context.Context, page, pageSize int) (ListResult, error) {\n")
	b.WriteString("\taggs, total, err := a.repo.List(ctx, page, pageSize)\n")
	b.WriteString("\tif err != nil {\n\t\treturn ListResult{}, err\n\t}\n")
	b.WriteString("\tfor _, agg := range aggs {\n")
	b.WriteString("\t\tif err := a.domainSvc.Validate(ctx, agg); err != nil {\n")
	b.WriteString("\t\t\treturn ListResult{}, err\n\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\tlist := make([]%sDTO, 0, len(aggs))\n", pascal))
	b.WriteString("\tfor _, agg := range aggs {\n")
	b.WriteString("\t\tlist = append(list, ToDTO(agg))\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn ListResult{List: list, Total: total}, nil\n}\n")
	return b.String()
}

func genModuleServiceCRUDLayer(projectModule string, meta ModuleMeta, pascal string, withAudit bool) map[string]string {
	return map[string]string{
		meta.domainRel("repository.go"):           genModuleDomainRepository(meta.Resource),
		persistenceRepoRel(meta, "repository.go"): genModuleMemoryRepository(projectModule, meta),
		meta.appRel("application.go"):             genModuleServiceApplication(projectModule, meta, pascal, withAudit),
		meta.appRel("dto.go"):                     genModuleDTO(projectModule, meta, pascal),
	}
}

func genModuleLightCQRSApplication(projectModule string, meta ModuleMeta, withAudit bool) string {
	appImport := meta.AppImportPath(projectModule)
	auditImport := ""
	ctorExtraParam := ""
	ctorExtraBody := ""
	handlerExtraParam := ""
	if withAudit {
		auditImport = "\t\"github.com/txbao/goeasy/audit\"\n"
		ctorExtraParam = ", recorder audit.Recorder"
		ctorExtraBody = "\tif recorder == nil {\n\t\trecorder = audit.NopRecorder{}\n\t}\n"
		handlerExtraParam = ", recorder"
	}
	return fmt.Sprintf(`package %s

import (
	"%s/command"
	"%s/query"
%s	domain "%s"
)

type Application struct {
	queries  *query.Handler
	commands *command.Handler
}

func NewApplication(repo domain.Repository, domainSvc *domain.DomainService%s) *Application {
%s	return &Application{
		queries:  query.NewHandler(repo, domainSvc),
		commands: command.NewHandler(repo, domainSvc%s),
	}
}

func (a *Application) Queries() *query.Handler   { return a.queries }
func (a *Application) Commands() *command.Handler { return a.commands }
`, meta.Resource, appImport, appImport, auditImport, meta.DomainImportPath(projectModule), ctorExtraParam, ctorExtraBody, handlerExtraParam)
}

func genModuleCommandCreate(projectModule string, meta ModuleMeta, pascal string, withAudit bool) string {
	auditImports := auditCommandImports("", withAudit)
	stub := ""
	if withAudit {
		stub = auditCommandRecordStub(meta.ModuleID, "create", meta.Resource, "cmd.ID")
	}
	return fmt.Sprintf(`package command

import (
	"context"

%s	domain "%s"
)

type Handler struct {
	repo      domain.Repository
	domainSvc *domain.DomainService
%s}

func NewHandler(repo domain.Repository, domainSvc *domain.DomainService%s) *Handler {
%s	return &Handler{repo: repo, domainSvc: domainSvc%s}
}

type CreateCommand struct {
	ID string `+"`json:\"id\" validate:\"required\"`"+`
}

func (h *Handler) Create(ctx context.Context, cmd CreateCommand) (string, error) {
	agg := domain.NewAggregate(cmd.ID)
	if err := h.domainSvc.Validate(ctx, agg); err != nil {
		return "", err
	}
	_ = domain.New%sCreated(cmd.ID)
	if err := h.repo.Save(ctx, agg); err != nil {
		return "", err
	}
%s
	return cmd.ID, nil
}
`, auditImports, meta.DomainImportPath(projectModule),
		auditCommandHandlerField(withAudit),
		auditCommandCtorParams(withAudit),
		auditCommandCtorBody(withAudit),
		auditCommandCtorAssign(withAudit),
		pascal, stub)
}

func genModuleCommandUpdate(meta ModuleMeta, withAudit bool) string {
	stub := ""
	if withAudit {
		stub = auditCommandRecordStub(meta.ModuleID, "update", meta.Resource, "cmd.ID")
	}
	auditImports := ""
	if withAudit {
		auditImports = "import (\n\t\"context\"\n\n\t\"github.com/txbao/goeasy/audit\"\n\t\"github.com/txbao/goeasy/contextx\"\n)\n\n"
	} else {
		auditImports = "import \"context\"\n\n"
	}
	return `package command

` + auditImports + `type UpdateCommand struct {
	ID string ` + "`json:\"-\" validate:\"-\"`" + `
}

func (h *Handler) Update(ctx context.Context, cmd UpdateCommand) error {
	agg, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if err := agg.Enable(); err != nil {
		return err
	}
	if err := h.repo.Save(ctx, agg); err != nil {
		return err
	}
` + stub + `
	return nil
}
`
}

func genModuleCommandDelete(meta ModuleMeta, withAudit bool) string {
	stub := ""
	if withAudit {
		stub = auditCommandRecordStub(meta.ModuleID, "delete", meta.Resource, "id")
	}
	auditImports := ""
	if withAudit {
		auditImports = "import (\n\t\"context\"\n\n\t\"github.com/txbao/goeasy/audit\"\n\t\"github.com/txbao/goeasy/contextx\"\n)\n\n"
	} else {
		auditImports = "import \"context\"\n\n"
	}
	return `package command

` + auditImports + `func (h *Handler) Delete(ctx context.Context, id string) error {
	if err := h.repo.Delete(ctx, id); err != nil {
		return err
	}
` + stub + `
	return nil
}
`
}

func genModuleLightCQRSCRUDLayer(projectModule string, meta ModuleMeta, pascal string, withAudit bool) map[string]string {
	return map[string]string{
		meta.domainRel("repository.go"):           genModuleDomainRepository(meta.Resource),
		persistenceRepoRel(meta, "repository.go"): genModuleMemoryRepository(projectModule, meta),
		meta.appRel("application.go"):             genModuleLightCQRSApplication(projectModule, meta, withAudit),
		meta.appRel("command/create.go"):          genModuleCommandCreate(projectModule, meta, pascal, withAudit),
		meta.appRel("command/update.go"):          genModuleCommandUpdate(meta, withAudit),
		meta.appRel("command/delete.go"):          genModuleCommandDelete(meta, withAudit),
		meta.appRel("query/get.go"):               genGetQuery(projectModule, schema.ClassifiedTable{ModuleName: meta.ModuleID}, pascal, meta),
		meta.appRel("query/list.go"):              genModuleListQuery(projectModule, meta),
		meta.appRel("list.go"):                    genAppList(meta, pascal),
		meta.appRel("dto.go"):                     genModuleDTO(projectModule, meta, pascal),
	}
}
