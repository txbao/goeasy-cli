package generator

import (
	"github.com/txbao/goeasy-cli/internal/schema"
)

// GRPCCol 供 gRPC 模板与 convert 生成使用的列元数据。
type GRPCCol struct {
	Name       string
	Pascal     string
	ProtoField string // pb 结构体字段名（protoc-gen-go，如 Id）
	GoType     string
	ProtoType  string
}

func grpcColsFrom(columns []schema.ColumnMeta) []GRPCCol {
	out := make([]GRPCCol, 0, len(columns))
	for _, c := range columns {
		gf := schema.GoFieldFromColumn(c)
		out = append(out, GRPCCol{
			Name:       c.Name,
			Pascal:     gf.Name,
			ProtoField: protoStructField(gf.Name),
			GoType:     gf.GoType,
			ProtoType:  schema.ProtoType(c),
		})
	}
	return out
}

// protoStructField pb 消息字段名（与 protoc-gen-go 一致，id → Id）。
func protoStructField(goPascal string) string {
	if goPascal == "ID" {
		return "Id"
	}
	return goPascal
}

func enrichGRPCData(data map[string]any, ct schema.ClassifiedTable, projectDir string, withCRUD bool, meta ModuleMeta) {
	data["ReadCols"] = grpcColsFrom(ct.ReadCols)
	data["CreateCols"] = grpcColsFrom(ct.CreateCols)
	data["UpdateCols"] = grpcColsFrom(ct.UpdateCols)
	data["WithPG"] = repositoryPGExists(projectDir, meta) || withCRUD
}
