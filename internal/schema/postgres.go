package schema

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// LoadPostgresTable 读取单表结构。
func LoadPostgresTable(dsn, schema, table string) (TableMeta, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return TableMeta{}, err
	}
	defer db.Close()

	if schema == "" {
		schema = "public"
	}
	cols, tableComment, err := loadPostgresColumns(db, schema, table)
	if err != nil {
		return TableMeta{}, err
	}
	if len(cols) == 0 {
		return TableMeta{}, fmt.Errorf("table %s.%s not found or has no columns", schema, table)
	}
	return TableMeta{Schema: schema, Name: table, TableComment: tableComment, Columns: cols}, nil
}

// ListPostgresTables 列出 schema 下所有基表。
func ListPostgresTables(dsn, schema string) ([]string, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if schema == "" {
		schema = "public"
	}
	const q = `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = $1 AND table_type = 'BASE TABLE'
ORDER BY table_name`
	rows, err := db.Query(q, schema)
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

func loadPostgresColumns(db *sql.DB, schema, table string) ([]ColumnMeta, string, error) {
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
       pg_catalog.col_description(cl.oid, c.ordinal_position::int) AS column_comment,
       obj_description(cl.oid) AS table_comment
FROM information_schema.columns c
JOIN pg_catalog.pg_class cl ON cl.relname = c.table_name
JOIN pg_catalog.pg_namespace n ON n.oid = cl.relnamespace AND n.nspname = c.table_schema
WHERE c.table_schema = $1 AND c.table_name = $2
ORDER BY c.ordinal_position`
	rows, err := db.Query(q, schema, table)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var cols []ColumnMeta
	var tableComment string
	for rows.Next() {
		var name, dataType, nullable string
		var colComment sql.NullString
		var tblComment sql.NullString
		var isPK bool
		var ord int
		if err := rows.Scan(&name, &dataType, &nullable, &isPK, &ord, &colComment, &tblComment); err != nil {
			return nil, "", err
		}
		if tblComment.Valid && tableComment == "" {
			tableComment = tblComment.String
		}
		cols = append(cols, ColumnMeta{
			Name:         name,
			DBType:       dataType,
			Nullable:     strings.EqualFold(nullable, "YES"),
			IsPrimaryKey: isPK,
			Ordinal:      ord,
			Comment:      nullString(colComment),
		})
	}
	return cols, tableComment, rows.Err()
}

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return strings.TrimSpace(ns.String)
	}
	return ""
}
