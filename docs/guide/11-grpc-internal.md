# 11 gRPC 项目集成（P3）

本文描述**业务项目**如何把 `add db proto` 生成的契约接入 goeasy 运行时。框架侧配置与 `grpcx` 行为见 [gRPC 与服务发现](../runtime/grpc-discovery.md)。

## 完整流程（推荐顺序）

| 步骤 | 命令 / 操作 | 产出 |
|------|-------------|------|
| 1 | 配置 `grpc.enabled`、`grpc.addr` | 服务监听 gRPC 端口 |
| 2 | `goeasy-cli add db proto --table sys_roles` | `api/proto/sys_roles.proto`、gRPC 桩、`register_sys_roles_grpc.go`；**默认自动尝试** `gen proto` |
| 3 | `goeasy-cli gen proto`（若上步未成功） | `api/proto/gen/sys_roles/*.pb.go`（`go build` 依赖此包，需提交） |
| 4 | 实现 `server.go` 中 `SysRolesServiceServer` 各 RPC | 可调用业务逻辑 |
| 5 | `go run ./cmd/service` | `grpcurl list` 出现 `sys_roles.SysRolesService` |

`add db crud --with-proto` 等价于 CRUD + proto + gRPC 桩一步完成。

## 启动链

```text
main → app.RegisterGRPC(bootstrap.RegisterGRPCServers)  // 签名 func(s, HTTPInfra)
     → bootstrap.RegisterSysRolesGRPC(s, infra)  // 装配 Application
     → sys_roles.Register(s, app) → pb.RegisterSysRolesServiceServer(..., NewServer(app))
```

`main` 模板已包含 `RegisterGRPC`。`add db proto` 会：

1. 生成 `internal/bootstrap/register_<module>_grpc.go`（装配 Application 并调用 `<module>.Register`）
2. 在 `internal/bootstrap/grpc.go` 的 marker 下追加 `Register<Module>GRPC(s, infra)`

```text
// grpc bootstrap modules (goeasy add db proto appends below)
	RegisterSysRolesGRPC(s, infra)
```

`internal/interface/grpc/register.go` 由项目模板提供，供手扩展；**默认注册链走 bootstrap，不修改 `register.go`**。

## 生成物

| 命令 | 文件 |
|------|------|
| `add db proto --table sys_roles` | `api/proto/sys_roles.proto` |
| 同上 | `internal/interface/grpc/<domain>/<resource>/server.go`、`handlers.go`、`convert.go` |
| 同上 | `internal/bootstrap/register_sys_roles_grpc.go`（装配 Application 并注册 Service） |
| 同上 | `internal/bootstrap/snippets/sys_roles_grpc.md` |
| `gen proto` | `api/proto/gen/sys_roles/sys_roles.pb.go`、`*_grpc.pb.go` |

`handlers.go` 已实现 List/Get/Create/Update/Delete，按 `codegen.app_style` 委托 app 层（与 HTTP 一致：`service` 直调 `Application` 方法；`light_cqrs` 走 `command`/`query`）。`bootstrap/grpc.go` 在 marker 下追加 `Register<Module>GRPC(s, infra)`。

**前提：** 需先 `add db crud --table <m>` 生成 app 层，再 `add db proto`。若顺序颠倒，CLI 会拒绝生成 proto 并提示先执行 crud。

**幂等：** 再次执行 `add db proto` 时，已存在的 `*.proto` 会跳过，但会**补齐缺失的** `internal/interface/grpc/<domain>/<resource>/` 与 `register_<module>_grpc.go`（无需 `--force`）。覆盖已有桩文件请加 `--force`。

**module_id 与领域路径：** `api/proto/sys_roles.proto`、`register_sys_roles_grpc.go` 使用**模块 ID**（表名）；`internal/interface/grpc/system/roles/`、`internal/app/system/roles/` 由 `codegen.domains` 解析为 **domain/resource**（与 HTTP 一致）。`register_*_grpc.go` import 实现包路径为 `internal/interface/grpc/<domain>/<resource>`。

## goeasy-cli gen proto

需本机安装：

- [protoc](https://github.com/protocolbuffers/protobuf/releases)
- `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`

在项目根（含 `go.mod`）：

```bat
goeasy-cli gen proto
goeasy-cli gen proto --file api/proto/sys_roles.proto
```

等价于（使用 `go.mod` 的 module 作为 `--go_opt=module=`）：

```bat
protoc --go_out=. --go_opt=module=<your/module> --go-grpc_out=. --go-grpc_opt=module=<your/module> api/proto/sys_roles.proto
```

## 实现与扩展

- 生成代码按 `app_style` 委托 app 层（`service`：`app.Get`/`Create` 等；`light_cqrs`：`Queries()`/`Commands()`），与 HTTP 一致
- 手改业务逻辑优先改 `internal/app/<module>`，gRPC/HTTP 同步生效
- 复杂校验可在 `handlers.go` 增加，或扩展 `convert.go`

详见各模块 `internal/bootstrap/snippets/<module>_grpc.md`。

## Server Reflection（grpcurl）

goeasy `grpcx.NewServer` **默认注册** reflection，`grpcurl list <addr>` 不应报 “does not support reflection”。

| grpcurl 结果 | 含义 | 处理 |
|--------------|------|------|
| 仅有 `grpc.reflection.*` | 端口通，**未注册业务 Service** | 执行 `gen proto`；确认 `bootstrap/grpc.go` 已调用 `Register<Module>GRPC`；见下方排障 |
| 有 `sys_roles.SysRolesService` | 注册成功 | `grpcurl list sys_roles.SysRolesService` 查看方法 |

示例：

```bat
grpcurl -plaintext localhost:9001 list
grpcurl -plaintext localhost:9001 list sys_roles.SysRolesService
```

## 调用其它服务

**联调（无需本地 proto）：** [12 跨服务 gRPC](12-grpc-cross-service.md) — `goeasy-cli grpc resolve/list/call --service <app_name>`。

**业务代码（推荐）：** [14 RPC Gateway 接入](14-rpc-gateway-integration.md) — `goeasy-cli add rpcdemo --remote user`，Query 内 `roles.GetByID(ctx, id)`。

低级 dial 示例见 `internal/infrastructure/client/grpc_client.go`。

## ETCD（optional）

```yaml
discovery:
  mode: etcd
  etcd:
    enabled: true
    endpoints: ["127.0.0.1:2379"]
    lease_ttl_sec: 30
    advertise_addr: "127.0.0.1:9001"   # 集群部署填可达 IP:port
    prefix: /goeasy/services
  services:
    sys_roles: "127.0.0.1:9001"
```

服务启动后向 etcd 注册 `app_name -> 可达 gRPC 地址`（KeepAlive 保活租约，优雅退出时撤销）。

验证（服务运行中）：

```bat
etcdctl get /goeasy/services --prefix
etcdctl lease list
```

| 现象 | 处理 |
|------|------|
| `lease list` 为 0 | 确认 `mode: etcd` 且 `etcd.enabled: true`；看启动日志 `discovery registered` 或错误 |
| key 存在但客户端连不上 | 检查 `advertise_addr` 或 `services[app_name]` 是否为集群可达地址，勿用 `0.0.0.0` |
| 查不到 key | key 为 `{prefix}/{app_name}`，与 `app_name` 配置一致 |

## 排障：有 proto / pb，无 gRPC 桩或编译失败

| 现象 | 原因 | 处理 |
|------|------|------|
| `api/proto/sys_roles.proto` 存在，无 `server.go` | 曾先跑 `add db proto` 后跑 crud，或旧版 CLI 在 proto 已存在时跳过 gRPC 桩 | 先 `add db crud --table sys_roles`，再 `add db proto --table sys_roles`（会自动补齐桩） |
| `no required module .../internal/interface/grpc/sys_roles` | 旧版 `register_*_grpc.go` import 用了 module_id 路径，桩在 `grpc/system/roles` | 升级 CLI 后 `add db proto --table sys_roles --force`，或手改 import 为 `.../internal/interface/grpc/system/roles` |
| 有 `api/proto/MODULE.proto` | 旧版 `add proto` 路径占位符未替换 | 删除后 `add proto sys_roles` 或改用 `add db proto` |
| 有 `*.pb.go`，`grpcurl list` 仍无 Service | 未生成或未注册 gRPC 桩 | 检查 `register_<module>_grpc.go` 与 `bootstrap/grpc.go` marker |
| 需覆盖已改动的桩 | — | `add db proto --table <m> --force` |
| `no required module .../command` | gRPC handlers 与 `app_style` 不一致（如 `service` 项目生成了 `light_cqrs` 桩） | 升级 CLI 后 `add db proto --table <m> --force`；切换风格后先 `add db crud --force` 再 `add db proto --force` |

## 集成测试（模板）

```bat
set GOEASY_CONFIG=configs/config.yaml
set GOEASY_GRPC_PEER_SERVICE=sys_roles
go test -tags=integration ./test/integration/...
```

## 相关文档

- [12 跨服务 gRPC 调用](12-grpc-cross-service.md)
- [10 库表契约](10-db-openapi-proto.md)
- [09 项目配置](09-project-config-p0-p1.md)
- [gRPC 与服务发现](../runtime/grpc-discovery.md)
