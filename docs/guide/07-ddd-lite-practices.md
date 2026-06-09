# 07 DDD Lite 实践

本框架采用 **DDD Lite**：保留分层与边界，不强制事件溯源、复杂 CQRS 基础设施。

## 四层职责

| 层 | 目录 | 做什么 | 不做什么 |
|----|------|--------|----------|
| 领域 | `domain/` | 规则、实体行为、聚合不变量 | HTTP、SQL、第三方 SDK |
| 应用 | `app/` | 用例编排、事务边界、DTO | 直接写 SQL |
| 接口 | `interface/` | 协议适配、参数校验 | 业务规则 |
| 基础设施 | `infrastructure/` | 仓储实现、MQ、缓存客户端 | 领域规则 |

## 实体要有行为

反例：到处改 `status` 字段的贫血模型。

正例（health 示范）：

```go
func (s *ServiceHealth) MarkHealthy(message string) error {
    if message == "" {
        return ErrInvalidMessage
    }
    s.state = StateUp
    s.message = message
    return nil
}
```

对外只暴露方法或 `ToStatus()` 快照，不暴露可随意修改的字段。

## 聚合根

`aggregate.go` 负责维护聚合内一致性（例如根实体 + 子集合）。跨聚合修改通过应用服务或领域事件协调，不在一个 Repository 里隐式改两个聚合。

脚手架 `health` 模块用技术场景演示聚合，**不会**生成 User/Order 等业务模板。

## 库表驱动生成（add db）

迁移落库后，可用 `goeasy add db crud` 从 PostgreSQL 自省列类型，生成 entity、`repository_pg`（goqu）与 HTTP CRUD。系统列（`id`、时间戳、软删）按 `configs/config.yaml` 中 `codegen` 段裁剪，避免 Create/Update 请求携带不可写字段。

复杂表（RBAC、多租户、JSONB 规则）建议生成后再手改仓储。规划见 [库表驱动](../../docs/plans/goeasy-cli-database-first-codegen.md)，命令见 [06](06-goeasy-cli-commands.md#add-db库表驱动postgres--mysql)。

## 领域服务

当规则涉及多个实体或不适合放在单个实体上时，使用 `domain/<m>/service.go`。

应用层调用顺序建议：

```text
Application → DomainService / Entity → Repository
```

读用例（Query）可以只读仓储返回 DTO，但写用例必须经过领域行为。

## 应用层风格（app_style）

默认 **`service`**：`Application` 上直接暴露 `Create/Get/List/Update/Delete`，文件更少，适合 CRUD 业务。

可选 **`light_cqrs`**：`command/` + `query/` 子包；`health` 示范模块固定为此风格。

**`full_cqrs`**：CLI 不生成，需手工演进读模型与投影。详见 [18 应用层风格](18-app-style.md)。

### service（默认）

```text
app/<domain>/<resource>/
├── application.go
└── dto.go
```

Handler 调用：`h.app.Create(...)` / `h.app.Get(...)` / `h.app.List(...)`。

### light_cqrs（可选）

```text
app/<domain>/<resource>/
├── application.go
├── command/
├── query/
├── list.go
├── port/
└── dto.go
```

- **Command**：创建、更新、状态变更
- **Query**：列表、详情

Handler 调用：`h.app.Commands().Xxx()` / `h.app.Queries().Get()`（List 可走 `h.app.List()` 门面）。

Handler（HTTP）只依赖 `Application`，不直接 new Repository。

### HTTP 按客户端 + 业务域分包

```text
internal/interface/http/
├── admin/<group>/<resource>/   # /api/v1/admin/<group>/<resource>  + AdminAuth
├── h5/<group>/<resource>/      # /api/v1/h5/<group>/<resource>      + MemberAuth
└── middleware/
```

示例（`sys_roles` + `codegen.domains.system`）：

| 目录 | URL |
|------|-----|
| `admin/system/roles/` | `GET /api/v1/admin/system/roles/1` |
| `h5/system/roles/` | `GET /api/v1/h5/system/roles/1` |

- HTTP handler 包名为 **resource**（如 `roles`）；仍委托 `app/sys_roles` Application。
- 各端 `ResponseDTO` 在 `http/<client>/<group>/<resource>/dto.go`；`app/<module>/dto.go` 为应用层共用。
- 生成：`goeasy-cli add db crud --table sys_roles`（默认 `app_style: service`，读取 `codegen.domains`）；`--app-style light_cqrs` 可切换；`--domain system --resource roles` 可显式指定；`--client h5` 追加 H5 端。
- 代码路径：`internal/domain/system/roles`、`internal/app/system/roles`、`internal/infrastructure/system/persistence/roles`。

## 依赖注入（bootstrap）

```text
internal/bootstrap/
├── wire.go                 # registerHealth + registerAllModules
├── modules.go              # 各业务模块一行注册（CLI 自动追加）
├── register_health.go
└── register_<domain>.go    # BC 级 DI + 路由（CLI 生成）
```

`wire.go` 保持稳定；`goeasy add module/crud` 生成 `register_<module>.go` 并更新 `modules.go`。

## 新增业务模块推荐流程

**库表驱动：**

1. `goeasy-cli add db crud --table <name>`
2. 实现 `domain/<name>` 业务逻辑
3. `go mod tidy` → `goeasy-cli migrate up`

**契约驱动（先 API/Proto）：**

1. 编写 `api/contracts/openapi/<name>.openapi.yaml`（及可选 `api/proto/<name>.proto`）
2. `goeasy-cli gen http --from api/contracts/openapi/<name>.openapi.yaml` → `goeasy-cli gen contract`
3. 手改 `domain` / `repository_pg`；落库后 `migrate up`

详见 [15 契约驱动生成](15-contract-first.md)。

## 业务模块 / 事件 / MQ 放哪一层

GoEasy 采用 **层优先 + domain 布局**（非 `internal/order/interfaces|application` 纵向切分）。对照：

| 能力 | 领域 `domain/` | 应用 `app/` | 接口 `interface/` | 基础设施 `infrastructure/` |
|------|----------------|-------------|-------------------|---------------------------|
| CRUD 业务模块 | `<bc>/<resource>/` entity、aggregate、repository 接口 | `<bc>/<resource>/` Application、dto、command/query | `http/<client>/<bc>/<resource>/` handler、router | `<bc>/persistence/<resource>/` repository_pg |
| 模块内领域事件 | `<bc>/<resource>/event.go`（`add crud` 自带占位） | `port.go` 定义 Publisher 接口 | — | `mq/` 或 `<bc>/...` 实现发布 |
| 跨模块集成事件 | `<bc>/event/<name>/`（`add event --domain <bc>`） | 手改接入 Command | — | `<bc>/event/<name>/publisher.go` |
| 消息队列消费 | 事件类型、Topic 常量 | CommandHandler.Consume | `mq/<module>/` 反序列化 Envelope | `mq/consumer.go`；`cmd/consumer` 进程 |
| 共享技术组件 | — | — | — | `shared/{dbx,cache,mq}`；运行时能力在 `goeasy` 模块 |

**bootstrap**：`register_<domain>.go` 装配同限界上下文下多个模块（`sys_roles` + `sys_menus` 合并）；`modules.go` 每域一行 `RegisterSystem`。

## 跨服务 gRPC（RPC Gateway）

调用其它微服务 gRPC 时，使用 **Port 子包 `app/<m>/port/` + infrastructure/rpc ACL + HTTPInfra.RPC 长连接**，业务 Query 一行 `gateway.GetByID(ctx, id)`。`query` 子包只 import `port`，不 import 父包 `app/<m>`（避免 import cycle）。

```bat
goeasy-cli gen proto --from-url <对端 proto>
goeasy-cli add rpcdemo --remote user
```

详见 [14 RPC Gateway 接入](14-rpc-gateway-integration.md)。MQ 接入见 [13 MQ 业务接入](13-mq-business-integration.md)。

## 持久化（sqlx + goqu）

- **sqlx**：执行 SQL、映射结构体 `db` 标签；连接来自 `goeasy/database.SQLX()`。
- **goqu**：在 `internal/infrastructure/shared/dbx` 按 `database.driver` 选方言，仓储内 `ToSQL()` 后由 sqlx 执行。
- 表名：`dbx.TableName(database.table_prefix, "<module>")`，与 `--with-migration` 一致。
- 禁止用字符串拼接用户输入；存量业务表（如 RBAC `sys_roles`）需手写仓储列映射，勿用 CLI 占位 `id+active` 模板。

## 脚手架边界

| 允许 | 禁止 |
|------|------|
| health 等技术示范 | CLI 默认生成 User、Order、Payment 等业务实体 |
| `add module` 由团队命名 | 在 framework 仓库写客户业务代码 |

## 延伸阅读

- [模板 v2 说明](../plans/goeasy-cli-ddd-lite-template-v2.md)
- 架构规则：仓库 `.rulesync/rules/architecture/`

## 下一步

[08 项目模板](08-templates.md)
