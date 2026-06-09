package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func relQueryList(meta ModuleMeta) string {
	return meta.appRel("query/list.go")
}

func relAppList(meta ModuleMeta) string {
	return meta.appRel("list.go")
}

func relHandlerCrud(meta ModuleMeta) string {
	return httpModuleRel("admin", meta, "handler_crud.go")
}

func isAdminHandlerCrudRel(rel string) bool {
	rel = filepath.ToSlash(rel)
	return strings.HasSuffix(rel, "/handler_crud.go") && strings.Contains(rel, "internal/interface/http/admin/")
}

// fileHasQueryListImportCycle 检测 query/list.go 是否仍 import 父级 app 包（旧版生成物）。
func fileHasQueryListImportCycle(data []byte, moduleSnake string) bool {
	s := string(data)
	if strings.Contains(s, "/internal/app/"+moduleSnake+`"`) {
		return true
	}
	if strings.Contains(s, "app.ListResult") || strings.Contains(s, "app.ToDTO") {
		return true
	}
	return strings.Contains(s, "/internal/app/") && strings.Contains(s, "\tapp \"") && !strings.Contains(s, "/command")
}

func fileHasStaleHandlerList(data []byte) bool {
	return strings.Contains(string(data), "Queries().List")
}

func isImportCycleCriticalRel(rel, moduleSnake string) bool {
	rel = filepath.ToSlash(rel)
	if strings.HasSuffix(rel, "/query/list.go") {
		return true
	}
	if strings.Contains(rel, "/internal/app/") && strings.HasSuffix(rel, "/list.go") {
		return true
	}
	return isAdminHandlerCrudRel(rel)
}

func isQueryListRel(rel string) bool {
	return strings.HasSuffix(filepath.ToSlash(rel), "/query/list.go")
}

func writeDBGeneratedFile(projectDir, rel, content, moduleSnake string, force bool) (skipped, cycleFixed bool, err error) {
	path := filepath.Join(projectDir, rel)
	shouldWrite := force
	if !shouldWrite {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			shouldWrite = true
		} else if err != nil {
			return false, false, err
		} else if isQueryListRel(rel) {
			existing, readErr := os.ReadFile(path)
			if readErr != nil {
				return false, false, readErr
			}
			if fileHasQueryListImportCycle(existing, moduleSnake) {
				shouldWrite = true
				cycleFixed = true
			}
		} else if isAdminHandlerCrudRel(rel) {
			existing, readErr := os.ReadFile(path)
			if readErr != nil {
				return false, false, readErr
			}
			if fileHasStaleHandlerList(existing) {
				shouldWrite = true
				cycleFixed = true
			}
		}
	}
	if !shouldWrite {
		return true, false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, cycleFixed, err
	}
	if cycleFixed {
		fmt.Fprintf(os.Stderr, "info: fixed import cycle risk in %s\n", rel)
	} else if force {
		fmt.Printf("  updated %s\n", rel)
	} else {
		fmt.Printf("  created %s\n", rel)
	}
	return false, cycleFixed, os.WriteFile(path, []byte(content), 0644)
}

func writeDBGeneratedFiles(projectDir string, files map[string]string, moduleSnake string, force bool) (skippedCount int, skippedCritical []string, cycleFixedCount int, err error) {
	for rel, content := range files {
		skipped, fixed, err := writeDBGeneratedFile(projectDir, rel, content, moduleSnake, force)
		if err != nil {
			return skippedCount, skippedCritical, cycleFixedCount, err
		}
		if fixed {
			cycleFixedCount++
		}
		if skipped {
			skippedCount++
			fmt.Fprintf(os.Stderr, "info: skip existing %s (use --force)\n", rel)
			if isImportCycleCriticalRel(rel, moduleSnake) {
				skippedCritical = append(skippedCritical, rel)
			}
			continue
		}
		if !fixed && !force {
			// created message already printed in writeDBGeneratedFile when !cycleFixed
		}
	}
	return skippedCount, skippedCritical, cycleFixedCount, nil
}

func warnImportCycleSkipped(moduleSnake string, skippedCritical []string) {
	if len(skippedCritical) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "warn: skipped %d file(s) that may cause import cycle (app/%s ↔ query):\n", len(skippedCritical), moduleSnake)
	for _, rel := range skippedCritical {
		fmt.Fprintf(os.Stderr, "  - %s\n", rel)
	}
	fmt.Fprintf(os.Stderr, "warn: re-run with --force or delete stale query/list.go (must not import internal/app/%s)\n", moduleSnake)
}
