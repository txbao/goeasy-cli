package configpath

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultRel = "configs/config.yaml"
	EnvVar     = "GOEASY_CONFIG"
)

// Resolve 返回项目内配置文件相对或绝对路径。
// 优先级：CLI flag > GOEASY_CONFIG > DefaultRel。
func Resolve(projectDir, flagValue string) string {
	p := strings.TrimSpace(flagValue)
	if p == "" {
		p = strings.TrimSpace(os.Getenv(EnvVar))
	}
	if p == "" {
		p = DefaultRel
	}
	if filepath.IsAbs(p) {
		return p
	}
	if strings.TrimSpace(projectDir) == "" {
		return p
	}
	return filepath.Join(projectDir, p)
}
