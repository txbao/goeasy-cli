package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateGRPCFromProto 契约驱动：从 .proto 生成 gRPC 接口层与 bootstrap 注册（需已有 app 层）。
func GenerateGRPCFromProto(opts GenGRPCOptions) error {
	files, err := resolveProtoSourceFiles(opts.ProjectDir, opts.ProtoFile, opts.ProtoDir)
	if err != nil {
		return err
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := generateGRPCFromProtoFile(opts, projectModule, f); err != nil {
			return err
		}
	}
	return nil
}

func generateGRPCFromProtoFile(opts GenGRPCOptions, projectModule, protoPath string) error {
	contract, err := parseProtoContract(protoPath)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "info: gen grpc from %s -> module %s\n", filepath.ToSlash(protoPath), contract.ModuleSnake)

	dbOpts := DBOptions{
		ModuleOptions: ModuleOptions{
			ProjectDir: opts.ProjectDir,
			ModuleName: contract.ModuleSnake,
			Force:      opts.Force,
			Group:      opts.Group,
			Resource:   opts.Resource,
			ConfigPath: defaultConfigPath(opts.ConfigPath),
		},
	}
	layoutMeta := moduleMetaFromDB(dbOpts, contract.ModuleSnake)
	if !appModuleExists(opts.ProjectDir, layoutMeta) {
		return fmt.Errorf("gRPC from proto requires app layer: run gen http or add db crud for %s first", contract.ModuleSnake)
	}
	return writeGRPCServerStub(dbOpts, projectModule, currentGoEasyModule(), contract.CT)
}
