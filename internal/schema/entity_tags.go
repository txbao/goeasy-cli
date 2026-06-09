package schema

import (
	"fmt"
	"strings"
)

// EntityJSONTag 领域实体 json 标签（列名与数据库一致，可空字段带 omitempty）。
func EntityJSONTag(c ColumnMeta, goType string) string {
	if c.Nullable || strings.HasPrefix(goType, "*") {
		return fmt.Sprintf(`json:"%s,omitempty"`, c.Name)
	}
	return fmt.Sprintf(`json:"%s"`, c.Name)
}
