package cmd

import (
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/grpccmd"

	"github.com/spf13/cobra"
)

var (
	grpcDir        string
	grpcConfig     string
	grpcService    string
	grpcTarget     string
	grpcMethod     string
	grpcData       string
	grpcPlaintext  bool
	grpcListSvc    string
)

var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "gRPC discovery and reflection call utilities",
}

func init() {
	rootCmd.AddCommand(grpcCmd)
	grpcCmd.PersistentFlags().StringVar(&grpcDir, "dir", ".", "Project root directory")
	grpcCmd.PersistentFlags().StringVarP(&grpcConfig, "config", "f", "", "Config file path (default: configs/config.yaml or GOEASY_CONFIG)")
	grpcCmd.PersistentFlags().StringVar(&grpcService, "service", "", "Logical service name (app_name / discovery.services key)")
	grpcCmd.PersistentFlags().StringVar(&grpcTarget, "target", "", "Explicit gRPC host:port (skips discovery)")

	grpcResolveCmd := &cobra.Command{
		Use:   "resolve",
		Short: "Resolve logical service name to gRPC target (direct or etcd)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return grpccmd.Resolve(grpccmd.ResolveOptions{CommonOptions: grpcCommonOpts()})
		},
	}
	grpcCmd.AddCommand(grpcResolveCmd)

	grpcListCmd := &cobra.Command{
		Use:   "list",
		Short: "List gRPC services on remote server (requires grpcurl + server reflection)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return grpccmd.List(grpccmd.ListOptions{
				CommonOptions: grpcCommonOpts(),
				Plaintext:     grpcPlaintext, GRPCService: grpcListSvc,
			})
		},
	}
	grpcListCmd.Flags().BoolVar(&grpcPlaintext, "plaintext", true, "Use plaintext (no TLS)")
	grpcListCmd.Flags().StringVar(&grpcListSvc, "rpc-service", "", "Optional fully-qualified gRPC service to list methods")
	grpcCmd.AddCommand(grpcListCmd)

	grpcCallCmd := &cobra.Command{
		Use:   "call",
		Short: "Invoke RPC via grpcurl reflection (no local .proto required)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return grpccmd.Call(grpccmd.CallOptions{
				CommonOptions: grpcCommonOpts(),
				Method:        grpcMethod, Data: grpcData, Plaintext: grpcPlaintext,
			})
		},
	}
	grpcCallCmd.Flags().StringVar(&grpcMethod, "method", "", "RPC method (e.g. sys_roles.SysRolesService/GetSysRoles)")
	grpcCallCmd.Flags().StringVar(&grpcData, "data", "{}", "JSON request body")
	grpcCallCmd.Flags().BoolVar(&grpcPlaintext, "plaintext", true, "Use plaintext (no TLS)")
	_ = grpcCallCmd.MarkFlagRequired("method")
	grpcCmd.AddCommand(grpcCallCmd)
}

func grpcCommonOpts() grpccmd.CommonOptions {
	abs, err := filepath.Abs(grpcDir)
	if err != nil {
		abs = grpcDir
	}
	return grpccmd.CommonOptions{
		Dir:        abs,
		ConfigPath: resolvedConfigPath(abs, grpcConfig),
		Service:    grpcService,
		Target:     grpcTarget,
	}
}
