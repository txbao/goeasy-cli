package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

func GenerateDBProto(opts DBOptions) error {
	dsn, driver, prefix, rules, err := readProjectCodegen(opts.ProjectDir, opts.ConfigPath)
	if err != nil {
		return err
	}
	tables, err := resolveDBTables(opts, dsn, prefix, driver)
	if err != nil {
		return err
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	var protoRels []string
	for _, physical := range tables {
		module := resolveModuleName(opts, physical, prefix)
		meta, err := loadTableMeta(driver, dsn, opts.Schema, physical)
		if err != nil {
			return err
		}
		ct := schema.Classify(meta, module, physical, rules)
		pascal := utils.ToPascal(module)
		snake := utils.ToSnake(module)
		if err := requireAppLayerForGRPC(opts.ProjectDir, snake, opts); err != nil {
			return err
		}
		content := genProtoFile(projectModule, ct, pascal, snake)
		rel := filepath.ToSlash(filepath.Join("api", "proto", snake+".proto"))
		skipped, err := writeProjectFileOrSkip(opts.ProjectDir, rel, content, opts.Force)
		if err != nil {
			return err
		}
		if skipped {
			fmt.Fprintf(os.Stderr, "info: skip existing %s (use --force)\n", rel)
		} else {
			fmt.Printf("  created %s\n", rel)
		}
		protoRels = append(protoRels, rel)
		if err := writeGRPCServerStub(opts, projectModule, currentGoEasyModule(), ct); err != nil {
			return err
		}
	}
	maybeRunGenProtoAfterDB(opts.ProjectDir, protoRels, opts.SkipGenProto)
	return nil
}

func protoFieldLine(c schema.ColumnMeta, typ, name string, num int) string {
	var b strings.Builder
	if strings.TrimSpace(c.Comment) != "" {
		b.WriteString("  // ")
		b.WriteString(strings.TrimSpace(c.Comment))
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("  %s %s = %d;\n", typ, name, num))
	return b.String()
}

func genProtoFile(projectModule string, ct schema.ClassifiedTable, pascal, snake string) string {
	var b strings.Builder
	if ct.TableComment != "" {
		b.WriteString("// ")
		b.WriteString(ct.TableComment)
		b.WriteString("\n")
	}
	b.WriteString("syntax = \"proto3\";\n\n")
	b.WriteString(fmt.Sprintf("package %s;\n\n", snake))
	b.WriteString(fmt.Sprintf("option go_package = \"%s/api/proto/gen/%s;%spb\";\n\n", projectModule, snake, snake))
	b.WriteString(fmt.Sprintf("service %sService {\n", pascal))
	b.WriteString(fmt.Sprintf("  rpc List%s (List%sRequest) returns (List%sResponse);\n", pascal, pascal, pascal))
	b.WriteString(fmt.Sprintf("  rpc Get%s (Get%sRequest) returns (%s);\n", pascal, pascal, pascal))
	b.WriteString(fmt.Sprintf("  rpc Create%s (Create%sRequest) returns (%s);\n", pascal, pascal, pascal))
	b.WriteString(fmt.Sprintf("  rpc Update%s (Update%sRequest) returns (%s);\n", pascal, pascal, pascal))
	b.WriteString(fmt.Sprintf("  rpc Delete%s (Delete%sRequest) returns (Delete%sResponse);\n", pascal, pascal, pascal))
	b.WriteString("}\n\n")

	b.WriteString(fmt.Sprintf("message %s {\n", pascal))
	for i, c := range ct.ReadCols {
		b.WriteString(protoFieldLine(c, schema.ProtoType(c), c.Name, i+1))
	}
	b.WriteString("}\n\n")

	b.WriteString(fmt.Sprintf("message List%sRequest {\n", pascal))
	b.WriteString("  int32 page = 1;\n  int32 page_size = 2;\n}\n\n")
	b.WriteString(fmt.Sprintf("message List%sResponse {\n", pascal))
	b.WriteString(fmt.Sprintf("  repeated %s list = 1;\n", pascal))
	b.WriteString("  int64 total = 2;\n  int32 page = 3;\n  int32 page_size = 4;\n  int32 total_pages = 5;\n}\n\n")

	b.WriteString(fmt.Sprintf("message Get%sRequest {\n  string id = 1;\n}\n\n", pascal))
	b.WriteString(fmt.Sprintf("message Create%sRequest {\n", pascal))
	for i, c := range ct.CreateCols {
		b.WriteString(protoFieldLine(c, schema.ProtoType(c), c.Name, i+1))
	}
	b.WriteString("}\n\n")
	b.WriteString(fmt.Sprintf("message Update%sRequest {\n  string id = 1;\n", pascal))
	for i, c := range ct.UpdateCols {
		b.WriteString(protoFieldLine(c, schema.ProtoType(c), c.Name, i+2))
	}
	b.WriteString("}\n\n")
	b.WriteString(fmt.Sprintf("message Delete%sRequest {\n  string id = 1;\n}\n\n", pascal))
	b.WriteString(fmt.Sprintf("message Delete%sResponse {\n  bool ok = 1;\n}\n", pascal))
	return b.String()
}

func GenerateDBAll(opts DBOptions) error {
	opts.All = true
	if err := GenerateDBCRUD(opts); err != nil {
		return err
	}
	if opts.WithProto && !opts.SkipProto {
		o2 := opts
		o2.All = true
		if err := GenerateDBProto(o2); err != nil {
			return err
		}
	}
	return nil
}

func maybeGenerateDBOpenAPI(opts DBOptions) error {
	if !opts.WithOpenAPI || opts.SkipOpenAPI {
		return nil
	}
	return GenerateDBOpenAPI(opts)
}
