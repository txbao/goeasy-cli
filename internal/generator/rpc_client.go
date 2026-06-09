package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
)

// RPCClientOptions add rpc（步骤 1：共享 proto + gateway + port）。
type RPCClientOptions struct {
	ProjectDir     string
	RemoteService  string
	ProtoModule    string
	FromURL        string
	SkipFetchProto bool
	Methods        string // 默认 all：Get,Create,Update,Delete,List
	Force          bool
}

// RPCClientBindOptions add rpc bind（步骤 2：consumer port + register wire）。
type RPCClientBindOptions struct {
	ProjectDir    string
	ProtoModule   string
	Consumer      string
	RemoteService string // 可选；空则从已有 gateway 目录推断
	Wire          bool
	Force         bool
	ConfigPath    string
}

// RPCClientMeta 跨服务 RPC 客户端元数据（共享 port + gateway）。
type RPCClientMeta struct {
	RPCDemoProtoMeta
	SharedPortRel    string
	SharedPortImport string
	UpdateFields     []rpcViewField
	Methods          rpcMethodSet
}

type rpcMethodSet struct {
	Get    bool
	Create bool
	Update bool
	Delete bool
	List   bool
}

func parseRPCMethods(raw string) rpcMethodSet {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" {
		return rpcMethodSet{Get: true, Create: true, Update: true, Delete: true, List: true}
	}
	set := rpcMethodSet{}
	for _, part := range strings.Split(raw, ",") {
		switch strings.TrimSpace(part) {
		case "get":
			set.Get = true
		case "create":
			set.Create = true
		case "update":
			set.Update = true
		case "delete":
			set.Delete = true
		case "list":
			set.List = true
		}
	}
	return set
}

func resolveRPCClientMeta(projectModule, projectDir, protoModule, remoteService, fromURL, methods string) RPCClientMeta {
	base := resolveRPCDemoProtoMeta(projectModule, projectDir, protoModule, remoteService, fromURL)
	updateFields := defaultUpdateFields(base)

	if contract, ok := loadRPCDemoProtoContract(projectDir, base.Module, remoteService, fromURL); ok {
		if len(contract.CT.UpdateCols) > 0 {
			updateFields = updateFieldsFromProto(contract.CT.UpdateCols)
		}
	}

	sharedPortRel := filepath.ToSlash(filepath.Join("internal", "infrastructure", "rpc", remoteService, "port", base.Module+".go"))
	return RPCClientMeta{
		RPCDemoProtoMeta: base,
		SharedPortRel:    sharedPortRel,
		SharedPortImport: projectModule + "/" + filepath.ToSlash(filepath.Join("internal", "infrastructure", "rpc", remoteService, "port")),
		UpdateFields:     updateFields,
		Methods:          parseRPCMethods(methods),
	}
}

func defaultUpdateFields(base RPCDemoProtoMeta) []rpcViewField {
	if preset, ok := rpcDemoProtoPresets[base.Module]; ok {
		return presetCreateFields(preset.fields)
	}
	return append([]rpcViewField(nil), base.CreateFields...)
}

func updateFieldsFromProto(cols []schema.ColumnMeta) []rpcViewField {
	skipID := map[string]bool{"id": true}
	var out []rpcViewField
	for _, c := range cols {
		if skipID[c.Name] {
			continue
		}
		out = append(out, columnToRPCViewField(c))
	}
	return out
}

// GenerateRPCClient 生成共享 RPC 基础设施（proto、pb、gateway、共享 port）。
func GenerateRPCClient(opts RPCClientOptions) error {
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

	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	meta := resolveRPCClientMeta(projectModule, opts.ProjectDir, opts.ProtoModule, opts.RemoteService, opts.FromURL, opts.Methods)

	demoOpts := RPCDemoOptions{
		ProjectDir:     opts.ProjectDir,
		RemoteService:  opts.RemoteService,
		ProtoModule:    opts.ProtoModule,
		FromURL:        opts.FromURL,
		SkipFetchProto: opts.SkipFetchProto,
	}
	if err := ensureRPCDemoProto(demoOpts, projectModule, meta.RPCDemoProtoMeta); err != nil {
		return err
	}

	files := map[string]string{
		meta.SharedPortRel: genSharedRPCPort(meta),
		meta.GatewayFile:   genSharedRPCGateway(projectModule, opts.RemoteService, meta),
	}
	for rel, content := range files {
		if err := writeRPCDemoFile(opts.ProjectDir, rel, content, opts.Force); err != nil {
			return err
		}
	}

	rpcData := toMap(BuildRPCDemoData(projectModule, opts.RemoteService))
	if err := ensureBootstrapRPCFile(RPCDemoOptions{ProjectDir: opts.ProjectDir, Force: opts.Force}, rpcData); err != nil {
		return err
	}
	return nil
}

// GenerateRPCClientBind 为业务模块绑定 RPC port 并可选 wire 到 register。
func GenerateRPCClientBind(opts RPCClientBindOptions) error {
	consumer := strings.TrimSpace(opts.Consumer)
	if consumer == "" {
		return fmt.Errorf("--consumer is required")
	}
	opts.ProtoModule = resolveRPCDemoProtoModule(opts.ProtoModule, "")

	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}

	remote := strings.TrimSpace(opts.RemoteService)
	if remote == "" {
		remote, err = findRPCGatewayRemote(opts.ProjectDir, opts.ProtoModule)
		if err != nil {
			return err
		}
	}

	meta := resolveRPCClientMeta(projectModule, opts.ProjectDir, opts.ProtoModule, remote, "", "all")
	gatewayPath := filepath.Join(opts.ProjectDir, filepath.FromSlash(meta.GatewayFile))
	if _, err := os.Stat(gatewayPath); err != nil {
		return fmt.Errorf("missing %s; run first: goeasy-cli add rpc %s --remote <service> --from-url <proto>",
			meta.GatewayFile, opts.ProtoModule)
	}
	portPath := filepath.Join(opts.ProjectDir, filepath.FromSlash(meta.SharedPortRel))
	if _, err := os.Stat(portPath); err != nil {
		return fmt.Errorf("missing %s; run first: goeasy-cli add rpc %s", meta.SharedPortRel, opts.ProtoModule)
	}

	consumerMeta := resolveModuleMetaForModule(ModuleOptions{
		ProjectDir: opts.ProjectDir,
		ModuleName: consumer,
		ConfigPath: opts.ConfigPath,
	}, opts.ConfigPath)

	consumerPortRel := filepath.ToSlash(filepath.Join("internal", "app", consumerMeta.Domain, consumerMeta.Resource, "port", meta.Module+".go"))
	if err := writeRPCDemoFile(opts.ProjectDir, consumerPortRel, genConsumerRPCPortAlias(projectModule, remote, meta), opts.Force); err != nil {
		return err
	}

	if opts.Wire {
		if err := wireRPCClientToRegister(opts, projectModule, consumerMeta, meta, remote); err != nil {
			snippetRel := filepath.ToSlash(filepath.Join("internal", "bootstrap", "snippets", consumer+"_rpc_"+meta.Module+".md"))
			if writeErr := writeRPCWireSnippet(opts.ProjectDir, snippetRel, consumerMeta, meta, remote); writeErr != nil {
				return fmt.Errorf("wire failed: %w (also failed to write snippet: %v)", err, writeErr)
			}
			fmt.Fprintf(os.Stderr, "warn: auto wire failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "info: manual steps written to %s\n", snippetRel)
		}
	}

	fmt.Fprintf(os.Stderr, "info: inject %s into %s NewApplication manually (see bootstrap snippets if wire skipped)\n",
		rpcClientGWVar(meta.Pascal), consumerMeta.AppImportPath(projectModule))
	return nil
}

func findRPCGatewayRemote(projectDir, protoModule string) (string, error) {
	rpcRoot := filepath.Join(projectDir, "internal", "infrastructure", "rpc")
	entries, err := os.ReadDir(rpcRoot)
	if err != nil {
		return "", fmt.Errorf("no rpc infrastructure under internal/infrastructure/rpc; run add rpc first")
	}
	var found []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		gw := filepath.Join(rpcRoot, e.Name(), protoModule+"_gateway.go")
		if st, err := os.Stat(gw); err == nil && !st.IsDir() {
			found = append(found, e.Name())
		}
	}
	switch len(found) {
	case 0:
		return "", fmt.Errorf("gateway for %s not found; run: goeasy-cli add rpc %s --remote <service>", protoModule, protoModule)
	case 1:
		return found[0], nil
	default:
		return "", fmt.Errorf("multiple gateways for %s (%s); pass --remote", protoModule, strings.Join(found, ", "))
	}
}

func rpcClientGWVar(pascal string) string {
	return rpcDemoGWParam(pascal) + "GW"
}
