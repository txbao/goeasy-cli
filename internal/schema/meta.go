package schema

// TableMeta 数据库表元数据。
type TableMeta struct {
	Schema       string
	Name         string
	TableComment string // 表备注（PG/MySQL）
	Columns      []ColumnMeta
}

// ColumnMeta 列元数据。
type ColumnMeta struct {
	Name         string
	DBType       string
	Nullable     bool
	IsPrimaryKey bool
	Ordinal      int
	Comment      string // 列备注（PG col_description / MySQL COLUMN_COMMENT）
}

// ClassifiedTable 带列角色的表（用于代码生成）。
type ClassifiedTable struct {
	TableMeta
	ModuleName      string // 逻辑模块名（不含 table_prefix）
	PhysicalName    string // 实际表名
	PKColumn        string
	SoftDeleteCol   string
	CreateOmit      []string
	UpdateOmit      []string
	BusinessCols    []ColumnMeta
	CreateCols      []ColumnMeta
	UpdateCols      []ColumnMeta
	ReadCols        []ColumnMeta
}
