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
func genModuleSkeletonServiceApplication(projectModule string, meta ModuleMeta, pascal string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"context\"\n\n", meta.Resource))
	b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n)\n\n", meta.DomainImportPath(projectModule)))
	b.WriteString("type Application struct {\n\trepo      domain.Repository\n\tdomainSvc *domain.DomainService\n}\n\n")
	b.WriteString("func NewApplication(repo domain.Repository, domainSvc *domain.DomainService) *Application {\n")
	b.WriteString("\treturn &Application{repo: repo, domainSvc: domainSvc}\n}\n\n")
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
	b.WriteString("\treturn cmd.ID, nil\n}\n\n")
	b.WriteString("func (a *Application) Update(ctx context.Context, cmd UpdateCommand) error {\n")
	b.WriteString("\tagg, err := a.repo.FindByID(ctx, cmd.ID)\n")
	b.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\tif err := agg.Enable(); err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\treturn a.repo.Save(ctx, agg)\n}\n\n")
	b.WriteString("func (a *Application) Delete(ctx context.Context, id string) error {\n")
	b.WriteString("\treturn a.repo.Delete(ctx, id)\n}\n")
	return b.String()
}

// genModuleServiceApplication add crud 用：在骨架基础上补 List。
func genModuleServiceApplication(projectModule string, meta ModuleMeta, pascal string) string {
	var b strings.Builder
	b.WriteString(genModuleSkeletonServiceApplication(projectModule, meta, pascal))
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

func genModuleServiceCRUDLayer(projectModule string, meta ModuleMeta, pascal string) map[string]string {
	return map[string]string{
		meta.domainRel("repository.go"):             genModuleDomainRepository(meta.Resource),
		persistenceRepoRel(meta, "repository.go"):   genModuleMemoryRepository(projectModule, meta),
		meta.appRel("application.go"):               genModuleServiceApplication(projectModule, meta, pascal),
		meta.appRel("dto.go"):                       genModuleDTO(projectModule, meta, pascal),
	}
}

func genModuleLightCQRSApplication(projectModule string, meta ModuleMeta) string {
	appImport := meta.AppImportPath(projectModule)
	return fmt.Sprintf(`package %s

import (
	"%s/command"
	"%s/query"
	domain "%s"
)

type Application struct {
	queries  *query.Handler
	commands *command.Handler
}

func NewApplication(repo domain.Repository, domainSvc *domain.DomainService) *Application {
	return &Application{
		queries:  query.NewHandler(repo, domainSvc),
		commands: command.NewHandler(repo, domainSvc),
	}
}

func (a *Application) Queries() *query.Handler   { return a.queries }
func (a *Application) Commands() *command.Handler { return a.commands }
`, meta.Resource, appImport, appImport, meta.DomainImportPath(projectModule))
}

func genModuleCommandCreate(projectModule string, meta ModuleMeta, pascal string) string {
	return fmt.Sprintf(`package command

import (
	"context"

	domain "%s"
)

type Handler struct {
	repo      domain.Repository
	domainSvc *domain.DomainService
}

func NewHandler(repo domain.Repository, domainSvc *domain.DomainService) *Handler {
	return &Handler{repo: repo, domainSvc: domainSvc}
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
	return cmd.ID, nil
}
`, meta.DomainImportPath(projectModule), pascal)
}

func genModuleCommandUpdate() string {
	return `package command

import "context"

type UpdateCommand struct {
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
	return h.repo.Save(ctx, agg)
}
`
}

func genModuleLightCQRSCRUDLayer(projectModule string, meta ModuleMeta, pascal string) map[string]string {
	return map[string]string{
		meta.domainRel("repository.go"):             genModuleDomainRepository(meta.Resource),
		persistenceRepoRel(meta, "repository.go"):   genModuleMemoryRepository(projectModule, meta),
		meta.appRel("application.go"):               genModuleLightCQRSApplication(projectModule, meta),
		meta.appRel("command/create.go"):            genModuleCommandCreate(projectModule, meta, pascal),
		meta.appRel("command/update.go"):            genModuleCommandUpdate(),
		meta.appRel("command/delete.go"):            genDeleteCommand(),
		meta.appRel("query/get.go"):                 genGetQuery(projectModule, schema.ClassifiedTable{ModuleName: meta.ModuleID}, pascal, meta),
		meta.appRel("query/list.go"):                genModuleListQuery(projectModule, meta),
		meta.appRel("list.go"):                      genAppList(meta, pascal),
		meta.appRel("dto.go"):                       genModuleDTO(projectModule, meta, pascal),
	}
}
