package grpccmd

import (
	"context"
	"path/filepath"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/discovery"
	"github.com/txbao/goeasy/grpcx"
)

// CommonOptions grpc 子命令公共参数。
type CommonOptions struct {
	Dir        string
	ConfigPath string
	Service    string
	Target     string // 显式 host:port，非空时跳过服务发现
}

func loadConfig(opts CommonOptions) (*config.Config, string, error) {
	cfgPath := opts.ConfigPath
	if !filepath.IsAbs(cfgPath) {
		cfgPath = filepath.Join(opts.Dir, cfgPath)
	}
	return config.MustLoad(cfgPath), cfgPath, nil
}

func resolveTarget(opts CommonOptions) (string, *config.Config, error) {
	if t := trim(opts.Target); t != "" {
		cfg, _, err := loadConfig(opts)
		if err != nil {
			return "", nil, err
		}
		return t, cfg, nil
	}
	service := trim(opts.Service)
	if service == "" {
		return "", nil, errServiceOrTargetRequired()
	}
	cfg, _, err := loadConfig(opts)
	if err != nil {
		return "", nil, err
	}
	reg := discovery.NewRegistry(cfg)
	target, err := grpcx.ResolveService(context.Background(), cfg, reg, service)
	if err != nil {
		return "", nil, err
	}
	return target, cfg, nil
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 {
		c := s[len(s)-1]
		if c != ' ' && c != '\t' {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}
