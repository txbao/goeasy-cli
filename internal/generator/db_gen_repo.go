package generator

import (
	"fmt"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func genRepositoryIface(pkg string) string {
	return fmt.Sprintf(`package %s

import "context"

type Repository interface {
	FindByID(ctx context.Context, id string) (*Aggregate, error)
	List(ctx context.Context, page, pageSize int) ([]*Aggregate, int64, error)
	Create(ctx context.Context, agg *Aggregate) (string, error)
	Update(ctx context.Context, agg *Aggregate) error
	Delete(ctx context.Context, id string) error
}
`, pkg)
}

func genRepositoryPG(projectModule, goeasy string, ct schema.ClassifiedTable, pascal, alias string, meta ModuleMeta) string {
	pkg := meta.Resource
	snake := meta.ModuleID
	table := ct.PhysicalName
	pk := ct.PKColumn
	if pk == "" {
		pk = "id"
	}
	soft := ct.SoftDeleteCol
	pkField := schema.GoFieldFromColumn(findCol(ct.ReadCols, pk))

	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"context\"\n\t\"encoding/json\"\n\t\"errors\"\n\t\"fmt\"\n\t\"time\"\n", pkg))
	b.WriteString("\n\t\"github.com/doug-martin/goqu/v9\"\n\t\"github.com/jmoiron/sqlx\"\n\n")
	b.WriteString(fmt.Sprintf("\tdomain \"%s\"\n", meta.DomainImportPath(projectModule)))
	b.WriteString(fmt.Sprintf("\t\"%s/internal/infrastructure/shared/dbx\"\n", projectModule))
	b.WriteString(fmt.Sprintf("\tzcache \"%s/cache\"\n", goeasy))
	b.WriteString(fmt.Sprintf("\tzcachekey \"%s/cachekey\"\n)\n\n", goeasy))

	b.WriteString("type PGRepository struct {\n\tdbx       *dbx.DB\n\ttable     string\n\tmodule    string\n\tcache     zcache.Cache\n\tkeyPrefix string\n\tcacheOn   bool\n\tcacheTTL  time.Duration\n}\n\n")
	b.WriteString("func NewPGRepository(sqlxDB *sqlx.DB, driver, table string, c zcache.Cache, keyPrefix, module string, cacheOn bool, cacheTTL time.Duration) domain.Repository {\n")
	b.WriteString("\tif sqlxDB == nil {\n\t\tpanic(\"nil sqlx.DB\")\n\t}\n")
	b.WriteString("\tif table == \"\" {\n")
	b.WriteString(fmt.Sprintf("\t\ttable = %q\n", table))
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\tif module == \"\" {\n\t\tmodule = %q\n\t}\n", snake))
	b.WriteString("\treturn &PGRepository{\n\t\tdbx: dbx.New(sqlxDB, driver), table: table,\n\t\tmodule: module, cache: c, keyPrefix: keyPrefix, cacheOn: cacheOn, cacheTTL: cacheTTL,\n\t}\n}\n\n")

	b.WriteString(fmt.Sprintf("func (r *PGRepository) invalidateEntityCache(ctx context.Context, id string) {\n"))
	b.WriteString("\tif !r.cacheOn || r.cache == nil {\n\t\treturn\n\t}\n")
	b.WriteString("\t_ = zcachekey.DeleteEntity(ctx, r.cache, r.keyPrefix, r.module, id)\n}\n\n")

	b.WriteString(fmt.Sprintf("type %sRow struct {\n", pascal))
	for _, c := range ct.ReadCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t%s %s %s\n", f.Name, f.GoType, f.Tag))
	}
	b.WriteString("}\n\n")

	b.WriteString(fmt.Sprintf("func rowToAgg(row *%sRow) *domain.Aggregate {\n", pascal))
	b.WriteString(fmt.Sprintf("\treturn domain.NewAggregate(domain.Rehydrate%s(", pascal))
	for i, c := range ct.ReadCols {
		f := schema.GoFieldFromColumn(c)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("row.%s", f.Name))
	}
	b.WriteString("))\n}\n\n")

	b.WriteString(fmt.Sprintf("func pkString(row *%sRow) string {\n", pascal))
	b.WriteString(fmt.Sprintf("\treturn fmt.Sprint(row.%s)\n}\n\n", pkField.Name))

	// FindByID
	b.WriteString("func (r *PGRepository) FindByID(ctx context.Context, id string) (*domain.Aggregate, error) {\n")
	b.WriteString(fmt.Sprintf("\tvar row %sRow\n", pascal))
	b.WriteString("\tif r.cacheOn && r.cache != nil {\n")
	b.WriteString("\t\tif raw, err := zcachekey.GetEntityBytes(ctx, r.cache, r.keyPrefix, r.module, id); err == nil {\n")
	b.WriteString("\t\t\tif json.Unmarshal(raw, &row) == nil {\n")
	b.WriteString("\t\t\t\treturn rowToAgg(&row), nil\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t} else if !errors.Is(err, zcache.ErrNotFound) {\n")
	b.WriteString("\t\t\treturn nil, err\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\tds := r.dbx.Qu.From(r.table).Select(" + goquSelectCols(ct.ReadCols) + ")")
	b.WriteString(fmt.Sprintf(".Where(goqu.Ex{%q: id})", pk))
	if soft != "" {
		b.WriteString(fmt.Sprintf(".Where(goqu.C(%q).IsNull())", soft))
	}
	b.WriteString("\n\tsql, args, err := ds.ToSQL()\n\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\tif err := dbx.GetContext(ctx, r.dbx, &row, sql, args...); err != nil {\n")
	b.WriteString("\t\tif dbx.IsNotFound(err) {\n\t\t\treturn nil, domain.ErrNotFound\n\t\t}\n\t\treturn nil, err\n\t}\n")
	b.WriteString("\tif r.cacheOn && r.cache != nil {\n")
	b.WriteString("\t\tif raw, err := json.Marshal(&row); err == nil {\n")
	b.WriteString("\t\t\t_ = zcachekey.SetEntityBytes(ctx, r.cache, r.keyPrefix, r.module, id, raw, r.cacheTTL)\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn rowToAgg(&row), nil\n}\n\n")

	// List
	b.WriteString("func (r *PGRepository) List(ctx context.Context, page, pageSize int) ([]*domain.Aggregate, int64, error) {\n")
	b.WriteString("\tif page < 1 {\n\t\tpage = 1\n\t}\n")
	b.WriteString("\tif pageSize < 1 {\n\t\tpageSize = 20\n\t}\n")
	b.WriteString("\toffset := (page - 1) * pageSize\n")
	b.WriteString("\tcountDS := r.dbx.Qu.From(r.table).Select(goqu.COUNT(\"*\"))\n")
	if soft != "" {
		b.WriteString(fmt.Sprintf("\tcountDS = countDS.Where(goqu.C(%q).IsNull())\n", soft))
	}
	b.WriteString("\tcountSQL, countArgs, err := countDS.ToSQL()\n")
	b.WriteString("\tif err != nil {\n\t\treturn nil, 0, err\n\t}\n")
	b.WriteString("\tvar total int64\n")
	b.WriteString("\tif err := dbx.GetContext(ctx, r.dbx, &total, countSQL, countArgs...); err != nil {\n\t\treturn nil, 0, err\n\t}\n")
	b.WriteString("\tif total == 0 {\n\t\treturn []*domain.Aggregate{}, 0, nil\n\t}\n")
	b.WriteString(fmt.Sprintf("\tvar rows []%sRow\n", pascal))
	b.WriteString("\tlistDS := r.dbx.Qu.From(r.table).Select(" + goquSelectCols(ct.ReadCols) + ")")
	if soft != "" {
		b.WriteString(fmt.Sprintf(".Where(goqu.C(%q).IsNull())", soft))
	}
	b.WriteString(fmt.Sprintf(".Order(goqu.I(%q).Desc()).Limit(uint(pageSize)).Offset(uint(offset))", pk))
	b.WriteString("\n\tlistSQL, listArgs, err := listDS.ToSQL()\n")
	b.WriteString("\tif err != nil {\n\t\treturn nil, 0, err\n\t}\n")
	b.WriteString("\tif err := dbx.SelectContext(ctx, r.dbx, &rows, listSQL, listArgs...); err != nil {\n\t\treturn nil, 0, err\n\t}\n")
	b.WriteString("\tout := make([]*domain.Aggregate, 0, len(rows))\n")
	b.WriteString("\tfor i := range rows {\n\t\tout = append(out, rowToAgg(&rows[i]))\n\t}\n")
	b.WriteString("\treturn out, total, nil\n}\n\n")

	// Create
	b.WriteString("func (r *PGRepository) Create(ctx context.Context, agg *domain.Aggregate) (string, error) {\n")
	b.WriteString("\troot := agg.Root()\n\trec := goqu.Record{\n")
	for _, c := range ct.CreateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t\t%q: root.%s,\n", c.Name, f.Name))
	}
	appendTouchTimestamps(&b, ct, true, true)
	b.WriteString("\t}\n")
	b.WriteString("\tds := r.dbx.Qu.Insert(r.table).Rows(rec).Returning(" + goquSelectCols(ct.ReadCols) + ")\n")
	b.WriteString("\tsql, args, err := ds.ToSQL()\n\tif err != nil {\n\t\treturn \"\", err\n\t}\n")
	b.WriteString(fmt.Sprintf("\tvar row %sRow\n", pascal))
	b.WriteString("\tif err := dbx.GetContext(ctx, r.dbx, &row, sql, args...); err != nil {\n\t\treturn \"\", err\n\t}\n")
	b.WriteString("\tid := pkString(&row)\n")
	b.WriteString("\tif r.cacheOn && r.cache != nil {\n")
	b.WriteString("\t\tif raw, err := json.Marshal(&row); err == nil {\n")
	b.WriteString("\t\t\t_ = zcachekey.SetEntityBytes(ctx, r.cache, r.keyPrefix, r.module, id, raw, r.cacheTTL)\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn id, nil\n}\n\n")

	// Update
	b.WriteString("func (r *PGRepository) Update(ctx context.Context, agg *domain.Aggregate) error {\n")
	b.WriteString("\troot := agg.Root()\n")
	b.WriteString(fmt.Sprintf("\tid := pkStringFromRoot(root, %q)\n", pkField.Name))
	b.WriteString("\tr.invalidateEntityCache(ctx, id)\n")
	b.WriteString("\trec := goqu.Record{\n")
	for _, c := range ct.UpdateCols {
		f := schema.GoFieldFromColumn(c)
		b.WriteString(fmt.Sprintf("\t\t%q: root.%s,\n", c.Name, f.Name))
	}
	appendTouchTimestamps(&b, ct, false, true)
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\tds := r.dbx.Qu.Update(r.table).Set(rec).Where(goqu.Ex{%q: id})", pk))
	if soft != "" {
		b.WriteString(fmt.Sprintf(".Where(goqu.C(%q).IsNull())", soft))
	}
	b.WriteString("\n\tsql, args, err := ds.ToSQL()\n\tif err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\tres, err := dbx.ExecContext(ctx, r.dbx, sql, args...)\n\tif err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\tn, _ := res.RowsAffected()\n\tif n == 0 {\n\t\treturn domain.ErrNotFound\n\t}\n")
	b.WriteString("\treturn nil\n}\n\n")

	b.WriteString(fmt.Sprintf("func pkStringFromRoot(root *domain.%s, _ string) string {\n", pascal))
	b.WriteString(fmt.Sprintf("\treturn fmt.Sprint(root.%s)\n}\n\n", pkField.Name))

	// Delete
	b.WriteString("func (r *PGRepository) Delete(ctx context.Context, id string) error {\n")
	b.WriteString("\tr.invalidateEntityCache(ctx, id)\n")
	if soft != "" {
		b.WriteString("\trec := goqu.Record{\n")
		b.WriteString(fmt.Sprintf("\t\t%q: goqu.L(\"CURRENT_TIMESTAMP\"),\n", soft))
		appendTouchTimestamps(&b, ct, false, true)
		b.WriteString("\t}\n")
		b.WriteString(fmt.Sprintf("\tds := r.dbx.Qu.Update(r.table).Set(rec).Where(goqu.Ex{%q: id}).Where(goqu.C(%q).IsNull())", pk, soft))
	} else {
		b.WriteString(fmt.Sprintf("\tds := r.dbx.Qu.Delete(r.table).Where(goqu.Ex{%q: id})", pk))
	}
	b.WriteString("\n\tsql, args, err := ds.ToSQL()\n\tif err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\tres, err := dbx.ExecContext(ctx, r.dbx, sql, args...)\n\tif err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\tn, _ := res.RowsAffected()\n\tif n == 0 {\n\t\treturn domain.ErrNotFound\n\t}\n")
	b.WriteString("\treturn nil\n}\n")

	return b.String()
}

func goquSelectCols(cols []schema.ColumnMeta) string {
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = fmt.Sprintf("%q", c.Name)
	}
	return strings.Join(parts, ", ")
}

func findCol(cols []schema.ColumnMeta, name string) schema.ColumnMeta {
	for _, c := range cols {
		if strings.EqualFold(c.Name, name) {
			return c
		}
	}
	return schema.ColumnMeta{Name: name, DBType: "text"}
}

func genMemoryRepository(projectModule string, ct schema.ClassifiedTable, pascal string, meta ModuleMeta) string {
	pkg := meta.Resource
	pkField := schema.GoFieldFromColumn(findCol(ct.ReadCols, ct.PKColumn))
	return fmt.Sprintf(`package %s

import (
	"context"
	"fmt"
	"sync"

	domain "%s"
)

type repository struct {
	mu   sync.RWMutex
	data map[string]*domain.Aggregate
}

func NewRepository() domain.Repository {
	return &repository{data: make(map[string]*domain.Aggregate)}
}

func (r *repository) FindByID(ctx context.Context, id string) (*domain.Aggregate, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	agg, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return agg, nil
}

func (r *repository) List(ctx context.Context, page, pageSize int) ([]*domain.Aggregate, int64, error) {
	_ = ctx
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	total := int64(len(r.data))
	if total == 0 {
		return []*domain.Aggregate{}, 0, nil
	}
	offset := (page - 1) * pageSize
	if offset >= len(r.data) {
		return []*domain.Aggregate{}, total, nil
	}
	all := make([]*domain.Aggregate, 0, len(r.data))
	for _, agg := range r.data {
		all = append(all, agg)
	}
	end := offset + pageSize
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], total, nil
}

func (r *repository) Create(ctx context.Context, agg *domain.Aggregate) (string, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	id := fmt.Sprint(agg.Root().%s)
	if id == "" || id == "0" {
		id = fmt.Sprintf("%%d", len(r.data)+1)
	}
	r.data[id] = agg
	return id, nil
}

func (r *repository) Update(ctx context.Context, agg *domain.Aggregate) error {
	_ = ctx
	id := fmt.Sprint(agg.Root().%s)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return domain.ErrNotFound
	}
	r.data[id] = agg
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.data, id)
	return nil
}
`, pkg, meta.DomainImportPath(projectModule), pkField.Name, pkField.Name)
}

func appendTouchTimestamps(b *strings.Builder, ct schema.ClassifiedTable, insert, update bool) {
	for _, c := range ct.ReadCols {
		if insert && strings.EqualFold(c.Name, "created_at") {
			b.WriteString(fmt.Sprintf("\t\t%q: goqu.L(\"CURRENT_TIMESTAMP\"),\n", c.Name))
		}
		if (insert || update) && strings.EqualFold(c.Name, "updated_at") {
			b.WriteString(fmt.Sprintf("\t\t%q: goqu.L(\"CURRENT_TIMESTAMP\"),\n", c.Name))
		}
	}
}
