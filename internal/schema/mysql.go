package schema

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// LoadMySQLTable 读取 MySQL 单表结构（schema 参数为 database 名）。
func LoadMySQLTable(dsn, database, table string) (TableMeta, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return TableMeta{}, err
	}
	defer db.Close()

	if database == "" {
		database = MySQLDatabaseFromDSN(dsn)
	}
	if database == "" {
		return TableMeta{}, fmt.Errorf("mysql database name required (set --schema or DSN path /dbname)")
	}
	cols, tableComment, err := loadMySQLColumns(db, database, table)
	if err != nil {
		return TableMeta{}, err
	}
	if len(cols) == 0 {
		return TableMeta{}, fmt.Errorf("table %s.%s not found or has no columns", database, table)
	}
	return TableMeta{Schema: database, Name: table, TableComment: tableComment, Columns: cols}, nil
}

// ListMySQLTables 列出 database 下所有基表。
func ListMySQLTables(dsn, database string) ([]string, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if database == "" {
		database = MySQLDatabaseFromDSN(dsn)
	}
	if database == "" {
		return nil, fmt.Errorf("mysql database name required")
	}
	const q = `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = ? AND table_type = 'BASE TABLE'
ORDER BY table_name`
	rows, err := db.Query(q, database)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

// MySQLDatabaseFromDSN 从 DSN 解析库名（user:pass@tcp(host:3306)/dbname?...）。
func MySQLDatabaseFromDSN(dsn string) string {
	i := strings.LastIndex(dsn, "/")
	if i < 0 {
		return ""
	}
	rest := dsn[i+1:]
	if j := strings.Index(rest, "?"); j >= 0 {
		rest = rest[:j]
	}
	return strings.TrimSpace(rest)
}

func loadMySQLColumns(db *sql.DB, database, table string) ([]ColumnMeta, string, error) {
	const q = `
SELECT c.column_name, c.data_type, c.is_nullable,
       EXISTS (
         SELECT 1 FROM information_schema.table_constraints tc
         JOIN information_schema.key_column_usage kcu
           ON tc.constraint_name = kcu.constraint_name
          AND tc.table_schema = kcu.table_schema
         WHERE tc.constraint_type = 'PRIMARY KEY'
           AND tc.table_schema = c.table_schema
           AND tc.table_name = c.table_name
           AND kcu.column_name = c.column_name
       ) AS is_pk,
       c.ordinal_position,
       c.column_comment,
       t.table_comment
FROM information_schema.columns c
JOIN information_schema.tables t
  ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE c.table_schema = ? AND c.table_name = ?
ORDER BY c.ordinal_position`
	rows, err := db.Query(q, database, table)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var cols []ColumnMeta
	var tableComment string
	for rows.Next() {
		var name, dataType, nullable, colComment, tblComment string
		var isPK bool
		var ord int
		if err := rows.Scan(&name, &dataType, &nullable, &isPK, &ord, &colComment, &tblComment); err != nil {
			return nil, "", err
		}
		if tableComment == "" {
			tableComment = strings.TrimSpace(tblComment)
		}
		cols = append(cols, ColumnMeta{
			Name:         name,
			DBType:       dataType,
			Nullable:     strings.EqualFold(nullable, "YES"),
			IsPrimaryKey: isPK,
			Ordinal:      ord,
			Comment:      strings.TrimSpace(colComment),
		})
	}
	return cols, tableComment, rows.Err()
}
