package generator

import (
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func writeCommandImportBlock(b *strings.Builder, projectModule string, meta ModuleMeta, cols []schema.ColumnMeta, needDomain bool) {
	b.WriteString("import (\n\t\"context\"\n")
	if schema.ColsNeedTimeImport(cols) {
		b.WriteString("\t\"time\"\n")
	}
	if needDomain {
		b.WriteString("\n")
		b.WriteString("\tdomain \"" + meta.DomainImportPath(projectModule) + "\"\n")
	}
	b.WriteString(")\n\n")
}

func persistenceRepoRel(meta ModuleMeta, file string) string {
	return meta.persistenceRel(file)
}

func sharedDBXImport(projectModule string) string {
	return filepath.ToSlash(filepath.Join(projectModule, "internal", "infrastructure", "shared", "dbx"))
}
