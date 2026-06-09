package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/schema"
	"github.com/txbao/goeasy-cli/internal/utils"
)

const grpcBootstrapMarker = "// grpc bootstrap modules (goeasy add db proto appends below)"

func appModuleExists(projectDir string, meta ModuleMeta) bool {
	_, err := os.Stat(filepath.Join(projectDir, meta.appRel("application.go")))
	return err == nil
}

func requireAppLayerForGRPC(projectDir, snake string, opts DBOptions) error {
	meta := moduleMetaFromDB(opts, snake)
	if !appModuleExists(projectDir, meta) {
		return fmt.Errorf("gRPC handlers require app layer: run add db crud --table %s first", snake)
	}
	return nil
}

func renderGRPCModule(opts DBOptions, projectModule, goeasy string, ct schema.ClassifiedTable) error {
	snake := utils.ToSnake(ct.ModuleName)
	pascal := utils.ToPascal(ct.ModuleName)
	alias := utils.ToIdent(ct.ModuleName)
	layoutMeta := moduleMetaFromDB(opts, ct.ModuleName)
	if err := requireAppLayerForGRPC(opts.ProjectDir, snake, opts); err != nil {
		return err
	}
	if err := ensureGRPCRegisterFile(projectModule, opts.ProjectDir); err != nil {
		return err
	}

	modOpts := ModuleOptions{
		ProjectDir: opts.ProjectDir,
		ModuleName: ct.ModuleName,
		Force:      opts.Force,
		Domain:     opts.Domain,
		Group:      opts.Group,
		Resource:   opts.Resource,
		ConfigPath: opts.ConfigPath,
		AppStyle:   opts.AppStyle,
	}
	style, err := resolveAppStyleForModule(modOpts)
	if err != nil {
		return err
	}

	data := toMap(BuildModuleData(ct.ModuleName, projectModule))
	enrichModuleMetaData(data, layoutMeta)
	enrichGRPCData(data, ct, opts.ProjectDir, true, layoutMeta)
	repl := moduleTemplateRepl(layoutMeta)

	handlersRel := layoutMeta.grpcRel("handlers.go")
	if err := renderScopedOrSkip("grpc", opts.ProjectDir, repl, data, opts.Force, handlersRel); err != nil {
		return err
	}

	readCols := data["ReadCols"].([]GRPCCol)
	createCols := data["CreateCols"].([]GRPCCol)
	updateCols := data["UpdateCols"].([]GRPCCol)
	handlers := genGRPCHandlers(projectModule, goeasy, pascal, alias, layoutMeta, snake, style, readCols, createCols, updateCols)
	skipped, err := writeProjectFileOrSkip(opts.ProjectDir, handlersRel, handlers, opts.Force)
	if err != nil {
		return err
	}
	if !skipped {
		fmt.Printf("  created %s\n", handlersRel)
	}

	pbImport := projectModule + "/api/proto/gen/" + snake
	convert := genGRPCConvertGo(layoutMeta, projectModule, data["ModuleAlias"].(string), pascal, pbImport,
		data["ReadCols"].([]GRPCCol),
		data["CreateCols"].([]GRPCCol),
		data["UpdateCols"].([]GRPCCol),
	)
	convertPath := layoutMeta.grpcRel("convert.go")
	skipped, err = writeProjectFileOrSkip(opts.ProjectDir, convertPath, convert, opts.Force)
	if err != nil {
		return err
	}
	if !skipped {
		fmt.Printf("  created %s\n", convertPath)
	}

	if err := renderRegisterGRPCFile(opts, data); err != nil {
		return err
	}
	return ensureGRPCBootstrapRegistry(opts, pascal)
}

func renderRegisterGRPCFile(opts DBOptions, data map[string]any) error {
	sub, err := fsSub("grpc")
	if err != nil {
		return err
	}
	tplPath := "internal/bootstrap/register_MODULE_grpc.go.tmpl"
	content, err := fs.ReadFile(sub, tplPath)
	if err != nil {
		return fmt.Errorf("read grpc register template: %w", err)
	}
	out, err := executeTemplate(filepath.Base(tplPath), content, data)
	if err != nil {
		return err
	}
	snake := data["ModuleSnake"].(string)
	rel := filepath.Join("internal", "bootstrap", "register_"+snake+"_grpc.go")
	skipped, err := writeProjectFileOrSkip(opts.ProjectDir, rel, string(out), opts.Force)
	if err != nil {
		return err
	}
	if skipped {
		fmt.Fprintf(os.Stderr, "info: skip existing %s (use --force)\n", rel)
	} else {
		fmt.Printf("  created %s\n", rel)
	}
	return writeGRPCWireSnippet(opts, snake, data["ModulePascal"].(string))
}

func ensureGRPCBootstrapRegistry(opts DBOptions, pascal string) error {
	funcName := "Register" + pascal + "GRPC"
	callLine := fmt.Sprintf("\t%s(s, infra)", funcName)
	grpcPath := filepath.Join(opts.ProjectDir, "internal", "bootstrap", "grpc.go")

	b, err := os.ReadFile(grpcPath)
	if err != nil {
		return fmt.Errorf("read grpc.go: %w (ensure project has internal/bootstrap/grpc.go)", err)
	}
	content := string(b)
	if strings.Contains(content, funcName+"(") {
		return nil
	}
	if !strings.Contains(content, grpcBootstrapMarker) {
		repaired, ok := insertRegistryMarker(content, grpcBootstrapMarker)
		if !ok {
			fmt.Fprintf(os.Stderr, "warn: grpc.go missing marker %q; skip %s\n", grpcBootstrapMarker, funcName)
			return nil
		}
		content = repaired
	}
	updated := strings.Replace(content, grpcBootstrapMarker+"\n", grpcBootstrapMarker+"\n"+callLine+"\n", 1)
	if updated == content {
		updated = strings.Replace(content, grpcBootstrapMarker, grpcBootstrapMarker+"\n"+callLine, 1)
	}
	if err := os.WriteFile(grpcPath, []byte(updated), 0644); err != nil {
		return err
	}
	fmt.Printf("  updated internal/bootstrap/grpc.go (+%s)\n", funcName)
	return nil
}
