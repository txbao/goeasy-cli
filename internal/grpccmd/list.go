package grpccmd

// ListOptions grpc list 参数。
type ListOptions struct {
	CommonOptions
	Plaintext bool
	// GRPCService 可选，如 sys_roles.SysRolesService；空则列出全部 Service。
	GRPCService string
}

// List 列出远端 gRPC 服务（Server Reflection）。
func List(opts ListOptions) error {
	target, _, err := resolveTarget(opts.CommonOptions)
	if err != nil {
		return err
	}
	if svc := trim(opts.GRPCService); svc != "" {
		return runGRPCURL(opts.Plaintext, target, "list", svc)
	}
	return runGRPCURL(opts.Plaintext, target, "list")
}
