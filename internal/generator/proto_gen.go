package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenProtoOptions 从 api/proto/*.proto 生成 Go pb 代码（需本机安装 protoc 与插件）。
type GenProtoOptions struct {
	ProjectDir string
	// ProtoFile 指定单个 proto 相对路径（如 api/proto/sys_roles.proto）；为空则处理 api/proto 下全部 .proto。
	ProtoFile string
	// FromURL 下载远程 .proto 到 api/proto/imported/ 后再 protoc（http/https/file）。
	FromURL string
}

// GenerateProtoGo 调用 protoc 生成 *.pb.go / *_grpc.pb.go。
func GenerateProtoGo(opts GenProtoOptions) error {
	dir := opts.ProjectDir
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	module, err := readModulePath(abs)
	if err != nil {
		return err
	}
	if _, err := exec.LookPath("protoc"); err != nil {
		return fmt.Errorf("protoc not found in PATH: install https://github.com/protocolbuffers/protobuf/releases and run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest")
	}
	if opts.FromURL != "" {
		rel, err := FetchProtoFromURL(abs, opts.FromURL)
		if err != nil {
			return err
		}
		opts.ProtoFile = rel
	}
	files, err := resolveProtoFiles(abs, opts.ProtoFile)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .proto files under %s", filepath.Join(abs, "api", "proto"))
	}

	args := []string{
		fmt.Sprintf("--go_out=%s", abs),
		fmt.Sprintf("--go_opt=module=%s", module),
		fmt.Sprintf("--go-grpc_out=%s", abs),
		fmt.Sprintf("--go-grpc_opt=module=%s", module),
	}
	for _, f := range files {
		args = append(args, f)
	}
	cmd := exec.Command("protoc", args...)
	cmd.Dir = abs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("protoc %s\n", strings.Join(files, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("protoc: %w", err)
	}
	fmt.Println("ok: proto Go code generated; implement internal/interface/grpc/<module>/server.go RPC methods")
	return nil
}

func resolveProtoFiles(projectDir, single string) ([]string, error) {
	if single != "" {
		p := single
		if !filepath.IsAbs(p) {
			p = filepath.Join(projectDir, p)
		}
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("proto file: %w", err)
		}
		rel, err := filepath.Rel(projectDir, p)
		if err != nil {
			return nil, err
		}
		return []string{filepath.ToSlash(rel)}, nil
	}
	root := filepath.Join(projectDir, "api", "proto")
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "gen" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".proto") {
			rel, err := filepath.Rel(projectDir, path)
			if err != nil {
				return err
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	return out, err
}
