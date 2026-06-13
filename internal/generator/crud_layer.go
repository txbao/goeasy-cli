package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

// ensureModuleCRUDLayer 为 add crud 补全 List 能力（domain/memory/app/dto）。
func ensureModuleCRUDLayer(opts ModuleOptions) error {
	style, err := resolveAppStyleForModule(opts)
	if err != nil {
		return err
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	meta := resolveModuleMetaForModule(opts, opts.ConfigPath)
	withAudit := resolveModuleAudit(opts, meta)
	pascal := utils.ToPascal(opts.ModuleName)

	var files map[string]string
	if style.IsService() {
		files = genModuleServiceCRUDLayer(projectModule, meta, pascal, withAudit)
	} else {
		files = genModuleLightCQRSCRUDLayer(projectModule, meta, pascal, withAudit)
	}
	for rel, content := range files {
		target := filepath.Join(opts.ProjectDir, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(content), 0644); err != nil {
			return err
		}
		fmt.Printf("  created %s\n", filepath.ToSlash(rel))
	}
	return nil
}

func removeModuleListQuery(projectDir string, meta ModuleMeta) {
	_ = os.Remove(filepath.Join(projectDir, meta.appRel("query/list.go")))
}

func genModuleDomainRepository(pkg string) string {
	return fmt.Sprintf(`package %s

import "context"

type Repository interface {
	FindByID(ctx context.Context, id string) (*Aggregate, error)
	List(ctx context.Context, page, pageSize int) ([]*Aggregate, int64, error)
	Save(ctx context.Context, agg *Aggregate) error
	Delete(ctx context.Context, id string) error
}
`, pkg)
}

func genModuleMemoryRepository(projectModule string, meta ModuleMeta) string {
	var b strings.Builder
	b.WriteString("package ")
	b.WriteString(meta.Resource)
	b.WriteString(`

import (
	"context"
	"fmt"
	"sync"

	domain "`)
	b.WriteString(meta.DomainImportPath(projectModule))
	b.WriteString(`"
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
		return nil, fmt.Errorf("`)
	b.WriteString(meta.ModuleID)
	b.WriteString(` %s not found", id)
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

func (r *repository) Save(ctx context.Context, agg *domain.Aggregate) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[agg.Root().ID()] = agg
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}
`)
	return b.String()
}

func genModuleListQuery(projectModule string, meta ModuleMeta) string {
	return genListQuery(projectModule, schema.ClassifiedTable{ModuleName: meta.ModuleID}, "", meta)
}

func genModuleDTO(projectModule string, meta ModuleMeta, pascal string) string {
	return fmt.Sprintf(`package %s

import domain "%s"

type %sDTO struct {
	ID     string `+"`json:\"id\"`"+`
	Active bool   `+"`json:\"active\"`"+`
}

func ToDTO(agg *domain.Aggregate) %sDTO {
	r := agg.Root()
	return %sDTO{ID: r.ID(), Active: r.Active()}
}

// ListResult 列表查询结果（供 HTTP List 使用）。
type ListResult struct {
	List  []%sDTO
	Total int64
}
`, meta.Resource, meta.DomainImportPath(projectModule), pascal, pascal, pascal, pascal)
}
