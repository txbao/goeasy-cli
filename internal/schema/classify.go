package schema

import "strings"

// 内置扫表排除（不可关闭）。
var AlwaysExcludeTables = []string{"_sqlx_migrations"}

var defaultCreateOmit = []string{"id", "created_at", "updated_at", "deleted_at"}
var defaultUpdateOmit = []string{"created_at", "updated_at", "deleted_at"}

// CodegenRules 列分类与写请求裁剪规则。
type CodegenRules struct {
	CreateOmit             []string
	UpdateOmit             []string
	SoftDeleteColumn       string
	TouchCreatedAtOnInsert bool
	TouchUpdatedAtOnUpdate bool
}

func DefaultCodegenRules() CodegenRules {
	return CodegenRules{
		CreateOmit:             append([]string(nil), defaultCreateOmit...),
		UpdateOmit:             append([]string(nil), defaultUpdateOmit...),
		SoftDeleteColumn:       "deleted_at",
		TouchCreatedAtOnInsert: true,
		TouchUpdatedAtOnUpdate: true,
	}
}

func (r CodegenRules) createOmitSet() map[string]bool {
	return toSet(r.CreateOmit)
}

func (r CodegenRules) updateOmitSet() map[string]bool {
	return toSet(r.UpdateOmit)
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[strings.ToLower(s)] = true
	}
	return m
}

func containsCI(set map[string]bool, name string) bool {
	return set[strings.ToLower(name)]
}

// Classify 根据规则标注列并拆分 Create/Update 业务列。
func Classify(t TableMeta, moduleName, physicalName string, rules CodegenRules) ClassifiedTable {
	ct := ClassifiedTable{
		TableMeta:    t,
		ModuleName:   moduleName,
		PhysicalName: physicalName,
		CreateOmit:   rules.CreateOmit,
		UpdateOmit:   rules.UpdateOmit,
		SoftDeleteCol: rules.SoftDeleteColumn,
		ReadCols:     append([]ColumnMeta(nil), t.Columns...),
	}
	createOmit := rules.createOmitSet()
	updateOmit := rules.updateOmitSet()

	for _, c := range t.Columns {
		if c.IsPrimaryKey && ct.PKColumn == "" {
			ct.PKColumn = c.Name
		}
		if strings.EqualFold(c.Name, rules.SoftDeleteColumn) {
			ct.SoftDeleteCol = c.Name
		}
		if containsCI(createOmit, c.Name) {
			continue
		}
		ct.CreateCols = append(ct.CreateCols, c)
	}
	for _, c := range t.Columns {
		if containsCI(updateOmit, c.Name) || containsCI(createOmit, c.Name) && strings.EqualFold(c.Name, "id") {
			continue
		}
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		ct.UpdateCols = append(ct.UpdateCols, c)
	}
	// 业务列 = 可读列中非纯系统列（用于 entity 字段）
	sys := toSet(append(rules.CreateOmit, "version"))
	for _, c := range t.Columns {
		if containsCI(sys, c.Name) && !c.IsPrimaryKey {
			continue
		}
		ct.BusinessCols = append(ct.BusinessCols, c)
	}
	return ct
}

// ShouldExcludeTable 判断是否排除该表（内置 + 用户 exclude）。
func ShouldExcludeTable(table string, userExclude []string) bool {
	for _, x := range AlwaysExcludeTables {
		if strings.EqualFold(table, x) {
			return true
		}
	}
	for _, pattern := range userExclude {
		if matchTablePattern(table, pattern) {
			return true
		}
	}
	return false
}

func matchTablePattern(table, pattern string) bool {
	if pattern == "" {
		return false
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(strings.ToLower(table), strings.ToLower(strings.TrimSuffix(pattern, "*")))
	}
	return strings.EqualFold(table, pattern)
}

// ModuleNameFromPhysical 从物理表名解析模块名（去掉 table_prefix）。
func ModuleNameFromPhysical(physical, prefix string) string {
	if prefix != "" && strings.HasPrefix(physical, prefix) {
		return strings.TrimPrefix(physical, prefix)
	}
	return physical
}
