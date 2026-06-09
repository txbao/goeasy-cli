package grpccmd

import (
	"fmt"
	"os"
	"strings"
)

// ResolveOptions grpc resolve 参数。
type ResolveOptions struct {
	CommonOptions
}

// Resolve 按 discovery（direct / etcd）解析逻辑服务名为 gRPC target。
func Resolve(opts ResolveOptions) error {
	target, cfg, err := resolveTarget(opts.CommonOptions)
	if err != nil {
		return err
	}
	service := trim(opts.Service)
	mode := strings.ToLower(strings.TrimSpace(cfg.Discovery.Mode))
	etcdOn := cfg.Discovery.Etcd.Enabled
	fmt.Fprintf(os.Stdout, "service=%s target=%s discovery_mode=%s etcd_enabled=%v\n",
		serviceOrDash(service, opts.Target), target, mode, etcdOn)
	if service != "" && etcdOn && strings.EqualFold(mode, "etcd") {
		prefix := strings.TrimSuffix(cfg.Discovery.Etcd.Prefix, "/")
		fmt.Fprintf(os.Stdout, "etcd_key=%s/%s\n", prefix, service)
	}
	return nil
}

func serviceOrDash(service, target string) string {
	if service != "" {
		return service
	}
	return "-"
}
