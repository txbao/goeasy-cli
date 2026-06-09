package generator

import (
	"fmt"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func genHTTPClientDTO(projectModule, client string, ct schema.ClassifiedTable, pascal string, meta ModuleMeta) string {
	cols := readColsForClient(client, ct.ReadCols)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\n", meta.PackageName()))
	needsTime := schema.ColsNeedTimeImport(cols)
	if needsTime {
		b.WriteString("import (\n\t\"time\"\n\n")
		b.WriteString(fmt.Sprintf("\tapp \"%s\"\n", meta.AppImportPath(projectModule)))
		b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n)\n\n", meta.DomainImportPath(projectModule)))
	} else {
		b.WriteString(fmt.Sprintf("import (\n\tapp \"%s\"\n", meta.AppImportPath(projectModule)))
		b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n)\n\n", meta.DomainImportPath(projectModule)))
	}
	b.WriteString("// ResponseDTO " + client + " 端 API 响应（与 app 层 DTO 分离）。\n")
	b.WriteString("type ResponseDTO struct {\n")
	for _, c := range cols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", f.Name, f.GoType, c.Name))
	}
	b.WriteString("}\n\n")
	b.WriteString("func ToResponse(agg *domain.Aggregate) ResponseDTO {\n\td := app.ToDTO(agg)\n\treturn ToResponseDTO(&d)\n}\n\n")
	b.WriteString("func ToResponseDTO(d *app." + pascal + "DTO) ResponseDTO {\n\tif d == nil {\n\t\treturn ResponseDTO{}\n\t}\n\treturn ResponseDTO{\n")
	for _, c := range cols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t\t%s: d.%s,\n", f.Name, f.Name))
	}
	b.WriteString("\t}\n}\n")
	return b.String()
}

func genRouter(meta ModuleMeta, fullCRUD bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport \"github.com/gin-gonic/gin\"\n\n", meta.PackageName()))
	b.WriteString("func RegisterRoutes(r *gin.RouterGroup, h *Handler) {\n")
	b.WriteString(fmt.Sprintf("\tg := r.Group(\"%s\")\n", meta.RoutePrefix()))
	b.WriteString("\tg.GET(\"/:id\", h.Get)\n")
	if fullCRUD {
		b.WriteString("\tg.POST(\"\", h.Create)\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func genRouterCrud(meta ModuleMeta, fullCRUD bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport \"github.com/gin-gonic/gin\"\n\n", meta.PackageName()))
	b.WriteString("func RegisterCRUDRoutes(r *gin.RouterGroup, h *Handler) {\n")
	b.WriteString(fmt.Sprintf("\tg := r.Group(\"%s\")\n", meta.RoutePrefix()))
	b.WriteString("\tg.GET(\"\", h.List)\n")
	if fullCRUD {
		b.WriteString("\tg.PUT(\"/:id\", h.Update)\n")
		b.WriteString("\tg.DELETE(\"/:id\", h.Delete)\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func genHandler(projectModule, client string, ct schema.ClassifiedTable, pascal, alias, goeasy string, fullCRUD bool, meta ModuleMeta, style AppStyle) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"net/http\"\n\n\t\"github.com/gin-gonic/gin\"\n\n", meta.PackageName()))
	b.WriteString(fmt.Sprintf("\tapp \"%s\"\n", meta.AppImportPath(projectModule)))
	if fullCRUD && style.IsLightCQRS() {
		b.WriteString(fmt.Sprintf("\t\"%s/command\"\n", meta.AppImportPath(projectModule)))
	}
	b.WriteString(fmt.Sprintf("\tzresp \"%s/response\"\n", goeasy))
	if fullCRUD {
		b.WriteString(fmt.Sprintf("\tzvalid \"%s/validator\"\n", goeasy))
	}
	b.WriteString(")\n\n")
	b.WriteString("type Handler struct {\n\tapp *app.Application\n}\n\n")
	b.WriteString("func NewHandler(application *app.Application) *Handler {\n\treturn &Handler{app: application}\n}\n\n")
	b.WriteString("func (h *Handler) Get(c *gin.Context) {\n\tid := c.Param(\"id\")\n")
	if style.IsService() {
		b.WriteString("\tagg, err := h.app.Get(c.Request.Context(), id)\n")
	} else {
		b.WriteString("\tagg, err := h.app.Queries().Get(c.Request.Context(), id)\n")
	}
	b.WriteString("\tif err != nil {\n\t\tzresp.Fail(c, http.StatusInternalServerError, err.Error())\n\t\treturn\n\t}\n")
	b.WriteString("\tzresp.Success(c, ToResponse(agg))\n}\n\n")
	if fullCRUD {
		b.WriteString("func (h *Handler) Create(c *gin.Context) {\n")
		if style.IsService() {
			b.WriteString("\tvar cmd app.CreateCommand\n")
		} else {
			b.WriteString("\tvar cmd command.CreateCommand\n")
		}
		b.WriteString("\tif err := c.ShouldBindJSON(&cmd); err != nil {\n")
		b.WriteString("\t\tzresp.Fail(c, http.StatusBadRequest, err.Error())\n\t\treturn\n\t}\n")
		b.WriteString("\tif err := zvalid.Validate(&cmd); err != nil {\n")
		b.WriteString("\t\tzresp.Fail(c, http.StatusBadRequest, zvalid.Format(err))\n\t\treturn\n\t}\n")
		if style.IsService() {
			b.WriteString("\tid, err := h.app.Create(c.Request.Context(), cmd)\n")
		} else {
			b.WriteString("\tid, err := h.app.Commands().Create(c.Request.Context(), cmd)\n")
		}
		b.WriteString("\tif err != nil {\n\t\tzresp.Fail(c, http.StatusInternalServerError, err.Error())\n\t\treturn\n\t}\n")
		b.WriteString("\tzresp.Success(c, gin.H{\"id\": id})\n}\n")
	}
	return b.String()
}

func genHandlerCrud(projectModule, client string, ct schema.ClassifiedTable, pascal, goeasy string, fullCRUD bool, meta ModuleMeta, style AppStyle) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"net/http\"\n\n\t\"github.com/gin-gonic/gin\"\n\n", meta.PackageName()))
	if fullCRUD && style.IsLightCQRS() {
		b.WriteString(fmt.Sprintf("\t\"%s/command\"\n", meta.AppImportPath(projectModule)))
	}
	if fullCRUD && style.IsService() {
		b.WriteString(fmt.Sprintf("\tapp \"%s\"\n", meta.AppImportPath(projectModule)))
	}
	b.WriteString(fmt.Sprintf("\tzpage \"%s/pagination\"\n", goeasy))
	b.WriteString(fmt.Sprintf("\tzresp \"%s/response\"\n", goeasy))
	b.WriteString(fmt.Sprintf("\tzvalid \"%s/validator\"\n)\n\n", goeasy))
	b.WriteString("func (h *Handler) List(c *gin.Context) {\n")
	b.WriteString("\tvar req zpage.Page\n")
	b.WriteString("\tif err := c.ShouldBindJSON(&req); err != nil {\n")
	b.WriteString("\t\treq = zpage.Parse(c)\n")
	b.WriteString("\t}\n")
	b.WriteString("\treq = zpage.Normalize(req)\n")
	b.WriteString("\tif err := zvalid.Validate(&req); err != nil {\n")
	b.WriteString("\t\tzresp.Fail(c, http.StatusBadRequest, zvalid.Format(err))\n\t\treturn\n\t}\n")
	b.WriteString("\tresult, err := h.app.List(c.Request.Context(), req.Page, req.PageSize)\n")
	b.WriteString("\tif err != nil {\n\t\tzresp.Fail(c, http.StatusInternalServerError, err.Error())\n\t\treturn\n\t}\n")
	b.WriteString("\tlist := make([]ResponseDTO, 0, len(result.List))\n")
	b.WriteString("\tfor i := range result.List {\n")
	b.WriteString("\t\tlist = append(list, ToResponseDTO(&result.List[i]))\n")
	b.WriteString("\t}\n")
	b.WriteString("\tzresp.Success(c, gin.H{\n")
	b.WriteString("\t\t\"list\": list,\n")
	b.WriteString("\t\t\"pagination\": zpage.MetaFrom(req.Page, req.PageSize, result.Total),\n")
	b.WriteString("\t})\n}\n\n")
	if fullCRUD {
		b.WriteString("func (h *Handler) Update(c *gin.Context) {\n")
		b.WriteString("\tid := c.Param(\"id\")\n")
		if style.IsService() {
			b.WriteString("\tvar cmd app.UpdateCommand\n")
		} else {
			b.WriteString("\tvar cmd command.UpdateCommand\n")
		}
		b.WriteString("\tif err := c.ShouldBindJSON(&cmd); err != nil {\n")
		b.WriteString("\t\tzresp.Fail(c, http.StatusBadRequest, err.Error())\n\t\treturn\n\t}\n")
		b.WriteString("\tif err := zvalid.Validate(&cmd); err != nil {\n")
		b.WriteString("\t\tzresp.Fail(c, http.StatusBadRequest, zvalid.Format(err))\n\t\treturn\n\t}\n")
		b.WriteString("\tcmd.ID = id\n")
		if style.IsService() {
			b.WriteString("\tif err := h.app.Update(c.Request.Context(), cmd); err != nil {\n")
		} else {
			b.WriteString("\tif err := h.app.Commands().Update(c.Request.Context(), cmd); err != nil {\n")
		}
		b.WriteString("\t\tzresp.Fail(c, http.StatusInternalServerError, err.Error())\n\t\treturn\n\t}\n")
		b.WriteString("\tzresp.Success(c, gin.H{\"updated\": id})\n}\n\n")
		b.WriteString("func (h *Handler) Delete(c *gin.Context) {\n")
		b.WriteString("\tid := c.Param(\"id\")\n")
		if style.IsService() {
			b.WriteString("\tif err := h.app.Delete(c.Request.Context(), id); err != nil {\n")
		} else {
			b.WriteString("\tif err := h.app.Commands().Delete(c.Request.Context(), id); err != nil {\n")
		}
		b.WriteString("\t\tzresp.Fail(c, http.StatusInternalServerError, err.Error())\n\t\treturn\n\t}\n")
		b.WriteString("\tzresp.Success(c, gin.H{\"deleted\": id})\n}\n")
	}
	return b.String()
}

func appendHTTPClientFiles(files map[string]string, clients []ClientSurface, projectModule, goeasy string, ct schema.ClassifiedTable, pascal, alias string, meta ModuleMeta, style AppStyle) {
	for _, cl := range clients {
		files[meta.HTTPRel(cl.Name, "dto.go")] = genHTTPClientDTO(projectModule, cl.Name, ct, pascal, meta)
		files[meta.HTTPRel(cl.Name, "handler.go")] = genHandler(projectModule, cl.Name, ct, pascal, alias, goeasy, cl.FullCRUD, meta, style)
		files[meta.HTTPRel(cl.Name, "router.go")] = genRouter(meta, cl.FullCRUD)
		files[meta.HTTPRel(cl.Name, "handler_crud.go")] = genHandlerCrud(projectModule, cl.Name, ct, pascal, goeasy, cl.FullCRUD, meta, style)
		files[meta.HTTPRel(cl.Name, "router_crud.go")] = genRouterCrud(meta, cl.FullCRUD)
	}
}
