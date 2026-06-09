package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const mqdemoModuleName = "mqdemo"

// GenerateMQDemo 生成 NSQ 消息生产/消费示范模块（DDD Lite）。
func GenerateMQDemo(opts ModuleOptions) error {
	opts.ModuleName = mqdemoModuleName
	if moduleExists(opts.ProjectDir, moduleMetaByID(mqdemoModuleName)) && !opts.Force {
		fmt.Fprintf(os.Stderr, "info: module %q already exists, skipping (use --force to overwrite)\n", mqdemoModuleName)
		return nil
	}
	projectModule, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return err
	}
	data := toMap(BuildModuleData(mqdemoModuleName, projectModule))
	skip := []string{
		"internal/bootstrap/register_mqdemo.go",
		"internal/bootstrap/register_mqdemo_consumer.go",
		"cmd/consumer/main.go",
	}
	if err := renderScoped("mqdemo", opts.ProjectDir, nil, data, opts.Force, skip...); err != nil {
		return err
	}
	if err := renderMQDemoRegisterFiles(opts, data); err != nil {
		return err
	}
	if err := ensureMQDemoModulesRegistry(opts); err != nil {
		return err
	}
	return ensureConsumerMain(opts, data)
}

func ensureMQDemoModulesRegistry(opts ModuleOptions) error {
	const funcName = "RegisterMQDemo"
	callLine := moduleRegisterCallLine(funcName)

	modulesPath := filepath.Join(opts.ProjectDir, "internal", "bootstrap", "modules.go")
	if _, err := os.Stat(modulesPath); os.IsNotExist(err) {
		if err := renderModulesFile(opts); err != nil {
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

func renderMQDemoRegisterFiles(opts ModuleOptions, data map[string]any) error {
	sub, err := fsSub("mqdemo")
	if err != nil {
		return err
	}
	files := []string{
		"internal/bootstrap/register_mqdemo.go.tmpl",
		"internal/bootstrap/register_mqdemo_consumer.go.tmpl",
	}
	for _, tplPath := range files {
		targetRel := strings.TrimSuffix(tplPath, ".tmpl")
		targetPath := filepath.Join(opts.ProjectDir, targetRel)
		if !opts.Force {
			if _, err := os.Stat(targetPath); err == nil {
				fmt.Fprintf(os.Stderr, "info: %s exists, skipping (use --force)\n", targetRel)
				continue
			}
		}
		content, err := fs.ReadFile(sub, tplPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", tplPath, err)
		}
		out, err := executeTemplate(filepath.Base(tplPath), content, data)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, out, 0644); err != nil {
			return err
		}
		fmt.Printf("  created %s\n", filepath.ToSlash(targetRel))
	}
	return nil
}

func ensureConsumerMain(opts ModuleOptions, data map[string]any) error {
	sub, err := fsSub("mqdemo")
	if err != nil {
		return err
	}
	tplPath := "cmd/consumer/main.go.tmpl"
	targetRel := "cmd/consumer/main.go"
	targetPath := filepath.Join(opts.ProjectDir, targetRel)
	if !opts.Force {
		if _, err := os.Stat(targetPath); err == nil {
			fmt.Fprintf(os.Stderr, "info: %s exists, skipping (use --force)\n", targetRel)
			return nil
		}
	}
	content, err := fs.ReadFile(sub, tplPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", tplPath, err)
	}
	out, err := executeTemplate(filepath.Base(tplPath), content, data)
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
