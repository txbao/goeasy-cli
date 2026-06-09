# 05 goeasy 运行时

goeasy 是业务服务在**线上实际依赖**的 Go 库，负责配置、HTTP 引擎、日志、治理、观测与企业公共能力。

## 启动模型

```go
cfg := config.MustLoad("configs/config.yaml")
application := app.New(cfg)
application.RegisterHTTP(bootstrap.RegisterRoutes)
application.Run()
```

- `config.MustLoad`：加载 YAML，失败则 panic（开发期快速失败）
- `app.New`：根据配置初始化日志、中间件、可选 P1–P4 组件
- `RegisterHTTP`：注入 `func(engine *gin.Engine, infra app.HTTPInfra) error`；返回 error 时 `Run()` 中止启动（不监听 HTTP），`main` 中 `log.Fatal` 输出原因（如 gRPC 对端不可达）
- `infra.RPC`：长连接 gRPC 客户端池（`grpcx.Registry`），bootstrap 内 `RPCClient(infra, "逻辑服务名")` 使用
- 其余 `infra`：`infra.DB`、`infra.TablePrefix`；P0 起另有 `infra.Cache`、`infra.RedisKeyPrefix`、`infra.EntityCacheEnabled`、`infra.EntityCacheTTL`（实体 key：`{prefix}:{module}:id:{id}`，见 `goeasy/cachekey`）
- `Run`：先 `InitInfra`（含真实 sqlx 连接），再注册路由，最后监听信号优雅关闭

`database.enabled: true` 且 `orm: sqlx` 时，`database.Open` 建立 PostgreSQL/MySQL 连接池并支持 `Ping`。

## 配置结构（configs/config.yaml）

常见段落：

| 配置块 | 说明 |
|--------|------|
| `server` | 端口、模式、超时 |
| `logger` | 级别、输出 |
| `database` / `redis` / `cache` / `mq` | P1 基础设施；`cache.enabled` 控制仓储实体缓存 |
| `consumer_http` | `cmd/consumer` 健康探针端口（与 `http` 分离） |
| `observability.mq` | MQ 日志：`log_payload` / `log_payload_max_bytes` |
| `enterprise.jwt` / `enterprise.member_jwt` | 管理后台与 H5 两套 JWT |
| `governance` | 熔断、限流、重试 |
| `observability` | 追踪、指标、健康探针、审计 |

框架探针路径默认为 `/healthz`（与业务 `GET /health` 区分）。

## 能力分期（P0–P4）

| 阶段 | 包 | 说明 |
|------|-----|------|
| P0 | `app` `config` `logger` `httpx` `response` | 最小可运行 HTTP 服务 |
| P1 | `database` `cache` `mq` `grpcx` `discovery` `storage` `scheduler` | 基础设施接口 + Noop |
| P2 | `breaker` `limiter` `retry` `loadbalance` | 治理 |
| P3 | `trace` `metrics` `health` `audit` | 观测 |
| P4 | `errors` `validator` `pagination` `idgen` `contextx` `jwt` `casbin` `crypto` `apisign` `eventbus` | 企业公共能力 |

业务服务通过 `app.App` 字段访问已初始化的组件（按需启用配置即可）。

## HTTP 与统一响应

- 引擎：**Gin**
- 统一 JSON 封装：`goeasy/response`（与 `interface` 层业务 DTO 分离）
- 中间件扩展：`goeasy/httpx`

业务 Handler 建议只做：参数绑定 → 调 Application → 写响应。

## 错误码

`goeasy/errors` 提供 `CodedError`，便于 HTTP 层映射 `code` / `msg`。

领域错误在 `domain/<m>/errors.go` 定义，应用层转换为对外 DTO 或错误码。

## 边界提醒

goeasy **不包含** User、Order、Payment 等业务实体；这些只在业务 `internal/domain` 中实现。

JWT、Casbin 等提供**引擎级**封装，权限模型与策略仍属业务代码。

## 延伸阅读

- [16 运行时能力清单](16-runtime-capabilities.md)
- [goeasy 运行时规范](../spec/goeasy-runtime-spec.md)
- [实现路线图](../plans/goeasy-runtime-implementation.md)

## 专题

- 实体缓存与 Redis：[实体缓存](../runtime/entity-cache.md)
- HTTP 中间件：[HTTP 中间件](../runtime/http-middleware.md)
- **项目配置清单（demo3）**：[09 项目配置 P0/P1](09-project-config-p0-p1.md)
- **库表 OpenAPI/Proto**：[10 库表契约](10-db-openapi-proto.md)（goeasy-cli）
- **gRPC 内部调用**：[11 gRPC 内部调用](11-grpc-internal.md)（goeasy-cli）
- **运行时 gRPC**：[gRPC 与服务发现](../runtime/grpc-discovery.md)

## 下一步

[06 goeasy-cli 命令](06-goeasy-cli-commands.md)
