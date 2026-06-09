package generator

import (
	"fmt"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func genEntity(projectModule string, ct schema.ClassifiedTable, pascal string, meta ModuleMeta) string {
	pkg := meta.Resource
	snake := meta.ModuleID
	needsTime := schema.ColsNeedTimeImport(ct.ReadCols)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"errors\"\n", pkg))
	if needsTime {
		b.WriteString("\t\"time\"\n")
	}
	b.WriteString(")\n\n")
	b.WriteString(fmt.Sprintf("var ErrNotFound = errors.New(%q)\n\n", snake+": not found"))
	b.WriteString(fmt.Sprintf("// %s 由 add db 根据表 %s 生成。\n", pascal, ct.PhysicalName))
	b.WriteString(fmt.Sprintf("type %s struct {\n", pascal))
	for _, c := range ct.ReadCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t%s %s `%s`\n", f.Name, f.GoType, schema.EntityJSONTag(c, f.GoType)))
	}
	b.WriteString("}\n\n")
	b.WriteString(fmt.Sprintf("func NewEmpty%s() *%s {\n\treturn &%s{}\n}\n\n", pascal, pascal, pascal))
	for _, c := range schema.UniqueColumnsByName(ct.CreateCols, ct.UpdateCols) {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("func (e *%s) Set%s(v %s) { e.%s = v }\n", pascal, f.Name, f.GoType, f.Name))
	}
	b.WriteString("\n")
	// Rehydrate
	b.WriteString(fmt.Sprintf("func Rehydrate%s(", pascal))
	for i, c := range ct.ReadCols {
		f := schema.GoFieldFromColumn(c)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%s %s", f.Name, f.GoType))
	}
	b.WriteString(fmt.Sprintf(") *%s {\n\treturn &%s{", pascal, pascal))
	for _, c := range ct.ReadCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("%s: %s, ", f.Name, f.Name))
	}
	b.WriteString("}\n}\n")
	return b.String()
}

func genAggregate(projectModule string, ct schema.ClassifiedTable, pascal string, meta ModuleMeta) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\n", meta.Resource))
	b.WriteString(fmt.Sprintf("type Aggregate struct {\n\troot *%s\n}\n\n", pascal))
	b.WriteString(fmt.Sprintf("func NewAggregate(root *%s) *Aggregate {\n\treturn &Aggregate{root: root}\n}\n\n", pascal))
	b.WriteString(fmt.Sprintf("func (a *Aggregate) Root() *%s { return a.root }\n", pascal))
	return b.String()
}

func genDTO(projectModule string, ct schema.ClassifiedTable, pascal string, meta ModuleMeta) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\n", meta.Resource))
	if schema.ColsNeedTimeImport(ct.ReadCols) {
		b.WriteString("import (\n\t\"time\"\n\n")
		b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n)\n\n", meta.DomainImportPath(projectModule)))
	} else {
		b.WriteString(fmt.Sprintf("import domain \"%s\"\n\n", meta.DomainImportPath(projectModule)))
	}
	b.WriteString(fmt.Sprintf("type %sDTO struct {\n", pascal))
	for _, c := range ct.ReadCols {
		f := schema.GoFieldFromColumn(c)
		jsonName := c.Name
		b.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", f.Name, f.GoType, jsonName))
	}
	b.WriteString("}\n\n")
	b.WriteString(fmt.Sprintf("func ToDTO(agg *domain.Aggregate) %sDTO {\n\tr := agg.Root()\n\treturn %sDTO{\n", pascal, pascal))
	for _, c := range ct.ReadCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t\t%s: r.%s,\n", f.Name, f.Name))
	}
	b.WriteString("\t}\n}\n\n")
	b.WriteString(fmt.Sprintf("// ListResult 列表查询结果（供 HTTP List 使用）。\n"))
	b.WriteString(fmt.Sprintf("type ListResult struct {\n\tList  []%sDTO\n\tTotal int64\n}\n", pascal))
	return b.String()
}

func genListQuery(projectModule string, ct schema.ClassifiedTable, pascal string, meta ModuleMeta) string {
	_ = pascal
	var b strings.Builder
	b.WriteString("package query\n\nimport (\n\t\"context\"\n\n")
	b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n)\n\n", meta.DomainImportPath(projectModule)))
	b.WriteString("func (h *Handler) List(ctx context.Context, page, pageSize int) ([]*domain.Aggregate, int64, error) {\n")
	b.WriteString("\taggs, total, err := h.repo.List(ctx, page, pageSize)\n")
	b.WriteString("\tif err != nil {\n\t\treturn nil, 0, err\n\t}\n")
	b.WriteString("\tfor _, agg := range aggs {\n")
	b.WriteString("\t\tif err := h.domainSvc.Validate(ctx, agg); err != nil {\n")
	b.WriteString("\t\t\treturn nil, 0, err\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn aggs, total, nil\n}\n")
	return b.String()
}

func genAppList(meta ModuleMeta, pascal string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport \"context\"\n\n", meta.Resource))
	b.WriteString(fmt.Sprintf("func (a *Application) List(ctx context.Context, page, pageSize int) (ListResult, error) {\n"))
	b.WriteString("\taggs, total, err := a.queries.List(ctx, page, pageSize)\n")
	b.WriteString("\tif err != nil {\n\t\treturn ListResult{}, err\n\t}\n")
	b.WriteString(fmt.Sprintf("\tlist := make([]%sDTO, 0, len(aggs))\n", pascal))
	b.WriteString("\tfor _, agg := range aggs {\n")
	b.WriteString("\t\tlist = append(list, ToDTO(agg))\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn ListResult{List: list, Total: total}, nil\n}\n")
	return b.String()
}

func genGetQuery(projectModule string, ct schema.ClassifiedTable, pascal string, meta ModuleMeta) string {
	var b strings.Builder
	b.WriteString("package query\n\nimport (\n\t\"context\"\n\n")
	b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n)\n\n", meta.DomainImportPath(projectModule)))
	b.WriteString("type Handler struct {\n\trepo      domain.Repository\n\tdomainSvc *domain.DomainService\n}\n\n")
	b.WriteString("func NewHandler(repo domain.Repository, domainSvc *domain.DomainService) *Handler {\n\treturn &Handler{repo: repo, domainSvc: domainSvc}\n}\n\n")
	b.WriteString("func (h *Handler) Get(ctx context.Context, id string) (*domain.Aggregate, error) {\n")
	b.WriteString("\tagg, err := h.repo.FindByID(ctx, id)\n")
	b.WriteString("\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\tif err := h.domainSvc.Validate(ctx, agg); err != nil {\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\treturn agg, nil\n}\n")
	return b.String()
}

func genDeleteCommand() string {
	return `package command

import "context"

func (h *Handler) Delete(ctx context.Context, id string) error {
	return h.repo.Delete(ctx, id)
}
`
}

func genCreateCommand(projectModule string, ct schema.ClassifiedTable, pascal, alias string, meta ModuleMeta) string {
	var b strings.Builder
	b.WriteString("package command\n\n")
	writeCommandImportBlock(&b, projectModule, meta, ct.CreateCols, true)
	b.WriteString("type Handler struct {\n\trepo      domain.Repository\n\tdomainSvc *domain.DomainService\n}\n\n")
	b.WriteString("func NewHandler(repo domain.Repository, domainSvc *domain.DomainService) *Handler {\n\treturn &Handler{repo: repo, domainSvc: domainSvc}\n}\n\n")
	b.WriteString("type CreateCommand struct {\n")
	for _, c := range ct.CreateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t%s %s `%s`\n", f.Name, f.GoType, schema.FieldTags(c, "create")))
	}
	b.WriteString("}\n\n")
	b.WriteString("func (h *Handler) Create(ctx context.Context, cmd CreateCommand) (string, error) {\n")
	b.WriteString(fmt.Sprintf("\troot := domain.NewEmpty%s()\n", pascal))
	for _, c := range ct.CreateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\troot.Set%s(cmd.%s)\n", f.Name, f.Name))
	}
	b.WriteString("\tagg := domain.NewAggregate(root)\n")
	b.WriteString("\tif err := h.domainSvc.Validate(ctx, agg); err != nil {\n\t\treturn \"\", err\n\t}\n")
	b.WriteString("\treturn h.repo.Create(ctx, agg)\n}\n")
	return b.String()
}

func genUpdateCommand(projectModule string, ct schema.ClassifiedTable, pascal, alias string, meta ModuleMeta) string {
	var b strings.Builder
	b.WriteString("package command\n\n")
	writeCommandImportBlock(&b, projectModule, meta, ct.UpdateCols, false)
	b.WriteString("type UpdateCommand struct {\n\tID string `json:\"-\" validate:\"-\"`\n")
	for _, c := range ct.UpdateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t%s %s `%s`\n", f.Name, f.GoType, schema.FieldTags(c, "update")))
	}
	b.WriteString("}\n\n")
	b.WriteString("func (h *Handler) Update(ctx context.Context, cmd UpdateCommand) error {\n")
	b.WriteString("\tagg, err := h.repo.FindByID(ctx, cmd.ID)\n")
	b.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\troot := agg.Root()\n")
	for _, c := range ct.UpdateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\troot.Set%s(cmd.%s)\n", f.Name, f.Name))
	}
	b.WriteString("\treturn h.repo.Update(ctx, agg)\n}\n")
	return b.String()
}

// genHandler / genHandlerCrud 已迁至 http_gen_source.go（按 HTTP client 分目录）。
