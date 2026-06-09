package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

// GenGRPCOptions 从 .proto 生成 gRPC 接口层（需已有 app 层）。
type GenGRPCOptions struct {
	ModuleOptions
	ProtoFile string
	ProtoDir  string
}

var (
	protoMessageRe = regexp.MustCompile(`(?ms)^message\s+(\w+)\s*\{([^}]*)\}`)
	protoFieldRe   = regexp.MustCompile(`(?m)^\s*(?://\s*(.+?)\s*\n\s*)?(\w+)\s+(\w+)\s*=\s*\d+\s*;`)
)

// ProtoContract 从 .proto 源文件解析的 gRPC 契约。
type ProtoContract struct {
	SourceFile   string
	ModuleSnake  string
	ModulePascal string
	CT           schema.ClassifiedTable
}

func resolveProtoSourceFiles(projectDir, single, dir string) ([]string, error) {
	if single != "" {
		if !filepath.IsAbs(single) {
			single = filepath.Join(projectDir, single)
		}
		return []string{single}, nil
	}
	root := dir
	if root == "" {
		root = filepath.Join(projectDir, "api", "proto")
	} else if !filepath.IsAbs(root) {
		root = filepath.Join(projectDir, root)
	}
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "imported" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".proto") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no .proto files under %s", root)
	}
	return out, nil
}

func parseProtoContract(path string) (ProtoContract, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return ProtoContract{}, err
	}
	content := string(b)
	base := strings.TrimSuffix(filepath.Base(path), ".proto")
	moduleSnake := utils.ToSnake(base)

	pkgRe := regexp.MustCompile(`(?m)^package\s+(\w+)\s*;`)
	if m := pkgRe.FindStringSubmatch(content); len(m) == 2 {
		moduleSnake = m[1]
	}

	svcRe := regexp.MustCompile(`service\s+(\w+)Service`)
	pascal := utils.ToPascal(moduleSnake)
	if m := svcRe.FindStringSubmatch(content); len(m) == 2 {
		pascal = strings.TrimSuffix(m[1], "Service")
	}

	messages := map[string][]schema.ColumnMeta{}
	for _, block := range protoMessageRe.FindAllStringSubmatch(content, -1) {
		name := block[1]
		messages[name] = parseProtoMessageFields(block[2])
	}

	readCols := messages[pascal]
	if len(readCols) == 0 {
		readCols = messages[utils.ToPascal(moduleSnake)]
	}
	createCols := messages["Create"+pascal+"Request"]
	updateCols := messages["Update"+pascal+"Request"]

	return ProtoContract{
		SourceFile:   path,
		ModuleSnake:  moduleSnake,
		ModulePascal: pascal,
		CT: schema.ClassifiedTable{
			ModuleName: moduleSnake,
			ReadCols:   readCols,
			CreateCols: createCols,
			UpdateCols: updateCols,
		},
	}, nil
}

func parseProtoMessageFields(body string) []schema.ColumnMeta {
	var out []schema.ColumnMeta
	for _, m := range protoFieldRe.FindAllStringSubmatch(body, -1) {
		comment := strings.TrimSpace(m[1])
		typ := m[2]
		name := m[3]
		out = append(out, schema.ColumnMeta{
			Name:     name,
			DBType:   protoTypeToDB(typ),
			Nullable: true,
			Comment:  comment,
		})
	}
	return out
}

func protoTypeToDB(typ string) string {
	switch typ {
	case "int64", "sint64", "sfixed64":
		return "bigint"
	case "int32", "sint32", "sfixed32":
		return "integer"
	case "bool":
		return "bool"
	case "float", "double":
		return "float8"
	default:
		return "varchar"
	}
}
