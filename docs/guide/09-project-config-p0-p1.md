# 09 项目配置清单（P0 / P1）

面向 `goeasy new` / `add db crud` 生成的业务项目（如 demo3）。运行时能力见 [实体缓存](../runtime/entity-cache.md)、[HTTP 中间件](../runtime/http-middleware.md)。

## 1. 最小可运行（数据库 + HTTP）

```yaml
app_name: demo3
env: dev

http:
  host: 0.0.0.0
  port: 8080

database:
  enabled: true
  orm: sqlx
  driver: postgres
  dsn: "postgres://user:pass@127.0.0.1:5432/demo3?sslmode=disable"
  table_prefix: ""
```

```bat
goeasy-cli migrate up
go run ./cmd/service
```

## 2. P0：Redis 实体缓存（`add db crud` 前置）

执行 **`goeasy-cli add db crud`** 前，CLI 会校验（不连库）：

- `database.enabled: true` 且 `database.dsn` 非空
- `redis.enabled: true` 且 `redis.addr` 非空

`add crud`（名字驱动）**不需要**上述配置，也不连接数据库。

```yaml
redis:
  enabled: true
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  key_prefix: demo3      # 建议与 app_name 一致
  default_ttl: 168h

cache:
  enabled: true
  entity_ttl: 168h
```

生效条件：`EntityCacheEnabled == true`。Key 示例：`demo3:sys_roles:id:1`（**逻辑模块名** `sys_roles`）。

**已有旧 `repository_pg.go` 时**，需重新生成：

```bat
goeasy-cli add db crud --table sys_roles --force
```

并确认 `internal/bootstrap/register_sys_roles.go` 中：

```go
NewPGRepository(sqlxDB, infra.DBDriver, table, infra.Cache, infra.RedisKeyPrefix, "sys_roles", infra.EntityCacheEnabled, infra.EntityCacheTTL)
```

持久化路径（P1）：`internal/infrastructure/persistence/repository/<module>/`。

## 3. P1：CORS

```yaml
http:
  cors:
    enabled: true
    allow_origins: ["*"]   # 生产请改为具体域名
    allow_credentials: false
```

由 goeasy `httpx` 自动挂载，无需在 `main` 手写。

## 4. P1：双 JWT（管理后台 + H5）

```yaml
enterprise:
  jwt:
    enabled: true
    secret: "change-me-admin"
    issuer: demo3-admin
    expire_min: 120
  member_jwt:
    enabled: true
    secret: "change-me-member"
    issuer: demo3-member
    expire_min: 10080
  casbin:
    enabled: false
```

路由示例（`internal/interface/http/middleware`）：

```go
// add db crud 生成的 register_<module>.go（推荐，路径仍为 /api/v1/<module>）
api := engine.Group("/api/v1", middleware.AdminAuth(infra))

admin := engine.Group("/api/v1/admin", middleware.AdminAuth(infra))
h5 := engine.Group("/api/v1/h5", middleware.MemberAuth(infra))
```

`infra` 仅出现在 `bootstrap/register_*.go`，勿写在 `router.go`。

## 5. demo 对照检查表

| 项 | 检查 |
|----|------|
| `configs/config.yaml` 含 `redis` / `cache` / `http.cors` / `member_jwt` | □ |
| `go.mod` replace 本地 goeasy（monorepo） | □ |
| 迁移已 `migrate up` | □ |
| `repository` 在 `persistence/repository/sys_roles/` | □ |
| `register_*.go` import 路径含 `persistence/repository/` | □ |
| `add db crud --force` 后编译通过 | □ |

## 6. 常见错误

| 现象 | 处理 |
|------|------|
| 缓存不生效 | 确认 `redis.enabled` 与 `cache.enabled` 均为 true |
| `NewPGRepository` 参数过少 | `--force` 重生成 `repository_pg.go` 与 `register_*.go` |
| import 仍指向 `persistence/sys_roles` | 手改或 `--force` 重跑 `add db crud` |
| JWT 503 | 对应 `jwt.enabled` / `member_jwt.enabled` 未开 |

## 7. HTTP 路由分组（推荐）

在 `codegen` 段配置模块名前缀 → URL 业务域，生成三层路由 `/api/v1/{client}/{group}/{resource}`：

```yaml
codegen:
  create_omit: [id, created_at, updated_at, deleted_at]
  update_omit: [created_at, updated_at, deleted_at]
  group_prefixes:
    sys_: system    # sys_roles → GET /api/v1/admin/system/roles/1
    ord_: order
```

| 表/模块 | 管理后台 URL | HTTP 目录 |
|---------|-------------|-----------|
| `sys_roles` | `/api/v1/admin/system/roles` | `internal/interface/http/admin/system/roles/` |
| `sys_configs` | `/api/v1/admin/system/configs` | `internal/interface/http/admin/system/configs/` |

- **domain/app** 仍为 `internal/{domain,app}/sys_roles`（逻辑模块名不变）。
- 显式覆盖域/资源：`goeasy-cli add db crud --table sys_roles --domain system --resource roles --force`
- 无 `group_prefixes` 且无 `--group` 时保持扁平 `/api/v1/admin/sys_roles`（旧项目兼容）。

详见 [04 项目结构](04-project-structure.md)、[07 DDD Lite 实践](07-ddd-lite-practices.md)。

## 8. P2：OpenAPI / Proto（可选）

```bat
goeasy-cli add db openapi --table sys_roles --force
goeasy-cli add db proto --table sys_roles --force
```

或与 CRUD 一次生成：`add db crud --table sys_roles --force --with-openapi --with-proto`。

详见 [10 库表契约](10-db-openapi-proto.md)。

## 9. P3：gRPC + 服务发现（optional）

```yaml
grpc:
  enabled: true
  addr: "0.0.0.0:9001"

discovery:
  mode: direct
  services:
    demo3: "127.0.0.1:9001"
    sys_roles: "127.0.0.1:9001"
```

`main` 模板已包含 `application.RegisterGRPC(bootstrap.RegisterGRPCServers)`。

```bat
goeasy-cli add db proto --table sys_roles --force
goeasy-cli gen proto
```

实现 `internal/interface/grpc/sys_roles/server.go` 后，`grpcurl list` 应出现 `sys_roles.SysRolesService`（仅 Reflection 表示尚未注册业务 Service）。

详见 [11 gRPC 项目集成](11-grpc-internal.md)。

## 下一步

- [11 gRPC 项目集成](11-grpc-internal.md)
- [10 库表契约 OpenAPI/Proto](10-db-openapi-proto.md)
- [06 CLI 命令](06-goeasy-cli-commands.md)
- [04 项目结构](04-project-structure.md)
