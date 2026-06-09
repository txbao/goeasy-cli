package generator

import (
	"fmt"
	"strings"
)

// genGRPCHandlers 按 app_style 生成 gRPC handlers（与 HTTP handler codegen 一致）。
func genGRPCHandlers(projectModule, goeasy, pascal, alias string, meta ModuleMeta, snake string, style AppStyle, readCols, createCols, updateCols []GRPCCol) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"context\"\n\t\"errors\"\n\n", meta.Resource))
	b.WriteString("\t\"google.golang.org/grpc/codes\"\n")
	b.WriteString("\t\"google.golang.org/grpc/status\"\n\n")
	b.WriteString(fmt.Sprintf("\t%sapp \"%s\"\n", alias, meta.AppImportPath(projectModule)))
	if style.IsLightCQRS() {
		b.WriteString(fmt.Sprintf("\t\"%s/command\"\n", meta.AppImportPath(projectModule)))
	}
	b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n", meta.DomainImportPath(projectModule)))
	b.WriteString(fmt.Sprintf("\tpb \"%s/api/proto/gen/%s\"\n", projectModule, snake))
	b.WriteString(fmt.Sprintf("\tzpage \"%s/pagination\"\n)\n\n", goeasy))

	b.WriteString(fmt.Sprintf("func (s *Server) List%s(ctx context.Context, req *pb.List%sRequest) (*pb.List%sResponse, error) {\n", pascal, pascal, pascal))
	b.WriteString("\tnorm := zpage.Normalize(zpage.Page{Page: int(req.GetPage()), PageSize: int(req.GetPageSize())})\n")
	b.WriteString("\tpage, pageSize := norm.Page, norm.PageSize\n")
	b.WriteString("\tresult, err := s.app.List(ctx, page, pageSize)\n")
	b.WriteString("\tif err != nil {\n\t\treturn nil, status.Errorf(codes.Internal, \"%%v\", err)\n\t}\n")
	b.WriteString(fmt.Sprintf("\tlist := make([]*pb.%s, 0, len(result.List))\n", pascal))
	b.WriteString("\tfor i := range result.List {\n\t\tlist = append(list, dtoToProto(&result.List[i]))\n\t}\n")
	b.WriteString(fmt.Sprintf("\treturn &pb.List%sResponse{\n", pascal))
	b.WriteString("\t\tList:       list,\n\t\tTotal:      result.Total,\n\t\tPage:       int32(page),\n\t\tPageSize:   int32(pageSize),\n")
	b.WriteString("\t\tTotalPages: int32(zpage.TotalPages(result.Total, pageSize)),\n\t}, nil\n}\n\n")

	b.WriteString(fmt.Sprintf("func (s *Server) Get%s(ctx context.Context, req *pb.Get%sRequest) (*pb.%s, error) {\n", pascal, pascal, pascal))
	if style.IsService() {
		b.WriteString("\tagg, err := s.app.Get(ctx, req.GetId())\n")
	} else {
		b.WriteString("\tagg, err := s.app.Queries().Get(ctx, req.GetId())\n")
	}
	b.WriteString("\tif err != nil {\n\t\treturn nil, grpcError(err)\n\t}\n\treturn aggregateToProto(agg), nil\n}\n\n")

	b.WriteString(fmt.Sprintf("func (s *Server) Create%s(ctx context.Context, req *pb.Create%sRequest) (*pb.%s, error) {\n", pascal, pascal, pascal))
	b.WriteString("\tcmd := createRequestToCommand(req)\n")
	if style.IsService() {
		b.WriteString("\tid, err := s.app.Create(ctx, cmd)\n")
	} else {
		b.WriteString("\tid, err := s.app.Commands().Create(ctx, cmd)\n")
	}
	b.WriteString("\tif err != nil {\n\t\treturn nil, grpcError(err)\n\t}\n")
	if style.IsService() {
		b.WriteString("\tagg, err := s.app.Get(ctx, id)\n")
	} else {
		b.WriteString("\tagg, err := s.app.Queries().Get(ctx, id)\n")
	}
	b.WriteString("\tif err != nil {\n\t\treturn nil, grpcError(err)\n\t}\n\treturn aggregateToProto(agg), nil\n}\n\n")

	b.WriteString(fmt.Sprintf("func (s *Server) Update%s(ctx context.Context, req *pb.Update%sRequest) (*pb.%s, error) {\n", pascal, pascal, pascal))
	b.WriteString("\tcmd := updateRequestToCommand(req)\n")
	if style.IsService() {
		b.WriteString("\tif err := s.app.Update(ctx, cmd); err != nil {\n")
	} else {
		b.WriteString("\tif err := s.app.Commands().Update(ctx, cmd); err != nil {\n")
	}
	b.WriteString("\t\treturn nil, grpcError(err)\n\t}\n")
	if style.IsService() {
		b.WriteString("\tagg, err := s.app.Get(ctx, req.GetId())\n")
	} else {
		b.WriteString("\tagg, err := s.app.Queries().Get(ctx, req.GetId())\n")
	}
	b.WriteString("\tif err != nil {\n\t\treturn nil, grpcError(err)\n\t}\n\treturn aggregateToProto(agg), nil\n}\n\n")

	b.WriteString(fmt.Sprintf("func (s *Server) Delete%s(ctx context.Context, req *pb.Delete%sRequest) (*pb.Delete%sResponse, error) {\n", pascal, pascal, pascal))
	if style.IsService() {
		b.WriteString("\tif err := s.app.Delete(ctx, req.GetId()); err != nil {\n")
	} else {
		b.WriteString("\tif err := s.app.Commands().Delete(ctx, req.GetId()); err != nil {\n")
	}
	b.WriteString("\t\treturn nil, grpcError(err)\n\t}\n")
	b.WriteString(fmt.Sprintf("\treturn &pb.Delete%sResponse{Ok: true}, nil\n}\n\n", pascal))

	b.WriteString(`func grpcError(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Errorf(codes.Internal, "%v", err)
}

`)
	b.WriteString(fmt.Sprintf("func dtoToProto(d *%sapp.%sDTO) *pb.%s {\n", alias, pascal, pascal))
	b.WriteString(fmt.Sprintf("\treturn &pb.%s{\n", pascal))
	for _, c := range readCols {
		b.WriteString(fmt.Sprintf("\t\t%s: dtoField%s(d),\n", c.ProtoField, c.Pascal))
	}
	b.WriteString("\t}\n}\n\n")

	b.WriteString(fmt.Sprintf("func aggregateToProto(agg *domain.Aggregate) *pb.%s {\n", pascal))
	b.WriteString(fmt.Sprintf("\td := %sapp.ToDTO(agg)\n", alias))
	b.WriteString("\treturn dtoToProto(&d)\n}\n\n")

	if style.IsService() {
		b.WriteString(fmt.Sprintf("func createRequestToCommand(req *pb.Create%sRequest) %sapp.CreateCommand {\n", pascal, alias))
		b.WriteString(fmt.Sprintf("\treturn %sapp.CreateCommand{\n", alias))
	} else {
		b.WriteString(fmt.Sprintf("func createRequestToCommand(req *pb.Create%sRequest) command.CreateCommand {\n", pascal))
		b.WriteString("\treturn command.CreateCommand{\n")
	}
	for _, c := range createCols {
		b.WriteString(fmt.Sprintf("\t\t%s: protoCreate%s(req),\n", c.Pascal, c.Pascal))
	}
	b.WriteString("\t}\n}\n\n")

	if style.IsService() {
		b.WriteString(fmt.Sprintf("func updateRequestToCommand(req *pb.Update%sRequest) %sapp.UpdateCommand {\n", pascal, alias))
		b.WriteString(fmt.Sprintf("\treturn %sapp.UpdateCommand{\n", alias))
	} else {
		b.WriteString(fmt.Sprintf("func updateRequestToCommand(req *pb.Update%sRequest) command.UpdateCommand {\n", pascal))
		b.WriteString("\treturn command.UpdateCommand{\n")
	}
	b.WriteString("\t\tID: req.GetId(),\n")
	for _, c := range updateCols {
		b.WriteString(fmt.Sprintf("\t\t%s: protoUpdate%s(req),\n", c.Pascal, c.Pascal))
	}
	b.WriteString("\t}\n}\n")
	return b.String()
}
