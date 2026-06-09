package generator

import (
	"fmt"
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func grpcWireSnippet(snake, pascal string) string {
	return fmt.Sprintf(`# gRPC 装配说明（%s）

## 生成与注册链

1. goeasy add db proto --table %s → api/proto/%s.proto、internal/interface/grpc/<domain>/<resource>、bootstrap/register_%s_grpc.go
2. goeasy gen proto → api/proto/gen/%s/*.pb.go
3. bootstrap/grpc.go 调用 Register%sGRPC(s, infra)
4. main 已 RegisterGRPC(bootstrap.RegisterGRPCServers)

## 验证

grpcurl -plaintext localhost:<grpc.port> list %s.%sService
grpcurl -plaintext -d "{\"id\":\"1\"}" localhost:<grpc.port> %s.%sService/Get%s
`, snake, snake, snake, snake, snake, pascal, snake, pascal, snake, pascal, pascal)
}

func writeGRPCServerStub(opts DBOptions, projectModule, goeasy string, ct schema.ClassifiedTable) error {
	if err := renderGRPCModule(opts, projectModule, goeasy, ct); err != nil {
		return err
	}
	return nil
}

func writeGRPCWireSnippet(opts DBOptions, snake, pascal string) error {
	rel := filepath.Join("internal", "bootstrap", "snippets", snake+"_grpc.md")
	content := grpcWireSnippet(snake, pascal)
	skipped, err := writeProjectFileOrSkip(opts.ProjectDir, rel, content, opts.Force)
	if err != nil {
		return err
	}
	if !skipped {
		fmt.Printf("  created %s\n", rel)
	}
	return nil
}

func maybeWriteGRPCStubs(opts DBOptions, projectModule, goeasy string, tables []string, dsn, driver, prefix string, rules schema.CodegenRules) error {
	if opts.SkipProto && !opts.WithProto {
		return nil
	}
	for _, physical := range tables {
		module := resolveModuleName(opts, physical, prefix)
		meta, err := loadTableMeta(driver, dsn, opts.Schema, physical)
		if err != nil {
			return err
		}
		ct := schema.Classify(meta, module, physical, rules)
		if err := writeGRPCServerStub(opts, projectModule, goeasy, ct); err != nil {
			return err
		}
	}
	return nil
}
