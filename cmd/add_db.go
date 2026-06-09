package cmd

import (
	"path/filepath"
	"strings"

	"github.com/txbao/goeasy-cli/internal/generator"

	"github.com/spf13/cobra"
)

var (
	dbConfigPath    string
	dbSchema        string
	dbTable         string
	dbTables        string
	dbAll           bool
	dbIncludePrefix string
	dbExclude       string
	dbModuleName    string
	dbWithProto     bool
	dbSkipProto     bool
	dbWithOpenAPI   bool
	dbSkipOpenAPI   bool
	dbSkipRegister  bool
	dbSkipGenProto  bool
	dbHTTPClients   []string
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Generate from database schema (sqlx + goqu)",
}

func init() {
	addCmd.AddCommand(dbCmd)
	dbCmd.PersistentFlags().StringVarP(&dbConfigPath, "config", "f", "", "Config file path (default: configs/config.yaml or GOEASY_CONFIG)")
	dbCmd.PersistentFlags().StringVar(&dbSchema, "schema", "public", "PG: schema (default public); MySQL: database name")
	dbCmd.PersistentFlags().StringVar(&dbTable, "table", "", "Single table (logical name, uses table_prefix)")
	dbCmd.PersistentFlags().StringVar(&dbTables, "tables", "", "Comma-separated tables")
	dbCmd.PersistentFlags().BoolVar(&dbAll, "all", false, "All tables in schema (excludes _sqlx_migrations)")
	dbCmd.PersistentFlags().StringVar(&dbIncludePrefix, "include-prefix", "", "Filter tables by prefix when using --all")
	dbCmd.PersistentFlags().StringVar(&dbExclude, "exclude", "", "Extra excluded tables/patterns (comma-separated)")
	dbCmd.PersistentFlags().StringVar(&dbModuleName, "module", "", "Override module name (default from table name)")
	dbCmd.PersistentFlags().BoolVar(&dbWithProto, "with-proto", false, "Also generate proto (db crud / db all)")
	dbCmd.PersistentFlags().BoolVar(&dbSkipProto, "skip-proto", false, "Skip proto in db all")
	dbCmd.PersistentFlags().BoolVar(&dbWithOpenAPI, "with-openapi", false, "Also generate OpenAPI (db crud / db all)")
	dbCmd.PersistentFlags().BoolVar(&dbSkipOpenAPI, "skip-openapi", false, "Skip OpenAPI in db all")
	dbCmd.PersistentFlags().BoolVar(&dbSkipRegister, "skip-register", false, "Do not write register_*.go / modules.go")
	dbCmd.PersistentFlags().BoolVar(&dbSkipGenProto, "skip-gen-proto", false, "Skip auto protoc after add db proto (default: try gen proto)")
	dbCmd.PersistentFlags().StringSliceVar(&dbHTTPClients, "client", []string{"admin"}, "HTTP client surface(s): admin (default), h5, app")

	dbCmd.AddCommand(&cobra.Command{
		Use:   "module",
		Short: "Generate module + persistence from table(s) (no CRUD HTTP overlay)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateDBModule(dbOpts())
		},
	})
	dbCmd.AddCommand(&cobra.Command{
		Use:   "crud",
		Short: "Generate module CRUD from database table(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateDBCRUD(dbOpts())
		},
	})
	dbCmd.AddCommand(&cobra.Command{
		Use:   "proto",
		Short: "Generate proto from database table(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			o := dbOpts()
			return generator.GenerateDBProto(o)
		},
	})
	dbCmd.AddCommand(&cobra.Command{
		Use:   "openapi",
		Short: "Generate OpenAPI 3 YAML from database table(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator.GenerateDBOpenAPI(dbOpts())
		},
	})
	dbCmd.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "Generate crud (+ optional proto) for matched tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			o := dbOpts()
			o.All = true
			return generator.GenerateDBAll(o)
		},
	})
}

func dbOpts() generator.DBOptions {
	abs, err := filepath.Abs(addProjectDir)
	if err != nil {
		abs = addProjectDir
	}
	var tables []string
	if dbTables != "" {
		for _, t := range strings.Split(dbTables, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tables = append(tables, t)
			}
		}
	}
	var exclude []string
	if dbExclude != "" {
		for _, t := range strings.Split(dbExclude, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				exclude = append(exclude, t)
			}
		}
	}
	domain := addHTTPDomain
	if domain == "" {
		domain = addHTTPGroup
	}
	cfgPath := resolvedConfigPath(abs, dbConfigPath)
	if cfgPath == "" {
		cfgPath = resolvedConfigPath(abs, addConfigPath)
	}
	return generator.DBOptions{
		ModuleOptions: generator.ModuleOptions{
			ProjectDir:    abs,
			Force:         addForce,
			Clients:       append([]string(nil), dbHTTPClients...),
			PublicClients: append([]string(nil), addHTTPPublicClients...),
			Domain:        domain,
			Group:         addHTTPGroup,
			Resource:      addHTTPResource,
			ConfigPath:    cfgPath,
			AppStyle:      addAppStyle,
		},
		ConfigPath:    cfgPath,
		Schema:        dbSchema,
		Table:         dbTable,
		Tables:        tables,
		All:           dbAll,
		IncludePrefix: dbIncludePrefix,
		Exclude:       exclude,
		ModuleName:    dbModuleName,
		WithProto:     dbWithProto,
		SkipProto:     dbSkipProto,
		WithOpenAPI:   dbWithOpenAPI,
		SkipOpenAPI:   dbSkipOpenAPI,
		SkipRegister:  dbSkipRegister,
		SkipGenProto:  dbSkipGenProto,
	}
}
