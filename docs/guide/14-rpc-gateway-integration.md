# 14 RPC Gateway 接入（DDD Lite）

跨服务 gRPC 的标准接入方式：业务层 **一行 Gateway 调用**，底层 **长连接 + ETCD/direct 发现**，符合 DDD Lite 分层（对标 go-zero `svcCtx.XxxRpc.GetById`）。

运行时能力见 [gRPC 与服务发现](../runtime/grpc-discovery.md)；联调 CLI 见 [12 跨服务 gRPC](12-grpc-cross-service.md)。

## 原则

| 层 | 职责 |
|----|------|
| `app/<module>/port/` | Gateway 接口 + 应用 DTO（无 pb；独立子包，避免 query import 父包循环） |
| `app/<module>/query` | `roles.GetByID(ctx, id)` — 业务只写这里 |
| `infrastructure/rpc/<remote>/` | ACL：pb stub + `HTTPInfra.RPC` 长连接 |
| `bootstrap` | `RPCClient(infra, "user")` 装配 Gateway；失败则启动中止 |

**禁止**：domain/app/interface 直接 `import` `api/proto/*.pb.go`。

## 快速开始

```bat
cd your-project

REM 一步：拉取对端契约并生成示范模块（--proto 可从 --from-url 自动推断，app_style 读 config）
goeasy-cli add rpcdemo --remote user --from-url ..\user\api\proto\sys_roles.proto

REM 对端仅有 sys_apis 时（如 demo1/demo2 联调；无需手写 --proto）：
goeasy-cli add rpcdemo --remote demo1 --from-url ..\demo1\api\proto\sys_apis.proto --force

REM 分步（已有 api/proto/gen/imported/<proto>/*.pb.go 时）：
goeasy-cli gen proto --from-url <对端 proto>
goeasy-cli add rpcdemo --remote user --proto sys_roles --force

REM 配置 discovery（调用方 configs/config.yaml）
REM    discovery.services.user: "127.0.0.1:28021"  （本地 dev；非 etcd 2379）

go mod tidy
go run ./cmd/service
```

| 参数 | 默认 | 说明 |
|------|------|------|
| `--remote` | `user` | 对端逻辑服务名（`discovery.services` 键） |
| `--proto` | `sys_roles` | 对端 gRPC 模块名；未指定时从 `--from-url` 文件名推断（如 `sys_apis.proto` → `sys_apis`） |
| `--from-url` | 空 | 对端 `.proto` 路径/URL；未提供时尝试 `../<remote>/api/proto/<proto>.proto` |
| `--app-style` | 读 config | 与 `add crud` 一致，默认 `service` |
| `--force` | false | 覆盖已有文件；切换 `--proto` 时会清理旧 `*_gateway.go` |

验证（`http.port` 见 `configs/config.yaml`，demo2 默认 `18091`）：

```bat
REM 业务健康（无需 JWT）
curl http://127.0.0.1:18091/health

REM GET：返回 proto 实体消息全部字段（如 sys_apis.SysApis）
curl http://127.0.0.1:18091/api/v1/admin/rpcdemo/1

REM POST：请求体字段与 proto Create<实体>Request 一致，响应为创建后的完整实体
curl -X POST http://127.0.0.1:18091/api/v1/admin/rpcdemo ^
  -H "Content-Type: application/json" ^
  -d "{\"name\":\"示例接口\",\"path\":\"/api/v1/demo\",\"method\":\"GET\",\"module\":\"demo\",\"description\":\"说明\",\"is_public\":0,\"status\":1,\"version\":1}"
```

字段由 CLI **自动解析对端 `.proto`**（实体 message → GET 响应；`Create<实体>Request` → POST 请求体），无需手写 DTO。

启动成功时应看到日志：`http listening on 0.0.0.0:18091`（框架**不打印**路由表）。

`add rpcdemo` 使用 **非阻塞 gRPC dial**（`RPCClientLazy`）：demo1 未启动时 demo2 仍可监听 HTTP；首次调用 rpcdemo 接口时才连接对端。

## 业务代码长什么样

**`app_style: service`（默认）** — `internal/app/rpcdemo/application.go`：

```go
func (a *Application) Get(ctx context.Context, id string) (*port.ApiView, error) {
	return a.sysApis.GetByID(ctx, id)
}

func (a *Application) Create(ctx context.Context, cmd CreateCommand) (*port.ApiView, error) {
	return a.sysApis.Create(ctx, &port.CreateInput{Name: cmd.Name, Path: cmd.Path /* ... */})
}
```

HTTP 层 `dto.go` 的 `ResponseDTO` / `CreateCommand` 字段与 proto 契约同步；`handler` 通过 `ToResponse(view)` 输出 JSON。

**`app_style: light_cqrs`** — `internal/app/rpcdemo/query/get.go`：

```go
func (h *Handler) Get(ctx context.Context, id string) (*port.ApiView, error) {
	return h.sysApis.GetByID(ctx, id)
}
```

`command/create.go` 调用 Gateway `Create`，返回完整 `port.*View`。

对比 go-zero：

```go
resp, err := l.svcCtx.CoreRpc.CoreAPIGetById(l.ctx, &rpcReq)
```

## 业务模块接入（add rpc 两步）

适用于 **已有业务模块**（如 `etcddemo`）内部调对端 gRPC，**不生成 HTTP 示范路由**。

```bat
REM 步骤 1：共享基础设施（与 consumer 无关，可多次 add 不同 proto）
goeasy-cli add rpc sys_roles --remote demo1 --from-url ..\demo1\api\proto\sys_roles.proto
goeasy-cli add rpc sys_apis  --remote demo1 --from-url ..\demo1\api\proto\sys_apis.proto

REM 步骤 2：绑定 consumer（可多个 consumer、可晚绑定）
goeasy-cli add rpc bind sys_roles --consumer etcddemo
goeasy-cli add rpc bind sys_apis  --consumer etcddemo

REM 步骤 3：手改 application.go 注入 Gateway，HTTP DTO 与 proto 解耦
```

| 步骤 | 产物 |
|------|------|
| 1 | `api/proto/imported/<proto>.proto`、`gen/imported/<proto>/*.pb.go` |
| 1 | `internal/infrastructure/rpc/<remote>/<proto>_gateway.go` |
| 1 | `internal/infrastructure/rpc/<remote>/port/<proto>.go`（**共享 Port**：Gateway 接口 + View/Input/ListResult） |
| 2 | `internal/app/<domain>/<resource>/port/<proto>.go`（type alias 到共享 port） |
| 2 | `register_<domain>.go` 内 `// goeasy-rpc-wire: <proto>` + `RPCClientLazy` + `New*Gateway`（`--wire` 默认 true） |

同一 `remote` 下多个 proto **共用** `RPCClientLazy(infra, "<remote>")`；bind 不重复 dial。

与 `add rpcdemo`：`rpcdemo` = HTTP↔proto 联调示范；`add rpc` + `bind` = 业务编排。

## 手动接入（非示范模块）

### 1. Port（app 层，独立子包 `port`）

```go
// internal/app/order/port/port.go
package port

type SysRolesGateway interface {
	GetByID(ctx context.Context, id string) (*RoleView, error)
	Create(ctx context.Context, in *CreateInput) (*RoleView, error)
}
```

`query` 子包只 import `app/order/port`，**不要** import 父包 `app/order`（否则与 `application.go` 形成 import cycle）。

### 2. ACL（infrastructure/rpc）

```go
// internal/infrastructure/rpc/user/sys_roles_gateway.go
func (g *SysRolesGateway) GetByID(ctx context.Context, id string) (*orderport.RoleView, error) {
	var out *sys_rolespb.SysRoles
	err := g.cli.Invoke(ctx, func(cctx context.Context) error {
		var e error
		out, e = g.stub.GetSysRoles(cctx, &sys_rolespb.GetSysRolesRequest{Id: id})
		return e
	})
	// 映射为 RoleView ...
}
```

### 3. bootstrap 装配

```go
cli, err := bootstrap.RPCClientLazy(infra, "user")
if err != nil {
    return fmt.Errorf("order: %w", err)
}
rolesGW := userpc.NewSysRolesGateway(cli)
application := orderapp.NewApplication(repo, rolesGW)
```

`RPCClientLazy` / `RPCClient` 来自 `internal/bootstrap/rpc.go`（`add rpcdemo` 自动创建）。**rpcdemo 示范**用 `RPCClientLazy`（非阻塞 dial）；手动模块若要求对端启动后才能起服务，可用 `RPCClient`。解析失败时 `RegisterRoutes` 返回 error，`main` 中 `log.Fatal` 退出。

## 配置要点

| 角色 | 关键配置 |
|------|----------|
| 提供方（demo1） | `grpc.enabled: true`；`grpc.addr` 监听 gRPC；`discovery.services.<app_name>` 或 ETCD 注册 **客户端可达** 地址 |
| 调用方（demo2） | `discovery.services.<remote>` **必填**（与 `--remote` 一致）；`grpc.enabled` 可 false |

**解析顺序**：`discovery.services.<remote>` 显式配置 **优先于** ETCD（本地 dev 直连）；未配 yaml 且 `mode: etcd` 时才查 ETCD。

本地 dev 推荐（demo1/demo2 联调）：

```yaml
discovery:
  mode: direct          # 或 mode: etcd 但须保证 etcd 可用且已注册
  etcd:
    enabled: false      # 无 etcd 集群时关闭
  services:
    demo1: "127.0.0.1:18083"   # 与 --remote demo1 对应；填对端 gRPC 端口，非 etcd 2379
```

常见误配：

- 把 ETCD 地址 `127.0.0.1:2379` 当作 gRPC 目标 → dial 失败或启动卡住
- `discovery.services` 未配 `--remote` 同名键 → `register http routes: rpcdemo: grpc client lazy "demo1": ...`
- 仅有 `/health`、`/healthz` 日志、无 `http listening` → `RegisterRPCDemo` 在解析/装配阶段失败，检查上述 discovery 配置

## 调用链

```text
interface/http  →  app/query  →  app/port (Gateway)
                                      ↑
                         infrastructure/rpc  →  HTTPInfra.RPC (长连接)
                                      →  discovery/etcd  →  远端 gRPC
```

## 排错

| 现象 | 处理 |
|------|------|
| 编译缺 `sys_roles` / `sys_apis` pb | `add rpcdemo --from-url <对端 proto>` 或先 `gen proto --from-url`；`--proto` 须与契约一致 |
| 残留旧 `sys_roles_gateway.go` | 切换 `--proto` 后 `add rpcdemo --force`（自动清理同 remote 下其它 `*_gateway.go`） |
| `Application` 仍是 `Queries()` 风格 | 项目 `codegen.app_style: service` 时需升级 CLI 并 `add rpcdemo --force` |
| dial 超时 | `grpc resolve` / `Test-NetConnection`；检查注册地址 |
| `HTTPInfra.RPC is nil` | 确保 `app.InitInfra()` 后注册路由；补 `bootstrap/rpc.go` |
| `curl` 连接失败 `(7) Could not connect` | HTTP 未监听：检查 `go run` 是否报错退出；确认 `http.port`（如 18091） |
| `/health` 可访问但以为「没有端口」 | `/health` 就在 `http.port` 上；日志中找 `http listening on` |
| rpcdemo 返回 500 / gRPC dial 错误 | demo1 未启动或 `discovery.services.demo1` 错误；先起 demo1 |
| 启动失败 `register http routes: rpcdemo: grpc client lazy` | `discovery.services.<remote>` 未配置或与 `--remote` 不一致 |
| 仅有 `/health` 无 `http listening`、curl 连接失败 | **discovery/etcd 未配好**：补 `services.demo1` 或修正 etcd；见上方「本地 dev 推荐」 |
| Gin 日志无 `rpcdemo` 路由 | 路由在 `RegisterRPCDemo` 之后注册；若卡在 discovery 解析则看不到；修复 discovery 后应出现 `GET /api/v1/admin/rpcdemo/:id` |
| GET/POST 仅返回 `id`/`active` 或 Create 只收 `id` | 旧版 rpcdemo 未按 proto 生成 DTO；升级 CLI 后 `add rpcdemo --remote <r> --from-url <proto> --force` |
| 业务模块需调对端 gRPC 但不想 add rpcdemo | `add rpc <proto>` + `add rpc bind --consumer <m>`；手改 `application.go` |
| bind 后 Gateway 变量未使用 | 手改 `NewApplication` 注入；wire 片段见 `bootstrap/snippets/` |
| `import cycle not allowed`（`app/rpcdemo`） | Port 须在 `app/<m>/port/` 子包；升级 CLI 后 `goeasy-cli add rpcdemo --force` 重生成 |
| ETCD 与 yaml 不一致 | **yaml `discovery.services` 优先**；本地 dev 建议 `mode: direct` + 显式 `services.<remote>` |

## 相关文档

- [07 DDD Lite 实践](07-ddd-lite-practices.md)
- [12 跨服务 gRPC](12-grpc-cross-service.md)
- [06 CLI 命令](06-goeasy-cli-commands.md)
