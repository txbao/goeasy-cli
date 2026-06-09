package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenContractOptions 批量从 api/ 契约生成 HTTP + gRPC 桩。
type GenContractOptions struct {
	GenHTTPOptions
	SkipHTTP  bool
	SkipGRPC  bool
	WithProto bool // 顺带 gen proto（protoc → *.pb.go）
}

// GenerateFromContracts 扫描 api/openapi 与 api/proto，生成接口层与应用层桩。
func GenerateFromContracts(opts GenContractOptions) error {
	if !opts.SkipHTTP {
		o := opts.GenHTTPOptions
		if o.OpenAPIFile == "" && o.OpenAPIDir == "" {
			o.OpenAPIDir = filepath.Join(opts.ProjectDir, APIContractsOpenAPI)
		}
		if err := GenerateHTTPFromOpenAPI(o); err != nil {
			if !isContractDirEmptyErr(err) {
				return fmt.Errorf("gen http: %w", err)
			}
			fmt.Fprintf(os.Stderr, "info: skip gen http (%v)\n", err)
		}
	}
	if opts.WithProto {
		if err := GenerateProtoGo(GenProtoOptions{ProjectDir: opts.ProjectDir}); err != nil {
			return fmt.Errorf("gen proto: %w", err)
		}
	}
	if !opts.SkipGRPC {
		g := GenGRPCOptions{
			ModuleOptions: opts.GenHTTPOptions.ModuleOptions,
			ProtoDir:      filepath.Join(opts.ProjectDir, "api", "proto"),
		}
		if err := GenerateGRPCFromProto(g); err != nil {
			if !isContractDirEmptyErr(err) {
				return fmt.Errorf("gen grpc: %w", err)
			}
			fmt.Fprintf(os.Stderr, "info: skip gen grpc (%v)\n", err)
		}
	}
	return nil
}

func isContractDirEmptyErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "no OpenAPI files") || strings.Contains(s, "no .proto files")
}
