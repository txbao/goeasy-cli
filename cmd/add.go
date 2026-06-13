package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/generator"

	"github.com/spf13/cobra"
)

var (
	addForce             bool
	addProjectDir        string
	addCrudWithMigration bool
	addCrudWithAudit     bool
	addHTTPClients       []string
	addHTTPPublicClients []string
	addHTTPDomain        string
	addHTTPGroup         string
	addHTTPResource      string
	addRemoteService     string
	addRPCProtoModule    string
	addRPCProtoFromURL   string
	addRPCSkipFetchProto bool
	addRPCMethods        string
	addRPCConsumer       string
	addRPCWire           bool
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add module, crud, db, repository, proto, event, mqdemo, rpcdemo, or rpc",
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.PersistentFlags().BoolVar(&addForce, "force", false, "Overwrite existing files")
	addCmd.PersistentFlags().StringVar(&addProjectDir, "dir", ".", "Existing project directory")
	addCmd.PersistentFlags().StringSliceVar(&addHTTPClients, "client", []string{"admin"}, "HTTP client surface(s): admin (default), h5, app")
	addCmd.PersistentFlags().StringSliceVar(&addHTTPPublicClients, "public", nil, "Client(s) without auth middleware (e.g. h5); must also appear in --client; admin is not allowed")
	addCmd.PersistentFlags().StringVar(&addHTTPDomain, "domain", "", "Bounded context / URL domain (e.g. system → /api/v1/admin/system/roles)")
	addCmd.PersistentFlags().StringVar(&addHTTPGroup, "group", "", "Deprecated alias of --domain")
	addCmd.PersistentFlags().StringVar(&addHTTPResource, "resource", "", "Resource name / Go package (default: from codegen.domains or table prefix)")
	bindAddSharedFlags(addCmd)

	addCmd.AddCommand(&cobra.Command{
		Use:   "module [name]",
		Short: "Generate full DDD module skeleton (no CRUD HTTP overlay)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateModule(genOpts(args[0], false))
		},
	})

	crudCmd := &cobra.Command{
		Use:   "crud [name]",
		Short: "Generate module + CRUD HTTP + repository_pg (recommended)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateCRUD(genOpts(args[0], addCrudWithMigration))
		},
	}
	crudCmd.Flags().BoolVar(&addCrudWithMigration, "with-migration", false, "Also create migrations/<driver>/ create_<module>_table up/down SQL")
	crudCmd.Flags().BoolVar(&addCrudWithAudit, "audit", false, "Inject audit.Recorder into Application and generate operation-log stubs")
	addCmd.AddCommand(crudCmd)

	addCmd.AddCommand(&cobra.Command{
		Use:        "repository [name]",
		Short:      "Add PostgreSQL repository stub (prefer add crud)",
		Deprecated: "use add crud which includes repository_pg; this command is for adding PG stub to an existing module only",
		Args:       cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "info: add repository is deprecated for new modules; use add crud\n")
			return generator.GenerateRepository(genOpts(args[0], false))
		},
	})
	addCmd.AddCommand(&cobra.Command{
		Use:   "proto [name]",
		Short: "Generate api/proto definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateProto(genOpts(args[0], false))
		},
	})
	addCmd.AddCommand(&cobra.Command{
		Use:   "event [name]",
		Short: "Generate domain event and publisher stub",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateEvent(genOpts(args[0], false))
		},
	})
	addCmd.AddCommand(&cobra.Command{
		Use:   "mqdemo",
		Short: "Generate NSQ message publish/consume demo module (DDD Lite)",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(addProjectDir)
			if err != nil {
				abs = addProjectDir
			}
			return generator.GenerateMQDemo(generator.ModuleOptions{
				ProjectDir: abs,
				Force:      addForce,
			})
		},
	})
	rpcdemoCmd := &cobra.Command{
		Use:   "rpcdemo",
		Short: "Generate cross-service gRPC Gateway demo module (DDD Lite)",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(addProjectDir)
			if err != nil {
				abs = addProjectDir
			}
			return generator.GenerateRPCDemo(generator.RPCDemoOptions{
				ProjectDir:     abs,
				RemoteService:  addRemoteService,
				ProtoModule:    addRPCProtoModule,
				FromURL:        addRPCProtoFromURL,
				SkipFetchProto: addRPCSkipFetchProto,
				ConfigPath:     resolvedConfigPath(abs, addConfigPath),
				AppStyle:       addAppStyle,
				Force:          addForce,
			})
		},
	}
	rpcdemoCmd.Flags().StringVar(&addRemoteService, "remote", "user", "Remote logical service name (app_name / discovery key)")
	rpcdemoCmd.Flags().StringVar(&addRPCProtoModule, "proto", "", "Remote gRPC proto module name (default: sys_roles, or inferred from --from-url basename)")
	rpcdemoCmd.Flags().StringVar(&addRPCProtoFromURL, "from-url", "", "Remote .proto path or URL; auto-fetch to api/proto/imported/ then protoc")
	rpcdemoCmd.Flags().BoolVar(&addRPCSkipFetchProto, "skip-fetch-proto", false, "Skip auto fetch/gen proto (caller already has api/proto/gen/imported/<proto>/*.pb.go)")
	addCmd.AddCommand(rpcdemoCmd)

	rpcCmd := &cobra.Command{
		Use:   "rpc [proto-module]",
		Short: "Fetch remote proto and generate gateway + shared port (step 1)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(addProjectDir)
			if err != nil {
				abs = addProjectDir
			}
			return generator.GenerateRPCClient(generator.RPCClientOptions{
				ProjectDir:     abs,
				RemoteService:  addRemoteService,
				ProtoModule:    args[0],
				FromURL:        addRPCProtoFromURL,
				SkipFetchProto: addRPCSkipFetchProto,
				Methods:        addRPCMethods,
				Force:          addForce,
			})
		},
	}
	rpcCmd.Flags().StringVar(&addRemoteService, "remote", "user", "Remote logical service name (app_name / discovery key)")
	rpcCmd.Flags().StringVar(&addRPCProtoFromURL, "from-url", "", "Remote .proto path or URL; auto-fetch to api/proto/imported/ then protoc")
	rpcCmd.Flags().BoolVar(&addRPCSkipFetchProto, "skip-fetch-proto", false, "Skip auto fetch/gen proto (caller already has api/proto/gen/imported/<proto>/*.pb.go)")
	rpcCmd.Flags().StringVar(&addRPCMethods, "methods", "all", "Comma-separated RPC methods: Get,Create,Update,Delete,List; or all")

	rpcBindCmd := &cobra.Command{
		Use:   "bind [proto-module]",
		Short: "Bind shared RPC port to a consumer module + optional register wire (step 2)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(addProjectDir)
			if err != nil {
				abs = addProjectDir
			}
			return generator.GenerateRPCClientBind(generator.RPCClientBindOptions{
				ProjectDir:    abs,
				ProtoModule:   args[0],
				Consumer:      addRPCConsumer,
				RemoteService: addRemoteService,
				Wire:          addRPCWire,
				Force:         addForce,
				ConfigPath:    resolvedConfigPath(abs, addConfigPath),
			})
		},
	}
	rpcBindCmd.Flags().StringVar(&addRPCConsumer, "consumer", "", "Consumer module ID (required), e.g. etcddemo")
	rpcBindCmd.Flags().StringVar(&addRemoteService, "remote", "", "Remote service name (optional if gateway already exists)")
	rpcBindCmd.Flags().BoolVar(&addRPCWire, "wire", true, "Idempotently wire RPCClientLazy + Gateway into register_<domain>.go")
	_ = rpcBindCmd.MarkFlagRequired("consumer")
	rpcCmd.AddCommand(rpcBindCmd)
	addCmd.AddCommand(rpcCmd)

	addCmd.AddCommand(&cobra.Command{
		Use:        "aggregate [name]",
		Short:      "Generate aggregate subset (deprecated)",
		Hidden:     true,
		Deprecated: "use add module or add crud instead",
		Args:       cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateAggregate(genOpts(args[0], false))
		},
	})
}

func genOpts(name string, withMigration bool) generator.ModuleOptions {
	abs, err := filepath.Abs(addProjectDir)
	if err != nil {
		abs = addProjectDir
	}
	domain := addHTTPDomain
	if domain == "" {
		domain = addHTTPGroup
	}
	return generator.ModuleOptions{
		ProjectDir:    abs,
		ModuleName:    name,
		Force:         addForce,
		WithMigration: withMigration,
		WithAudit:     addCrudWithAudit,
		Clients:       append([]string(nil), addHTTPClients...),
		PublicClients: append([]string(nil), addHTTPPublicClients...),
		Domain:        domain,
		Group:         addHTTPGroup,
		Resource:      addHTTPResource,
		ConfigPath:    resolvedConfigPath(abs, addConfigPath),
		AppStyle:      addAppStyle,
	}
}
