package schema

import "strings"

// GoField 生成 Go 结构体字段信息。
type GoField struct {
	Name    string // Pascal
	DBName  string
	GoType  string
	Tag     string
}

func GoFieldFromColumn(c ColumnMeta) GoField {
	return GoField{
		Name:   toPascalCol(c.Name),
		DBName: c.Name,
		GoType: goTypeForDB(c.DBType, c.Nullable),
		Tag:    "`db:\"" + c.Name + "\"`",
	}
}

func goTypeForDB(dbType string, nullable bool) string {
	t := strings.ToLower(dbType)
	var base string
	switch {
	case strings.Contains(t, "bigint"), strings.Contains(t, "int8"):
		base = "int64"
	case strings.Contains(t, "smallint"), strings.Contains(t, "int2"):
		base = "int16"
	case strings.Contains(t, "integer"), strings.Contains(t, "int4"), strings.Contains(t, "serial"):
		base = "int32"
	case strings.Contains(t, "bool"):
		base = "bool"
	case strings.Contains(t, "timestamp"), strings.Contains(t, "date"):
		base = "time.Time"
	case strings.Contains(t, "float"), strings.Contains(t, "double"):
		base = "float64"
	case strings.Contains(t, "numeric"), strings.Contains(t, "decimal"):
		base = "string"
	default:
		base = "string"
	}
	if nullable && base != "string" && !strings.HasPrefix(base, "[]") {
		return "*" + base
	}
	return base
}

func toPascalCol(name string) string {
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		if strings.EqualFold(p, "id") {
			parts[i] = "ID"
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, "")
}

// ProtoType 映射 proto3 类型。
func ProtoType(c ColumnMeta) string {
	t := strings.ToLower(c.DBType)
	switch {
	case strings.Contains(t, "bigint"), strings.Contains(t, "int8"):
		return "int64"
	case strings.Contains(t, "smallint"), strings.Contains(t, "int2"):
		return "int32"
	case strings.Contains(t, "integer"), strings.Contains(t, "int4"):
		return "int32"
	case strings.Contains(t, "bool"):
		return "bool"
	case strings.Contains(t, "timestamp"), strings.Contains(t, "date"):
		return "string" // V1 用 string，V2 google.protobuf.Timestamp
	default:
		return "string"
	}
}
