package grpccmd

import (
	"encoding/json"
	"fmt"
)

// CallOptions grpc call 参数（Reflection + JSON，无需本地 .proto）。
type CallOptions struct {
	CommonOptions
	Method    string
	Data      string
	Plaintext bool
}

// Call 解析 target 后通过 grpcurl 调用 RPC。
func Call(opts CallOptions) error {
	if trim(opts.Method) == "" {
		return fmt.Errorf("--method is required (e.g. sys_roles.SysRolesService/GetSysRoles)")
	}
	data := opts.Data
	if data == "" {
		data = "{}"
	}
	if !json.Valid([]byte(data)) {
		return fmt.Errorf("invalid --data JSON")
	}
	target, _, err := resolveTarget(opts.CommonOptions)
	if err != nil {
		return err
	}
	return runGRPCURL(opts.Plaintext, target, "-d", data, opts.Method)
}
