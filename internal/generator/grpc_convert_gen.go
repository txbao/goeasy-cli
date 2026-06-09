package generator

import (
	"fmt"
	"strings"
)

func genGRPCConvertGo(meta ModuleMeta, projectModule, moduleAlias, modulePascal, pbImport string, readCols, createCols, updateCols []GRPCCol) string {
	dtoType := moduleAlias + "app." + modulePascal + "DTO"
	createReq := "pb.Create" + modulePascal + "Request"
	updateReq := "pb.Update" + modulePascal + "Request"
	appImport := meta.AppImportPath(projectModule)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\n", meta.Resource))
	b.WriteString("import (\n\t\"time\"\n\n")
	b.WriteString(fmt.Sprintf("\t%sapp \"%s\"\n", moduleAlias, appImport))
	b.WriteString(fmt.Sprintf("\tpb \"%s\"\n)\n\n", pbImport))

	for _, c := range readCols {
		b.WriteString(genDTOFieldFn("dtoField"+c.Pascal, dtoType, c))
		b.WriteString("\n\n")
	}
	for _, c := range createCols {
		b.WriteString(genProtoCreateFn("protoCreate"+c.Pascal, createReq, c))
		b.WriteString("\n\n")
	}
	for _, c := range updateCols {
		b.WriteString(genProtoUpdateFn("protoUpdate"+c.Pascal, updateReq, c))
		b.WriteString("\n\n")
	}
	return b.String()
}

func genDTOFieldFn(fn, dtoType string, c GRPCCol) string {
	df := "d." + c.Pascal
	switch c.GoType {
	case "time.Time":
		return fmt.Sprintf("func %s(d *%s) string {\n\tif d == nil {\n\t\treturn \"\"\n\t}\n\treturn %s.Format(time.RFC3339)\n}", fn, dtoType, df)
	case "*time.Time":
		return fmt.Sprintf("func %s(d *%s) string {\n\tif d == nil || %s == nil {\n\t\treturn \"\"\n\t}\n\treturn %s.Format(time.RFC3339)\n}", fn, dtoType, df, df)
	case "*int16":
		return fmt.Sprintf("func %s(d *%s) int32 {\n\tif d == nil || %s == nil {\n\t\treturn 0\n\t}\n\treturn int32(*%s)\n}", fn, dtoType, df, df)
	case "int16":
		return fmt.Sprintf("func %s(d *%s) int32 {\n\tif d == nil {\n\t\treturn 0\n\t}\n\treturn int32(%s)\n}", fn, dtoType, df)
	case "int32":
		return fmt.Sprintf("func %s(d *%s) int32 {\n\tif d == nil {\n\t\treturn 0\n\t}\n\treturn %s\n}", fn, dtoType, df)
	case "int64":
		return fmt.Sprintf("func %s(d *%s) int64 {\n\tif d == nil {\n\t\treturn 0\n\t}\n\treturn %s\n}", fn, dtoType, df)
	case "bool":
		return fmt.Sprintf("func %s(d *%s) bool {\n\tif d == nil {\n\t\treturn false\n\t}\n\treturn %s\n}", fn, dtoType, df)
	default:
		return fmt.Sprintf("func %s(d *%s) string {\n\tif d == nil {\n\t\treturn \"\"\n\t}\n\treturn %s\n}", fn, dtoType, df)
	}
}

func genProtoCreateFn(fn, reqType string, c GRPCCol) string {
	getter := "req.Get" + c.ProtoField + "()"
	switch c.GoType {
	case "*int16":
		return fmt.Sprintf("func %s(req *%s) *int16 {\n\tv := int16(%s)\n\treturn &v\n}", fn, reqType, getter)
	case "int16":
		return fmt.Sprintf("func %s(req *%s) int16 {\n\treturn int16(%s)\n}", fn, reqType, getter)
	case "int32":
		return fmt.Sprintf("func %s(req *%s) int32 {\n\treturn %s\n}", fn, reqType, getter)
	case "int64":
		return fmt.Sprintf("func %s(req *%s) int64 {\n\treturn %s\n}", fn, reqType, getter)
	case "bool":
		return fmt.Sprintf("func %s(req *%s) bool {\n\treturn %s\n}", fn, reqType, getter)
	default:
		return fmt.Sprintf("func %s(req *%s) string {\n\treturn %s\n}", fn, reqType, getter)
	}
}

func genProtoUpdateFn(fn, reqType string, c GRPCCol) string {
	return genProtoCreateFn(fn, reqType, c)
}
