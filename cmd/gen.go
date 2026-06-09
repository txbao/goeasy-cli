package cmd

import (
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/generator"

	"github.com/spf13/cobra"
)

var (
	genProjectDir   string
	genProtoFile    string
	genProtoURL     string
	genOpenAPIFile  string
	genOpenAPIDir   string
	genProtoSrcFile string
	genProtoSrcDir  string
	genSkipHTTP     bool
	genSkipGRPC     bool
	genWithProtoPB  bool
	genSkipApp          bool
	genAllowOverwrite   bool
	genMergeHTTP        bool
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate Go code from project contracts (OpenAPI, proto, etc.)",
}

var genProtoCmd = &cobra.Command{
	Use:   "proto",
	Short: "Run protoc on api/proto/*.proto (requires protoc + protoc-gen-go + protoc-gen-go-grpc)",
	RunE: func(cmd *cobra.Command, args []string) error {
		abs, err := filepath.Abs(genProjectDir)
		if err != nil {
			return err
		}
		return generator.GenerateProtoGo(generator.GenProtoOptions{
			ProjectDir: abs,
			ProtoFile:  genProtoFile,
			FromURL:    genProtoURL,
		})
	},
}

var genHTTPCmd = &cobra.Command{
	Use:   "http",
	Short: "Generate HTTP interface + app stubs from OpenAPI (contract-first)",
	RunE: func(cmd *cobra.Command, args []string) error {
		abs, err := filepath.Abs(genProjectDir)
		if err != nil {
			return err
		}
		opts := moduleOptsFromFlags(abs)
		if genMergeHTTP {
			opts.Force = false
		}
		return generator.GenerateHTTPFromOpenAPI(generator.GenHTTPOptions{
			ModuleOptions:  opts,
			OpenAPIFile:    genOpenAPIFile,
			OpenAPIDir:     genOpenAPIDir,
			WithApp:        !genSkipApp && !genMergeHTTP,
			AllowOverwrite: genAllowOverwrite,
			MergeHTTP:      genMergeHTTP,
		})
	},
}

var genGRPCCmd = &cobra.Command{
	Use:   "grpc",
	Short: "Generate gRPC interface stubs from .proto (requires existing app layer; run gen proto for *.pb.go)",
	RunE: func(cmd *cobra.Command, args []string) error {
		abs, err := filepath.Abs(genProjectDir)
		if err != nil {
			return err
		}
		return generator.GenerateGRPCFromProto(generator.GenGRPCOptions{
			ModuleOptions: moduleOptsFromFlags(abs),
			ProtoFile: genProtoSrcFile,
			ProtoDir:  genProtoSrcDir,
		})
	},
}

var genContractCmd = &cobra.Command{
	Use:   "contract",
	Short: "Batch generate from api/contracts/openapi + api/proto (contract-first workflow)",
	RunE: func(cmd *cobra.Command, args []string) error {
		abs, err := filepath.Abs(genProjectDir)
		if err != nil {
			return err
		}
		opts := moduleOptsFromFlags(abs)
		if genMergeHTTP {
			opts.Force = false
		}
		return generator.GenerateFromContracts(generator.GenContractOptions{
			GenHTTPOptions: generator.GenHTTPOptions{
				ModuleOptions:  opts,
				OpenAPIDir:     genOpenAPIDir,
				WithApp:        !genSkipApp && !genMergeHTTP,
				AllowOverwrite: genAllowOverwrite,
				MergeHTTP:      genMergeHTTP,
			},
			SkipHTTP:  genSkipHTTP,
			SkipGRPC:  genSkipGRPC,
			WithProto: genWithProtoPB,
		})
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.PersistentFlags().StringVar(&genProjectDir, "dir", ".", "Project root directory")
	bindAddSharedFlags(genCmd)

	genProtoCmd.Flags().StringVar(&genProtoFile, "file", "", "Single proto file (e.g. api/proto/sys_roles.proto); default all under api/proto")
	genProtoCmd.Flags().StringVar(&genProtoURL, "from-url", "", "Download remote .proto to api/proto/imported/ then run protoc")
	genCmd.AddCommand(genProtoCmd)

	genHTTPCmd.Flags().BoolVar(&addForce, "force", false, "Overwrite existing files")
	genHTTPCmd.Flags().StringVar(&genOpenAPIFile, "from", "", "OpenAPI file (e.g. api/contracts/openapi/sys_roles.openapi.yaml)")
	genHTTPCmd.Flags().StringVar(&genOpenAPIDir, "dir-api", "", "OpenAPI directory (default api/contracts/openapi)")
	genHTTPCmd.Flags().StringSliceVar(&addHTTPClients, "client", []string{}, "HTTP clients to generate (default: from OpenAPI paths)")
	genHTTPCmd.Flags().StringSliceVar(&addHTTPPublicClients, "public", nil, "Client(s) without auth middleware (e.g. h5)")
	genHTTPCmd.Flags().StringVar(&addHTTPDomain, "domain", "", "Bounded context override (alias: --group)")
	genHTTPCmd.Flags().StringVar(&addHTTPGroup, "group", "", "Deprecated alias of --domain")
	genHTTPCmd.Flags().StringVar(&addHTTPResource, "resource", "", "HTTP route resource override")
	genHTTPCmd.Flags().BoolVar(&genSkipApp, "skip-app", false, "Do not update app/domain stubs (only HTTP interface)")
	genHTTPCmd.Flags().BoolVar(&genAllowOverwrite, "allow-overwrite", false, "Allow --force to overwrite add db crud generated files")
	genHTTPCmd.Flags().BoolVar(&genMergeHTTP, "merge-http", false, "Incremental HTTP: skip app/domain, never overwrite existing HTTP files; generate non-CRUD OpenAPI routes")
	genCmd.AddCommand(genHTTPCmd)

	genGRPCCmd.Flags().BoolVar(&addForce, "force", false, "Overwrite existing files")
	genGRPCCmd.Flags().StringVar(&genProtoSrcFile, "from", "", "Proto source file (e.g. api/proto/sys_roles.proto)")
	genGRPCCmd.Flags().StringVar(&genProtoSrcDir, "dir-proto", "", "Proto directory (default api/proto, skips imported/)")
	genGRPCCmd.Flags().StringVar(&addHTTPGroup, "group", "", "Reserved for future route metadata")
	genGRPCCmd.Flags().StringVar(&addHTTPResource, "resource", "", "Reserved for future route metadata")
	genCmd.AddCommand(genGRPCCmd)

	genContractCmd.Flags().BoolVar(&addForce, "force", false, "Overwrite existing files")
	genContractCmd.Flags().StringVar(&genOpenAPIDir, "dir-api", "", "OpenAPI directory (default api/contracts/openapi)")
	genContractCmd.Flags().BoolVar(&genSkipHTTP, "skip-http", false, "Skip OpenAPI → HTTP generation")
	genContractCmd.Flags().BoolVar(&genSkipGRPC, "skip-grpc", false, "Skip proto → gRPC stub generation")
	genContractCmd.Flags().BoolVar(&genWithProtoPB, "with-proto", true, "Run protoc to generate *.pb.go before gRPC stubs")
	genContractCmd.Flags().BoolVar(&genSkipApp, "skip-app", false, "Do not update app/domain stubs")
	genContractCmd.Flags().BoolVar(&genAllowOverwrite, "allow-overwrite", false, "Allow --force to overwrite add db crud generated files")
	genContractCmd.Flags().BoolVar(&genMergeHTTP, "merge-http", false, "Incremental HTTP only (see gen http --merge-http)")
	genContractCmd.Flags().StringSliceVar(&addHTTPClients, "client", []string{}, "HTTP clients override")
	genContractCmd.Flags().StringSliceVar(&addHTTPPublicClients, "public", nil, "Client(s) without auth middleware (e.g. h5)")
	genContractCmd.Flags().StringVar(&addHTTPDomain, "domain", "", "Bounded context override (alias: --group)")
	genContractCmd.Flags().StringVar(&addHTTPGroup, "group", "", "Deprecated alias of --domain")
	genContractCmd.Flags().StringVar(&addHTTPResource, "resource", "", "HTTP route resource override")
	genCmd.AddCommand(genContractCmd)
}

func moduleOptsFromFlags(projectDir string) generator.ModuleOptions {
	domain := addHTTPDomain
	if domain == "" {
		domain = addHTTPGroup
	}
	return generator.ModuleOptions{
		ProjectDir:    projectDir,
		Force:         addForce,
		Clients:       append([]string(nil), addHTTPClients...),
		PublicClients: append([]string(nil), addHTTPPublicClients...),
		Domain:        domain,
		Group:         addHTTPGroup,
		Resource:      addHTTPResource,
		ConfigPath:    resolvedConfigPath(projectDir, addConfigPath),
		AppStyle:      addAppStyle,
	}
}
