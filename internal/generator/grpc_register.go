package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ensureGRPCRegistry 已废弃：模块 gRPC 改由 bootstrap/register_<module>_grpc.go 注册。
func ensureGRPCRegistry(projectDir, projectModule, snake string) error {
	return nil
}

func ensureGRPCRegisterFile(projectModule, projectDir string) error {
	registerPath := filepath.Join(projectDir, "internal", "interface", "grpc", "register.go")
	if _, err := os.Stat(registerPath); err == nil {
		return nil
	}
	return renderGRPCRegisterFile(projectDir, projectModule)
}

func renderGRPCRegisterFile(projectDir, projectModule string) error {
	goeasy := currentGoEasyModule()
	data := map[string]any{
		"ModuleName":   projectModule,
		"GoEasyModule": goeasy,
	}
	sub, err := fsSub("project")
	if err != nil {
		return err
	}
	tplPath := "internal/interface/grpc/register.go.tmpl"
	content, err := fs.ReadFile(sub, tplPath)
	if err != nil {
		return fmt.Errorf("read grpc register template: %w", err)
	}
	out, err := executeTemplate(filepath.Base(tplPath), content, data)
	if err != nil {
		return err
	}
	target := filepath.Join(projectDir, "internal", "interface", "grpc", "register.go")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(target, out, 0644); err != nil {
		return err
	}
	fmt.Printf("  created internal/interface/grpc/register.go\n")
	return nil
}
