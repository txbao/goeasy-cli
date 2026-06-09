package schema

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var varcharLenRE = regexp.MustCompile(`varchar\s*\(\s*(\d+)\s*\)`)

// FieldTags 生成 json + validate 结构体标签（用于 Command 等写模型）。
// mode: "create" | "update"
func FieldTags(c ColumnMeta, mode string) string {
	jsonTag := fmt.Sprintf(`json:"%s"`, c.Name)
	v := validateRules(c, mode)
	if v == "" {
		return jsonTag
	}
	return jsonTag + ` validate:"` + v + `"`
}

func validateRules(c ColumnMeta, mode string) string {
	var parts []string
	if mode == "create" {
		if !c.Nullable {
			parts = append(parts, "required")
		} else {
			parts = append(parts, "omitempty")
		}
	} else {
		parts = append(parts, "omitempty")
	}
	maxLen := varcharMaxLen(c.DBType)
	if maxLen > 0 {
		parts = append(parts, fmt.Sprintf("max=%d", maxLen))
	}
	t := strings.ToLower(c.DBType)
	switch {
	case strings.Contains(t, "int"), strings.Contains(t, "serial"):
		// 业务整型默认允许 0
	case strings.Contains(t, "bool"):
	default:
		if strings.Contains(t, "char") || strings.Contains(t, "text") {
			if maxLen == 0 {
				parts = append(parts, "max=1024")
			}
		}
	}
	return strings.Join(parts, ",")
}

func varcharMaxLen(dbType string) int {
	t := strings.ToLower(dbType)
	if m := varcharLenRE.FindStringSubmatch(t); len(m) == 2 {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	if strings.Contains(t, "character varying") {
		return 255
	}
	return 0
}
