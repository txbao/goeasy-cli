package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/txbao/goeasy-cli/internal/utils"
)

func isStandardCRUDOperation(client, method, rawPath string, layout ModuleMeta) bool {
	if client == "" || layout.Domain == "" || layout.Resource == "" {
		return false
	}
	m := strings.ToLower(strings.TrimSpace(method))
	if m == "parameters" || m == "" {
		return false
	}
	base := "/api/v1/" + client + layout.RoutePrefix()
	p := strings.TrimSuffix(strings.TrimSpace(rawPath), "/")
	if p == base {
		return m == "get" || m == "post"
	}
	if strings.HasPrefix(p, base+"/") && strings.Contains(p, "{id}") {
		return m == "get" || m == "put" || m == "patch" || m == "delete"
	}
	return false
}

func extraRouteSuffix(client string, meta ModuleMeta, path string) string {
	base := "/api/v1/" + client + meta.RoutePrefix()
	p := strings.TrimSuffix(strings.TrimSpace(path), "/")
	if !strings.HasPrefix(p, base) {
		return p
	}
	rel := strings.TrimPrefix(p, base)
	if rel == "" {
		return "/"
	}
	return rel
}

func handlerNameFromExtra(ep OpenAPIExtraEndpoint, meta ModuleMeta) string {
	if ep.OperationID != "" {
		return utils.ToPascal(ep.OperationID)
	}
	suffix := strings.Trim(extraRouteSuffix(ep.Client, meta, ep.Path), "/")
	suffix = strings.ReplaceAll(suffix, "/", "_")
	suffix = strings.ReplaceAll(suffix, "{", "")
	suffix = strings.ReplaceAll(suffix, "}", "")
	if suffix == "" {
		return utils.ToPascal(ep.Method)
	}
	return utils.ToPascal(ep.Method) + utils.ToPascal(suffix)
}

func ginMethod(ep OpenAPIExtraEndpoint) string {
	switch strings.ToLower(ep.Method) {
	case "get":
		return "GET"
	case "post":
		return "POST"
	case "put":
		return "PUT"
	case "patch":
		return "PATCH"
	case "delete":
		return "DELETE"
	default:
		return strings.ToUpper(ep.Method)
	}
}

func genHandlerOpenAPI(goeasy string, meta ModuleMeta, extras []OpenAPIExtraEndpoint) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\n", meta.PackageName()))
	b.WriteString("import (\n\t\"net/http\"\n\n\t\"github.com/gin-gonic/gin\"\n\n")
	b.WriteString(fmt.Sprintf("\tzresp \"%s/response\"\n", goeasy))
	b.WriteString(")\n\n")
	b.WriteString("// Extra OpenAPI handlers (gen http --merge-http). Wire via RegisterOpenAPIRoutes in register_*.go.\n")
	for _, ep := range extras {
		name := handlerNameFromExtra(ep, meta)
		comment := ep.Summary
		if comment == "" {
			comment = ep.OperationID
		}
		if comment == "" {
			comment = ginMethod(ep) + " " + ep.Path
		}
		b.WriteString(fmt.Sprintf("// %s %s\n", name, comment))
		b.WriteString(fmt.Sprintf("func (h *Handler) %s(c *gin.Context) {\n", name))
		b.WriteString("\tzresp.Fail(c, http.StatusNotImplemented, \"TODO: implement " + name + "\")\n")
		b.WriteString("}\n\n")
	}
	return b.String()
}

func genRouterOpenAPI(meta ModuleMeta, extras []OpenAPIExtraEndpoint) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport \"github.com/gin-gonic/gin\"\n\n", meta.PackageName()))
	b.WriteString("func RegisterOpenAPIRoutes(r *gin.RouterGroup, h *Handler) {\n")
	b.WriteString(fmt.Sprintf("\tg := r.Group(\"%s\")\n", meta.RoutePrefix()))
	for _, ep := range extras {
		route := extraRouteSuffix(ep.Client, meta, ep.Path)
		if route == "/" {
			route = ""
		}
		b.WriteString(fmt.Sprintf("\tg.Handle(\"%s\", \"%s\", h.%s)\n", ginMethod(ep), route, handlerNameFromExtra(ep, meta)))
	}
	b.WriteString("}\n")
	return b.String()
}

func writeHTTPLayerExtras(opts GenHTTPOptions, projectModule string, contract OpenAPIContract, meta ModuleMeta, clients []ClientSurface) error {
	if len(contract.ExtraEndpoints) == 0 {
		return nil
	}
	goeasy := currentGoEasyModule()
	byClient := map[string][]OpenAPIExtraEndpoint{}
	for _, ep := range contract.ExtraEndpoints {
		byClient[ep.Client] = append(byClient[ep.Client], ep)
	}
	files := map[string]string{}
	for _, cl := range clients {
		extras := byClient[cl.Name]
		if len(extras) == 0 {
			continue
		}
		files[meta.HTTPRel(cl.Name, "handler_openapi.go")] = genHandlerOpenAPI(goeasy, meta, extras)
		files[meta.HTTPRel(cl.Name, "router_openapi.go")] = genRouterOpenAPI(meta, extras)
	}
	if len(files) == 0 {
		return nil
	}
	skipped, _, written, err := writeDBGeneratedFiles(opts.ProjectDir, files, meta.ModuleID, false)
	if err != nil {
		return err
	}
	created := len(files) - skipped
	if created > 0 {
		fmt.Fprintf(os.Stderr, "info: wrote %d openapi extra file(s); wire RegisterOpenAPIRoutes in register_%s.go\n", created, meta.Domain)
	}
	if skipped > 0 {
		fmt.Fprintf(os.Stderr, "info: %d openapi extra file(s) skipped (already exist)\n", skipped)
	}
	_ = written
	_ = projectModule
	return nil
}
