package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const rpcWireMarkerPrefix = "// goeasy-rpc-wire: "

func wireRPCClientToRegister(opts RPCClientBindOptions, projectModule string, consumerMeta ModuleMeta, meta RPCClientMeta, remote string) error {
	registerRel := filepath.ToSlash(filepath.Join("internal", "bootstrap", "register_"+consumerMeta.Domain+".go"))
	registerPath := filepath.Join(opts.ProjectDir, registerRel)
	b, err := os.ReadFile(registerPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", registerRel, err)
	}
	content := string(b)
	moduleMarker := "// goeasy-module: " + consumerMeta.ModuleID
	if !strings.Contains(content, moduleMarker) {
		return fmt.Errorf("%s missing marker %s", registerRel, moduleMarker)
	}
	wireMarker := rpcWireMarkerPrefix + meta.Module
	if strings.Contains(content, wireMarker) {
		fmt.Fprintf(os.Stderr, "info: %s already wired for %s, skipping\n", consumerMeta.ModuleID, meta.Module)
		return nil
	}

	fnBody, fnStart, ok := extractModuleRegisterFunc(content, consumerMeta)
	if !ok {
		return fmt.Errorf("cannot locate register function for %s in %s", consumerMeta.ModuleID, registerRel)
	}

	remoteImport := projectModule + "/" + filepath.ToSlash(filepath.Join("internal", "infrastructure", "rpc", remote))
	remoteAlias := rpcRemoteImportAlias(content, remote)

	var insert strings.Builder
	if !strings.Contains(fnBody, "RPCClientLazy(infra, \""+remote+"\")") {
		insert.WriteString(fmt.Sprintf("\tcli, err := RPCClientLazy(infra, %q)\n", remote))
		insert.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n")
	}
	gwVar := rpcClientGWVar(meta.Pascal)
	insert.WriteString(wireMarker + "\n")
	insert.WriteString(fmt.Sprintf("\t%s := %s.New%s(cli)\n", gwVar, remoteAlias, meta.GatewayName))
	insert.WriteString("\t_ = " + gwVar + "\n")

	newFnBody := insert.String() + fnBody
	newContent := content[:fnStart] + newFnBody + content[fnStart+len(fnBody):]
	newContent = ensureRemoteRPCImport(newContent, remoteAlias, remoteImport)
	newContent = ensureRPCClientLazyImport(newContent)

	if err := os.WriteFile(registerPath, []byte(newContent), 0644); err != nil {
		return err
	}
	fmt.Printf("  updated %s (+%s wire)\n", registerRel, meta.Module)
	return nil
}

func extractModuleRegisterFunc(content string, meta ModuleMeta) (body string, bodyStart int, ok bool) {
	moduleMarker := "// goeasy-module: " + meta.ModuleID
	idx := strings.Index(content, moduleMarker)
	if idx < 0 {
		return "", 0, false
	}
	fnIdx := strings.LastIndex(content[:idx], "func ")
	if fnIdx < 0 {
		return "", 0, false
	}
	brace := strings.Index(content[fnIdx:], "{")
	if brace < 0 {
		return "", 0, false
	}
	bodyStart = fnIdx + brace + 1
	depth := 1
	for i := bodyStart; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[bodyStart:i], bodyStart, true
			}
		}
	}
	return "", 0, false
}

func rpcRemoteImportAlias(content, remote string) string {
	needle := `internal/infrastructure/rpc/` + remote + `"`
	if idx := strings.Index(content, needle); idx >= 0 {
		lineStart := strings.LastIndex(content[:idx], "\n") + 1
		line := content[lineStart:idx]
		if aliasIdx := strings.Index(line, `"`); aliasIdx >= 0 {
			before := strings.TrimSpace(line[:aliasIdx])
			if before != "" && !strings.HasPrefix(before, "import") {
				return strings.TrimSpace(strings.TrimSuffix(before, `"`))
			}
		}
	}
	return remote
}

func ensureRemoteRPCImport(content, alias, importPath string) string {
	if strings.Contains(content, importPath) {
		return content
	}
	importLine := fmt.Sprintf("\t%s \"%s\"\n", alias, importPath)
	return insertIntoImportBlock(content, importLine)
}

func ensureRPCClientLazyImport(content string) string {
	if strings.Contains(content, "RPCClientLazy") {
		return content
	}
	return content
}

func insertIntoImportBlock(content, importLine string) string {
	const imp = "import ("
	idx := strings.Index(content, imp)
	if idx < 0 {
		return content
	}
	close := strings.Index(content[idx:], ")")
	if close < 0 {
		return content
	}
	pos := idx + close
	return content[:pos] + importLine + content[pos:]
}

func writeRPCWireSnippet(projectDir, rel string, consumerMeta ModuleMeta, meta RPCClientMeta, remote string) error {
	gwVar := rpcClientGWVar(meta.Pascal)
	body := fmt.Sprintf(`# RPC wire: %s → %s

在 `+"`register_%s.go`"+` 的 `+"`// goeasy-module: %s`"+` 函数内添加：

`+"```go"+`
cli, err := RPCClientLazy(infra, %q)
if err != nil {
	return err
}
%s := remote.New%s(cli)
`+"```"+`

并在 `+"`NewApplication`"+` 中注入 `+"`%s`"+`（需手改 application.go）。

共享 Port：`+"`internal/infrastructure/rpc/%s/port/%s.go`"+`
业务 Port alias：`+"`internal/app/%s/%s/port/%s.go`"+`
`,
		meta.Module, consumerMeta.ModuleID,
		consumerMeta.Domain, consumerMeta.ModuleID,
		remote, gwVar, meta.GatewayName,
		gwVar,
		remote, meta.Module,
		consumerMeta.Domain, consumerMeta.Resource, meta.Module,
	)
	target := filepath.Join(projectDir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	return os.WriteFile(target, []byte(body), 0644)
}
