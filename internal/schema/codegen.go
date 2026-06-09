package schema

import "strings"

// ColsNeedTimeImport 判断列集合生成的 Go 类型是否需要 import "time"。
func ColsNeedTimeImport(cols []ColumnMeta) bool {
	for _, c := range cols {
		if strings.Contains(GoFieldFromColumn(c).GoType, "time.") {
			return true
		}
	}
	return false
}

// UniqueColumnsByName 按列名去重合并多组列（保留首次出现顺序）。
func UniqueColumnsByName(groups ...[]ColumnMeta) []ColumnMeta {
	seen := make(map[string]bool)
	var out []ColumnMeta
	for _, cols := range groups {
		for _, c := range cols {
			key := strings.ToLower(c.Name)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, c)
		}
	}
	return out
}
