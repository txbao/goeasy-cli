package generator

import (
	"fmt"
	"strings"
)

func genSharedRPCPort(meta RPCClientMeta) string {
	var b strings.Builder
	b.WriteString("package port\n\nimport \"context\"\n\n")
	b.WriteString(fmt.Sprintf("// %s 调用对端 gRPC %s（由 infrastructure/rpc/%s 实现）。\n",
		meta.GatewayName, meta.Pascal+"Service", filepathBaseDir(meta.GatewayFile)))
	b.WriteString(fmt.Sprintf("type %s interface {\n", meta.GatewayName))
	if meta.Methods.Get {
		b.WriteString(fmt.Sprintf("\tGetByID(ctx context.Context, id string) (*%s, error)\n", meta.ViewName))
	}
	if meta.Methods.Create {
		b.WriteString(fmt.Sprintf("\tCreate(ctx context.Context, in *CreateInput) (*%s, error)\n", meta.ViewName))
	}
	if meta.Methods.Update {
		b.WriteString(fmt.Sprintf("\tUpdate(ctx context.Context, in *UpdateInput) (*%s, error)\n", meta.ViewName))
	}
	if meta.Methods.Delete {
		b.WriteString("\tDelete(ctx context.Context, id string) (bool, error)\n")
	}
	if meta.Methods.List {
		b.WriteString(fmt.Sprintf("\tList(ctx context.Context, page, pageSize int) (*ListResult, error)\n"))
	}
	b.WriteString("}\n\n")
	b.WriteString(fmt.Sprintf("// %s 对端快照（应用层 DTO，不依赖 pb）。\n", meta.ViewName))
	b.WriteString(fmt.Sprintf("type %s struct {\n", meta.ViewName))
	b.WriteString(genRPCDemoStructFields(meta.ViewFields, func(f rpcViewField) string { return "" }))
	b.WriteString("}\n\n")
	if meta.Methods.Create {
		b.WriteString("// CreateInput 对端 Create 请求字段（与 proto Create" + meta.Pascal + "Request 一致）。\n")
		b.WriteString("type CreateInput struct {\n")
		b.WriteString(genRPCDemoStructFields(meta.CreateFields, func(f rpcViewField) string { return "" }))
		b.WriteString("}\n\n")
	}
	if meta.Methods.Update {
		b.WriteString("// UpdateInput 对端 Update 请求字段（与 proto Update" + meta.Pascal + "Request 一致）。\n")
		b.WriteString("type UpdateInput struct {\n")
		b.WriteString("\tID string\n")
		b.WriteString(genRPCDemoStructFields(meta.UpdateFields, func(f rpcViewField) string { return "" }))
		b.WriteString("}\n\n")
	}
	if meta.Methods.List {
		b.WriteString("// ListResult 对端 List 响应（与 proto List" + meta.Pascal + "Response 一致）。\n")
		b.WriteString("type ListResult struct {\n")
		b.WriteString(fmt.Sprintf("\tList       []*%s\n", meta.ViewName))
		b.WriteString("\tTotal      int64\n")
		b.WriteString("\tPage       int32\n")
		b.WriteString("\tPageSize   int32\n")
		b.WriteString("\tTotalPages int32\n")
		b.WriteString("}\n")
	}
	return b.String()
}

func filepathBaseDir(gatewayFile string) string {
	parts := strings.Split(strings.ReplaceAll(gatewayFile, "\\", "/"), "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return ""
}

func genSharedRPCGateway(projectModule, remoteService string, meta RPCClientMeta) string {
	var b strings.Builder
	remotePkg := remoteService
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\n", remotePkg))
	b.WriteString(fmt.Sprintf("\t%s \"%s\"\n", meta.PbAlias, meta.ImportPath))
	b.WriteString("\t\"github.com/txbao/goeasy/grpcx\"\n\n")
	b.WriteString(fmt.Sprintf("\trpcport \"%s\"\n)\n\n", meta.SharedPortImport))
	b.WriteString(fmt.Sprintf("// %s 调用 %s 的 %s（ACL）。\n", meta.GatewayName, remotePkg, meta.Pascal+"Service"))
	b.WriteString(fmt.Sprintf("type %s struct {\n\tcli  *grpcx.Client\n\tstub %s.%sServiceClient\n}\n\n", meta.GatewayName, meta.PbAlias, meta.Pascal))
	b.WriteString(fmt.Sprintf("func New%s(cli *grpcx.Client) *%s {\n", meta.GatewayName, meta.GatewayName))
	b.WriteString(fmt.Sprintf("\treturn &%s{\n\t\tcli:  cli,\n\t\tstub: %s.New%sServiceClient(cli.Conn()),\n\t}\n}\n\n", meta.GatewayName, meta.PbAlias, meta.Pascal))

	if meta.Methods.Get {
		b.WriteString(genSharedRPCGatewayGet(meta))
	}
	if meta.Methods.Create {
		b.WriteString(genSharedRPCGatewayCreate(meta))
	}
	if meta.Methods.Update {
		b.WriteString(genSharedRPCGatewayUpdate(meta))
	}
	if meta.Methods.Delete {
		b.WriteString(genSharedRPCGatewayDelete(meta))
	}
	if meta.Methods.List {
		b.WriteString(genSharedRPCGatewayList(meta))
	}
	return b.String()
}

func genSharedRPCGatewayGet(meta RPCClientMeta) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("func (g *%s) GetByID(ctx context.Context, id string) (*rpcport.%s, error) {\n", meta.GatewayName, meta.ViewName))
	b.WriteString(fmt.Sprintf("\tvar out *%s.%s\n", meta.PbAlias, meta.Pascal))
	b.WriteString("\terr := g.cli.Invoke(ctx, func(callCtx context.Context) error {\n\t\tvar callErr error\n")
	b.WriteString(fmt.Sprintf("\t\tout, callErr = g.stub.Get%s(callCtx, &%s.Get%sRequest{Id: id})\n", meta.Pascal, meta.PbAlias, meta.Pascal))
	b.WriteString("\t\treturn callErr\n\t})\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"Get%s id=%%s: %%w\", id, err)\n\t}\n", meta.Pascal))
	b.WriteString(genSharedRPCMapPbToView(meta, "out"))
	b.WriteString("}\n")
	return b.String()
}

func genSharedRPCGatewayCreate(meta RPCClientMeta) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\nfunc (g *%s) Create(ctx context.Context, in *rpcport.CreateInput) (*rpcport.%s, error) {\n", meta.GatewayName, meta.ViewName))
	b.WriteString(fmt.Sprintf("\tvar out *%s.%s\n", meta.PbAlias, meta.Pascal))
	b.WriteString("\terr := g.cli.Invoke(ctx, func(callCtx context.Context) error {\n\t\tvar callErr error\n")
	b.WriteString(fmt.Sprintf("\t\tout, callErr = g.stub.Create%s(callCtx, &%s.Create%sRequest{\n", meta.Pascal, meta.PbAlias, meta.Pascal))
	for _, f := range meta.CreateFields {
		b.WriteString(fmt.Sprintf("\t\t\t%s: in.%s,\n", f.GoName, f.GoName))
	}
	b.WriteString("\t\t})\n\t\treturn callErr\n\t})\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"Create%s: %%w\", err)\n\t}\n", meta.Pascal))
	b.WriteString(genSharedRPCMapPbToView(meta, "out"))
	b.WriteString("}\n")
	return b.String()
}

func genSharedRPCGatewayUpdate(meta RPCClientMeta) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\nfunc (g *%s) Update(ctx context.Context, in *rpcport.UpdateInput) (*rpcport.%s, error) {\n", meta.GatewayName, meta.ViewName))
	b.WriteString(fmt.Sprintf("\tvar out *%s.%s\n", meta.PbAlias, meta.Pascal))
	b.WriteString("\terr := g.cli.Invoke(ctx, func(callCtx context.Context) error {\n\t\tvar callErr error\n")
	b.WriteString(fmt.Sprintf("\t\tout, callErr = g.stub.Update%s(callCtx, &%s.Update%sRequest{\n", meta.Pascal, meta.PbAlias, meta.Pascal))
	b.WriteString("\t\t\tId: in.ID,\n")
	for _, f := range meta.UpdateFields {
		b.WriteString(fmt.Sprintf("\t\t\t%s: in.%s,\n", f.GoName, f.GoName))
	}
	b.WriteString("\t\t})\n\t\treturn callErr\n\t})\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"Update%s: %%w\", err)\n\t}\n", meta.Pascal))
	b.WriteString(genSharedRPCMapPbToView(meta, "out"))
	b.WriteString("}\n")
	return b.String()
}

func genSharedRPCGatewayDelete(meta RPCClientMeta) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\nfunc (g *%s) Delete(ctx context.Context, id string) (bool, error) {\n", meta.GatewayName))
	b.WriteString(fmt.Sprintf("\tvar out *%s.Delete%sResponse\n", meta.PbAlias, meta.Pascal))
	b.WriteString("\terr := g.cli.Invoke(ctx, func(callCtx context.Context) error {\n\t\tvar callErr error\n")
	b.WriteString(fmt.Sprintf("\t\tout, callErr = g.stub.Delete%s(callCtx, &%s.Delete%sRequest{Id: id})\n", meta.Pascal, meta.PbAlias, meta.Pascal))
	b.WriteString("\t\treturn callErr\n\t})\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString(fmt.Sprintf("\t\treturn false, fmt.Errorf(\"Delete%s id=%%s: %%w\", id, err)\n\t}\n", meta.Pascal))
	b.WriteString("\treturn out.GetOk(), nil\n}\n")
	return b.String()
}

func genSharedRPCGatewayList(meta RPCClientMeta) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\nfunc (g *%s) List(ctx context.Context, page, pageSize int) (*rpcport.ListResult, error) {\n", meta.GatewayName))
	b.WriteString(fmt.Sprintf("\tvar out *%s.List%sResponse\n", meta.PbAlias, meta.Pascal))
	b.WriteString("\terr := g.cli.Invoke(ctx, func(callCtx context.Context) error {\n\t\tvar callErr error\n")
	b.WriteString(fmt.Sprintf("\t\tout, callErr = g.stub.List%s(callCtx, &%s.List%sRequest{\n", meta.Pascal, meta.PbAlias, meta.Pascal))
	b.WriteString("\t\t\tPage:     int32(page),\n")
	b.WriteString("\t\t\tPageSize: int32(pageSize),\n")
	b.WriteString("\t\t})\n\t\treturn callErr\n\t})\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"List%s: %%w\", err)\n\t}\n", meta.Pascal))
	b.WriteString("\tlist := make([]*rpcport." + meta.ViewName + ", 0, len(out.GetList()))\n")
	b.WriteString("\tfor _, item := range out.GetList() {\n")
	b.WriteString("\t\tlist = append(list, &rpcport." + meta.ViewName + "{\n")
	for _, f := range meta.ViewFields {
		b.WriteString(fmt.Sprintf("\t\t\t%s: item.%s(),\n", f.GoName, f.PbGetter))
	}
	b.WriteString("\t\t})\n\t}\n")
	b.WriteString("\treturn &rpcport.ListResult{\n")
	b.WriteString("\t\tList:       list,\n")
	b.WriteString("\t\tTotal:      out.GetTotal(),\n")
	b.WriteString("\t\tPage:       out.GetPage(),\n")
	b.WriteString("\t\tPageSize:   out.GetPageSize(),\n")
	b.WriteString("\t\tTotalPages: out.GetTotalPages(),\n")
	b.WriteString("\t}, nil\n}\n")
	return b.String()
}

func genSharedRPCMapPbToView(meta RPCClientMeta, pbVar string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\treturn &rpcport.%s{\n", meta.ViewName))
	for _, f := range meta.ViewFields {
		b.WriteString(fmt.Sprintf("\t\t%s: %s.%s(),\n", f.GoName, pbVar, f.PbGetter))
	}
	b.WriteString("\t}, nil\n")
	return b.String()
}

func genConsumerRPCPortAlias(projectModule, remoteService string, meta RPCClientMeta) string {
	sharedImport := meta.SharedPortImport
	alias := remoteService + "port"
	var b strings.Builder
	b.WriteString("package port\n\nimport (\n")
	b.WriteString(fmt.Sprintf("\t%s \"%s\"\n", alias, sharedImport))
	b.WriteString(")\n\n")
	b.WriteString(fmt.Sprintf("// 业务侧 Port：type alias 到共享 %s/%s/port（由 add rpc bind 生成）。\n", remoteService, meta.Module))
	b.WriteString("type (\n")
	b.WriteString(fmt.Sprintf("\t%s = %s.%s\n", meta.GatewayName, alias, meta.GatewayName))
	b.WriteString(fmt.Sprintf("\t%s = %s.%s\n", meta.ViewName, alias, meta.ViewName))
	b.WriteString(fmt.Sprintf("\tCreateInput = %s.CreateInput\n", alias))
	b.WriteString(fmt.Sprintf("\tUpdateInput = %s.UpdateInput\n", alias))
	b.WriteString(fmt.Sprintf("\tListResult = %s.ListResult\n", alias))
	b.WriteString(")\n")
	_ = projectModule
	return b.String()
}
