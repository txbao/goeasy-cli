package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

func GenerateDBOpenAPI(opts DBOptions) error {
	dsn, driver, prefix, rules, err := readProjectCodegen(opts.ProjectDir, opts.ConfigPath)
	if err != nil {
		return err
	}
	tables, err := resolveDBTables(opts, dsn, prefix, driver)
	if err != nil {
		return err
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	clients, err := NormalizeClients(opts.Clients)
	if err != nil {
		return err
	}
	for _, physical := range tables {
		module := resolveModuleName(opts, physical, prefix)
		meta, err := loadTableMeta(driver, dsn, opts.Schema, physical)
		if err != nil {
			return err
		}
		ct := schema.Classify(meta, module, physical, rules)
		pascal := utils.ToPascal(module)
		snake := utils.ToSnake(module)
		layout := moduleMetaFromDB(opts, module)
		for _, cl := range clients {
			content := genOpenAPIFile(projectModule, ct, pascal, snake, cl.Name, layout)
			rel := layout.OpenAPIRel(cl.Name)
			skipped, err := writeProjectFileOrSkip(opts.ProjectDir, rel, content, opts.Force)
			if err != nil {
				return err
			}
			if skipped {
				fmt.Fprintf(os.Stderr, "info: skip existing %s (use --force)\n", rel)
				continue
			}
			fmt.Printf("  created %s\n", rel)
		}
	}
	return nil
}

func colDescription(c schema.ColumnMeta) string {
	if strings.TrimSpace(c.Comment) != "" {
		return yamlQuote(c.Comment)
	}
	return yamlQuote(c.Name)
}

func yamlQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

func openAPIType(goType string) string {
	switch {
	case strings.Contains(goType, "int"):
		return "integer"
	case goType == "bool":
		return "boolean"
	case goType == "float64":
		return "number"
	default:
		return "string"
	}
}

func genOpenAPIFile(projectModule string, ct schema.ClassifiedTable, pascal, snake, httpClient string, layout HTTPRouteLayout) string {
	if httpClient == "" {
		httpClient = "admin"
	}
	basePath := "/api/v1/" + httpClient + layout.RoutePrefix()
	tableDesc := ct.TableComment
	if tableDesc == "" {
		tableDesc = ct.PhysicalName + " 模块"
	}
	var b strings.Builder
	b.WriteString("openapi: 3.0.3\n")
	b.WriteString("info:\n")
	b.WriteString(fmt.Sprintf("  title: %s API\n", pascal))
	b.WriteString(fmt.Sprintf("  description: %s\n", yamlQuote(tableDesc)))
	b.WriteString("  version: 1.0.0\n")
	b.WriteString("paths:\n")

	// List
	b.WriteString(fmt.Sprintf("  %s:\n", basePath))
	b.WriteString("    get:\n")
	b.WriteString(fmt.Sprintf("      tags: [%s]\n", yamlQuote(snake)))
	b.WriteString("      summary: 分页列表\n")
	b.WriteString(fmt.Sprintf("      operationId: list%s\n", pascal))
	b.WriteString("      parameters:\n")
	b.WriteString("        - name: page\n          in: query\n          schema: { type: integer, minimum: 1 }\n          description: 页码\n")
	b.WriteString("        - name: page_size\n          in: query\n          schema: { type: integer, minimum: 1, maximum: 200 }\n          description: 每页条数\n")
	b.WriteString("      responses:\n")
	b.WriteString("        '200':\n          description: 成功\n")
	b.WriteString("          content:\n            application/json:\n              schema:\n                $ref: '#/components/schemas/ListResponse'\n")

	// Create
	b.WriteString("    post:\n")
	b.WriteString(fmt.Sprintf("      tags: [%s]\n", yamlQuote(snake)))
	b.WriteString("      summary: 创建\n")
	b.WriteString(fmt.Sprintf("      operationId: create%s\n", pascal))
	b.WriteString("      requestBody:\n        required: true\n")
	b.WriteString("        content:\n          application/json:\n")
	b.WriteString(fmt.Sprintf("            schema:\n              $ref: '#/components/schemas/Create%sRequest'\n", pascal))
	b.WriteString("      responses:\n        '200':\n          description: 成功\n")

	// Get/Update/Delete by id
	idPath := basePath + "/{id}"
	b.WriteString(fmt.Sprintf("  %s:\n", idPath))
	b.WriteString("    get:\n")
	b.WriteString(fmt.Sprintf("      tags: [%s]\n", yamlQuote(snake)))
	b.WriteString("      summary: 详情\n")
	b.WriteString(fmt.Sprintf("      operationId: get%s\n", pascal))
	b.WriteString("      parameters:\n        - name: id\n          in: path\n          required: true\n          schema: { type: string }\n")
	b.WriteString("      responses:\n        '200':\n          description: 成功\n")
	b.WriteString("          content:\n            application/json:\n              schema:\n")
	b.WriteString(fmt.Sprintf("                $ref: '#/components/schemas/%sDTO'\n", pascal))
	b.WriteString("    put:\n")
	b.WriteString(fmt.Sprintf("      tags: [%s]\n", yamlQuote(snake)))
	b.WriteString("      summary: 更新\n")
	b.WriteString(fmt.Sprintf("      operationId: update%s\n", pascal))
	b.WriteString("      parameters:\n        - name: id\n          in: path\n          required: true\n          schema: { type: string }\n")
	b.WriteString("      requestBody:\n        required: true\n")
	b.WriteString("        content:\n          application/json:\n")
	b.WriteString(fmt.Sprintf("            schema:\n              $ref: '#/components/schemas/Update%sRequest'\n", pascal))
	b.WriteString("      responses:\n        '200':\n          description: 成功\n")
	b.WriteString("    delete:\n")
	b.WriteString(fmt.Sprintf("      tags: [%s]\n", yamlQuote(snake)))
	b.WriteString("      summary: 删除\n")
	b.WriteString(fmt.Sprintf("      operationId: delete%s\n", pascal))
	b.WriteString("      parameters:\n        - name: id\n          in: path\n          required: true\n          schema: { type: string }\n")
	b.WriteString("      responses:\n        '200':\n          description: 成功\n")

	b.WriteString("components:\n  schemas:\n")
	b.WriteString("    PaginationMeta:\n      type: object\n      properties:\n")
	b.WriteString("        page: { type: integer }\n        page_size: { type: integer }\n")
	b.WriteString("        total: { type: integer, format: int64 }\n")
	b.WriteString("        total_pages: { type: integer }\n")
	b.WriteString(fmt.Sprintf("    %sDTO:\n      type: object\n      properties:\n", pascal))
	for _, c := range ct.ReadCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("        %s:\n          type: %s\n          description: %s\n", c.Name, openAPIType(f.GoType), colDescription(c)))
	}
	b.WriteString(fmt.Sprintf("    Create%sRequest:\n      type: object\n      required: []\n      properties:\n", pascal))
	for _, c := range ct.CreateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("        %s:\n          type: %s\n          description: %s\n", c.Name, openAPIType(f.GoType), colDescription(c)))
	}
	b.WriteString(fmt.Sprintf("    Update%sRequest:\n      type: object\n      properties:\n", pascal))
	for _, c := range ct.UpdateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("        %s:\n          type: %s\n          description: %s\n", c.Name, openAPIType(f.GoType), colDescription(c)))
	}
	b.WriteString("    ListResponse:\n      type: object\n      properties:\n")
	b.WriteString("        code: { type: integer }\n        message: { type: string }\n")
	b.WriteString("        data:\n          type: object\n          properties:\n")
	b.WriteString(fmt.Sprintf("            list:\n              type: array\n              items:\n                $ref: '#/components/schemas/%sDTO'\n", pascal))
	b.WriteString("            pagination:\n              $ref: '#/components/schemas/PaginationMeta'\n")
	_ = projectModule
	return b.String()
}
