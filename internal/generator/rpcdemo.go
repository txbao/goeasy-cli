package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const rpcdemoModuleName = "rpcdemo"

// RPCDemoOptions add rpcdemo 参数。
type RPCDemoOptions struct {
	ProjectDir     string
	RemoteService  string // 对端逻辑服务名（app_name），默认 user
	ProtoModule    string // 对端 gRPC 模块名，默认 sys_roles
	FromURL        string // 对端 .proto 路径/URL；空则尝试 ../<remote>/api/proto/<proto>.proto
	SkipFetchProto bool
	ConfigPath     string
	AppStyle       string // CLI --app-style；空则读 config
	Force          bool
}

// GenerateRPCDemo 生成跨服务 gRPC 调用示范模块（DDD Lite + Gateway）。
func GenerateRPCDemo(opts RPCDemoOptions) error {
	if opts.RemoteService == "" {
		opts.RemoteService = "user"
	}
	opts.ProtoModule = resolveRPCDemoProtoModule(opts.ProtoModule, opts.FromURL)
	if strings.TrimSpace(opts.FromURL) != "" && strings.TrimSpace(opts.ProtoModule) != "" {
		inferred := inferProtoModuleFromURL(opts.FromURL)
		if inferred != "" && inferred != opts.ProtoModule {
			fmt.Fprintf(os.Stderr, "warn: --from-url points to %s.proto but --proto is %s; using --proto\n", inferred, opts.ProtoModule)
		}
	}
	if moduleExists(opts.ProjectDir, moduleMetaByID(rpcdemoModuleName)) && !opts.Force {
		if err := ensureRPCDemoModulesRegistry(opts); err != nil {
			return err
		}
		projectModule, err := readModulePath(opts.ProjectDir)
		if err != nil {
			return err
		}
		if err := ensureBootstrapRPCFile(opts, toMap(BuildRPCDemoData(projectModule, opts.RemoteService))); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "info: module %q already exists, skipping (use --force to overwrite)\n", rpcdemoModuleName)
		return nil
	}
	style, err := resolveAppStyleForModule(ModuleOptions{
		ProjectDir: opts.ProjectDir,
		ConfigPath: opts.ConfigPath,
		AppStyle:   opts.AppStyle,
		ModuleName: rpcdemoModuleName,
	})
	if err != nil {
		return err
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	protoMeta := resolveRPCDemoProtoMeta(projectModule, opts.ProjectDir, opts.ProtoModule, opts.RemoteService, opts.FromURL)
	if err := ensureRPCDemoProto(opts, projectModule, protoMeta); err != nil {
		return err
	}
	data := toMap(BuildRPCDemoData(projectModule, opts.RemoteService))
	enrichRPCDemoData(data, protoMeta)
	repl := map[string]string{
		"REMOTE": opts.RemoteService,
	}
	if err := renderScoped("rpcdemo", opts.ProjectDir, repl, data, opts.Force,
		"internal/interface/http/admin/rpcdemo/dto.go",
	); err != nil {
		return err
	}
	if err := writeRPCDemoGeneratedFiles(opts, projectModule, style, protoMeta); err != nil {
		return err
	}
	if err := ensureBootstrapRPCFile(opts, data); err != nil {
		return err
	}
	return ensureRPCDemoModulesRegistry(opts)
}

func BuildRPCDemoData(projectModule, remoteService string) TemplateData {
	d := BuildModuleData(rpcdemoModuleName, projectModule)
	d.RemoteService = remoteService
	return d
}

func ensureRPCDemoModulesRegistry(opts RPCDemoOptions) error {
	const funcName = "RegisterRPCDemo"
	callLine := moduleRegisterCallLine(funcName)

	modulesPath := filepath.Join(opts.ProjectDir, "internal", "bootstrap", "modules.go")
	if _, err := os.Stat(modulesPath); os.IsNotExist(err) {
		if err := renderModulesFile(ModuleOptions{ProjectDir: opts.ProjectDir}); err != nil {
			return err
		}
	}
	b, err := os.ReadFile(modulesPath)
	if err != nil {
		return err
	}
	content := string(b)
	if strings.Contains(content, funcName+"(") {
		return nil
	}
	if !strings.Contains(content, modulesRegistryMarker) {
		repaired, ok := insertRegistryMarker(content, modulesRegistryMarker)
		if !ok {
			fmt.Fprintf(os.Stderr, "warn: modules.go missing registry marker; skip %s\n", funcName)
			return nil
		}
		content = repaired
	}
	updated := strings.Replace(content, modulesRegistryMarker+"\n", modulesRegistryMarker+"\n"+callLine+"\n", 1)
	if updated == content {
		updated = strings.Replace(content, modulesRegistryMarker, modulesRegistryMarker+"\n"+callLine, 1)
	}
	if err := os.WriteFile(modulesPath, []byte(updated), 0644); err != nil {
		return err
	}
	fmt.Printf("  updated internal/bootstrap/modules.go (+%s)\n", funcName)
	return nil
}

func ensureBootstrapRPCFile(opts RPCDemoOptions, data map[string]any) error {
	targetRel := "internal/bootstrap/rpc.go"
	targetPath := filepath.Join(opts.ProjectDir, targetRel)
	if !opts.Force {
		if b, err := os.ReadFile(targetPath); err == nil && strings.Contains(string(b), "RPCClientLazy") {
			return nil
		}
	}
	sub, err := fsSub("project")
	if err != nil {
		return err
	}
	content, err := fs.ReadFile(sub, "internal/bootstrap/rpc.go.tmpl")
	if err != nil {
		return fmt.Errorf("read rpc.go.tmpl: %w", err)
	}
	out, err := executeTemplate("rpc.go.tmpl", content, data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(targetPath, out, 0644); err != nil {
		return err
	}
	fmt.Printf("  created %s\n", targetRel)
	return nil
}
