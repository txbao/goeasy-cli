package generator

type DBOptions struct {
	ModuleOptions
	ConfigPath    string
	Schema        string
	Table         string
	Tables        []string
	All           bool
	IncludePrefix string
	Exclude       []string
	ModuleName    string // --module 覆盖逻辑模块名
	WithProto     bool
	SkipProto     bool
	SkipGenProto  bool // 跳过 add db proto 后自动 protoc（默认 false = 尝试 gen proto）
	WithOpenAPI   bool
	SkipOpenAPI   bool
	SkipRegister  bool
	Group         string // HTTP 路由业务域（覆盖 config 推断）
	Resource      string // HTTP 路由资源名
}
