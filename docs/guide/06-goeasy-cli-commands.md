# 06 goeasy-cli 命令

CLI 二进制名为 **`goeasy`**（`go install` 后）。本地编译产物为 **`goeasy.exe`**，需在 `goeasy-cli` 目录用 `.\goeasy.exe` 调用。

## 命令总览

| 命令 | 说明 |
|------|------|
| `goeasy-cli version` | 版本信息 |
| `goeasy-cli new <name>` | 创建 DDD Lite 项目 |
| `goeasy-cli init <name>` | 同 `new` |
| `goeasy-cli add module <name>` | 完整业务模块骨架（无 CRUD HTTP 占位） |
| `goeasy-cli add crud <name>` | **推荐**：模块 + CRUD + `repository_pg`（`--with-migration` 可选） |
| `goeasy-cli add repository <name>` | 已废弃：仅为已有模块补 PG |
| `goeasy-cli add proto <name>` | `api/proto` gRPC 契约 |
| `goeasy-cli add event <name>` | 按事件名的领域事件 + 发布桩 |
| `goeasy-cli add mqdemo` | NSQ 消息生产/消费示范模块（DDD Lite） |
| `goeasy-cli add rpcdemo` | 跨服务 gRPC Gateway 示范模块（DDD Lite） |
| `goeasy-cli add rpc <proto>` | 共享 RPC 客户端（proto + gateway + 共享 port，步骤 1） |
| `goeasy-cli add rpc bind <proto>` | 业务模块绑定 RPC port + register wire（步骤 2） |
| `goeasy-cli mq publish` | CLI 直发 NSQ 信封消息（读项目 config） |
| `goeasy-cli grpc resolve` | 逻辑服务名 → gRPC target（direct / etcd） |
| `goeasy-cli grpc list` | 列出对端 gRPC Service（grpcurl + Reflection） |
| `goeasy-cli grpc call` | 调用 RPC（无需本地 .proto） |
| `goeasy-cli add db module` | 从表生成 domain + 持久化（无 CRUD HTTP 覆盖） |
| `goeasy-cli add db crud` | 从表自省生成模块 + CRUD + `repository_pg` |
| `goeasy-cli add db proto` | 从表结构生成 `api/proto/*.proto`（含 List RPC） |
| `goeasy-cli add db openapi` | 从表结构生成 `api/generated/openapi/*.openapi.yaml` |
| `goeasy-cli add db all` | 对匹配表批量 `crud`（可加 `--with-proto` / `--with-openapi`） |
| `goeasy-cli gen proto` | 对 `api/proto/*.proto` 运行 protoc 生成 `*.pb.go`（需本机 protoc 与插件） |
| （P3）gRPC | 见 [11 gRPC 项目集成](11-grpc-internal.md)；`add db proto` + `gen proto` + 实现 `server.go` |
| `goeasy-cli upgrade template` | 内嵌模板升级说明 |
| `goeasy-cli upgrade framework` | 查看 go.mod 中 goeasy 版本 |
| `goeasy-cli migrate up` | 应用 `migrations/<driver>/*.up.sql`（默认 postgres） |
| `goeasy-cli migrate down --steps <n>` | 回滚最近 n 个迁移 |
| `goeasy-cli migrate status` | 查看迁移状态 |
| `goeasy-cli migrate version` | 查看当前迁移版本 |
| `goeasy-cli migrate goto <version>` | 迁移到指定版本 |
| `goeasy-cli migrate force <version>` | 强制设置版本，不执行 SQL |
| `goeasy-cli migrate create <name>` | 新建 up/down SQL 对 |

## new / init

```bat
goeasy-cli new mysvc --module github.com/org/mysvc --download=false
```

### 常用参数

| 参数 | 默认 | 说明 |
|------|------|------|
| `--module` | 项目名 | Go module 路径，**强烈建议显式指定** |
| `--template` | `default` | 见 [08 项目模板](08-templates.md) |
| `--version` | `v1.0.0` | 远端模板版本（配合 `--download`） |
| `--download` | `false` | `true` 时尝试拉远端，失败回退内嵌模板 |
| `--goeasy-module` | `github.com/txbao/goeasy` | 运行时 module（或 `GOEASY_MODULE`） |
| `--goeasy-replace` | 自动检测 | monorepo 内 replace 本地 goeasy |

未传 `--module` 时 CLI 会输出警告，仍可使用项目名作为 module。

## add 命令（在已生成项目根目录）

全局参数（`add` / `add db` / `gen` 共用）：

| 参数 | 默认 | 说明 |
|------|------|------|
| `--dir` | `.` | 项目根目录 |
| `-f` / `--config` | `configs/config.yaml` | 配置文件；也可用环境变量 `GOEASY_CONFIG` |
| `--force` | false | 覆盖已存在文件；**默认**已存在则 `info: skip` 并继续（幂等） |
| `--app-style` | `service` | 应用层风格：`service` \| `light_cqrs`（别名 `light`）\| `full_cqrs`（别名 `full`，CLI 拒绝生成） |
| `--client` | `admin` | HTTP 客户端：`admin`、`h5`、`app`；可重复指定 |
| `--public` | 空 | 指定 client **不挂**鉴权中间件（如 `--public h5`）；须同时出现在 `--client`；**不允许** `--public admin` |
| `--domain` | 空 | 限界上下文（如 `system`）；`--group` 为兼容别名 |
| `--resource` | 空 | 资源名 / Go 包名（如 `roles`） |

详见 [18 应用层风格](18-app-style.md)。

**`add module` / `add crud` 会自动生成** `internal/bootstrap/register_<domain>.go`，并向 `modules.go` 追加 `Register<Domain>(engine, infra)`。`wire.go` 保持调用 `registerAllModules`，一般无需手改。

### 命令关系

```text
add crud          ≈ module + CRUD HTTP + repository_pg（推荐入口；可选 --with-migration）
add repository    → 已废弃：仅在为已有模块补 PG 时使用
add proto         → 独立，生成 api/proto/*.proto
add db crud|proto|all → 从数据库表自省生成（需 database.enabled + dsn + redis.enabled + addr）
add event         → 独立，按「事件名」生成，非按业务模块名
add aggregate     → 已隐藏/废弃，请改用 add module/crud
```

### 各命令说明

| 命令 | 生成内容 | 典型场景 |
|------|----------|----------|
| **`add crud <name>`** | module 骨架 + List 应用层 + CRUD HTTP + **占位** `repository_pg.go`（`id`/`active` 桩，与 `register_*.go` 同签名的缓存感知 `NewPGRepository`）；`--with-migration` 仅写本地 SQL；**不连接数据库**，**不需** `database`/`redis` 配置 | 标准 REST 业务模块（**首选**） |
| **`add module <name>`** | 最小分层骨架 + 默认 `admin` HTTP（`Get`/`Create`）+ 内存仓储 + `register_<m>.go`；**无** `query/list`、无 List API | 无库表、演示/集成类模块 |
| **`add repository <name>`** | （**已废弃**）仅补 `repository_pg.go` | 老项目补 PG；新模块用 `add crud` |
| **`add proto <name>`** | `api/proto/<name>.proto`（仅 **Get** RPC 最小桩，**无** gRPC server/register）；生产推荐 `add db proto` | 手写契约前的占位 |
| **`add event <name>`** | `domain/<bc>/event/<event>/event.go` + `infrastructure/<bc>/event/<event>/publisher.go`（默认 bc=`integration`；可用 `--domain`） | 跨模块集成事件桩；**模块内事件**请改 `domain/<bc>/<resource>/event.go` |

### 推荐工作流

**新建带 CRUD 的业务模块（如 `sys_roles`）：**

```bat
cd mysvc
goeasy-cli add crud sys_roles
goeasy-cli add crud sys_roles --with-migration
go mod tidy
goeasy-cli migrate up
REM 检查 internal/bootstrap/modules.go 已含 RegisterSysRoles（CLI 自动追加）
```

`Create` 接口支持 JSON `{"id":"..."}` 或 query `?id=`。

**公开 H5（无需会员 JWT）：**

```bat
goeasy-cli add crud products --client admin --client h5 --public h5
```

生成后 `register_*.go` 中：`/api/v1/admin` 仍挂 `AdminAuth`；`/api/v1/h5` 为公开路由组（无 `MemberAuth`）。`enterprise.jwt.enabled: false` 时 admin 返回 503，与 H5 是否公开无关。

名字驱动生成的 `repository_pg.go` 仅为 **可编译占位**（固定 `id`/`active` 列）。迁移落库后若需按真实表结构生成 goqu 仓储，请执行：

```bat
goeasy-cli migrate up
goeasy-cli add db crud --table sys_roles --force
```

**仅加 gRPC 最小契约（占位，无 server 桩）：**

```bat
goeasy-cli add proto sys_roles
REM 产出 api/proto/sys_roles.proto（仅 Get）；完整 CRUD 请用 add db proto
```

**库表 proto + 注册业务 gRPC Service（须先 crud）：**

```bat
goeasy-cli add db crud --table sys_roles --force
goeasy-cli add db proto --table sys_roles
REM add db proto 默认尝试自动 gen proto；无 protoc 时按提示手动执行下一行
goeasy-cli gen proto --file api/proto/sys_roles.proto
go mod tidy
go run .\cmd\service\main.go
REM proto/register 用 module_id（sys_roles）；实现代码在 internal/app/system/roles、internal/interface/grpc/system/roles
```

见 [11 gRPC 项目集成](11-grpc-internal.md)。

**加集成事件：**

```bat
goeasy-cli add event order-paid
goeasy-cli add event apis-paid --domain iam
REM 产出 internal/domain/iam/event/apis_paid/ ；模块内事件无需 add event
```

**加 NSQ 消息示范（mqdemo）：**

```bat
goeasy-cli add mqdemo
go mod tidy
go run cmd\consumer\main.go
goeasy-cli mq publish --event-type demo.message.published --payload "{\"text\":\"hello\"}"
```

详见 [12 NSQ 消息示范](12-nsq-mqdemo.md)。业务模块接入见 [13 MQ 业务接入](13-mq-business-integration.md)。

**加跨服务 gRPC 示范（rpcdemo）：**

```bat
REM 一步（推荐）：自动 fetch + protoc + 生成 Gateway（--proto 可从 --from-url 推断）
goeasy-cli add rpcdemo --remote demo1 --from-url ..\demo1\api\proto\sys_apis.proto --force

REM 或分步
goeasy-cli gen proto --from-url <对端 sys_roles.proto>
goeasy-cli add rpcdemo --remote user --proto sys_roles --force
```

| 参数 | 默认 | 说明 |
|------|------|------|
| `--remote` | `user` | 对端 `discovery.services` 键 |
| `--proto` | `sys_roles` | 对端 gRPC 模块名；省略时从 `--from-url` 文件名推断 |
| `--from-url` | 空 | 拉取对端 `.proto` 到 `api/proto/imported/` 并 protoc |
| `--skip-fetch-proto` | false | 跳过自动拉取（已有 `api/proto/gen/imported/<proto>/*.pb.go`） |
| `--app-style` | 读 config | 默认 `service`；`light_cqrs` 生成 `command/` + `query/` |
| `--force` | false | 覆盖已有文件；切换 `--proto` 时清理旧 `*_gateway.go` |

生成内容：`GET /api/v1/admin/rpcdemo/:id` 响应 proto **实体 message 全部字段**；`POST /api/v1/admin/rpcdemo` 请求体对齐 proto **`Create<实体>Request`**，并调用对端 `Create*` gRPC，响应为创建后的完整实体。字段从 `--from-url` / `api/proto/imported/<proto>.proto` 自动解析。

详见 [14 RPC Gateway 接入](14-rpc-gateway-integration.md)。

**业务模块内部调 gRPC（add rpc 两步，无 HTTP 示范）：**

```bat
REM 步骤 1：共享基础设施（无需 consumer）
goeasy-cli add rpc sys_roles --remote demo1 --from-url ..\demo1\api\proto\sys_roles.proto
goeasy-cli add rpc sys_apis  --remote demo1 --from-url ..\demo1\api\proto\sys_apis.proto

REM 步骤 2：绑定 consumer（可多个、可晚于步骤 1）
goeasy-cli add rpc bind sys_roles --consumer etcddemo
goeasy-cli add rpc bind sys_apis  --consumer etcddemo
```

| 参数（`add rpc`） | 默认 | 说明 |
|------|------|------|
| `--remote` | `user` | 对端 `discovery.services` 键 |
| `--from-url` | 空 | 拉取对端 `.proto` 到 `api/proto/imported/` 并 protoc |
| `--methods` | `all` | `Get,Create,Update,Delete,List` 或 `all` |
| `--skip-fetch-proto` | false | 跳过自动拉取（已有 pb） |
| `--force` | false | 覆盖 gateway / 共享 port |

| 参数（`add rpc bind`） | 默认 | 说明 |
|------|------|------|
| `--consumer` | — | **必填**，消费方模块 ID，如 `etcddemo` |
| `--remote` | 空 | 可选；未指定时从已有 `*_gateway.go` 目录推断 |
| `--wire` | true | 向 `register_<domain>.go` 幂等插入 `RPCClientLazy` + Gateway |
| `--force` | false | 覆盖 consumer port alias |

步骤 1 生成：`api/proto/imported/<proto>.proto`、`api/proto/gen/imported/<proto>/*.pb.go`、`internal/infrastructure/rpc/<remote>/<proto>_gateway.go`、`internal/infrastructure/rpc/<remote>/port/<proto>.go`（共享 Port）。

步骤 2 生成：`internal/app/<domain>/<resource>/port/<proto>.go`（type alias 到共享 port）；`--wire` 时在 `// goeasy-module: <consumer>` 函数内插入 `// goeasy-rpc-wire: <proto>`。**不自动改** `application.go` 的 `NewApplication`，需手改注入 Gateway。

与 `add rpcdemo` 对比：`add rpcdemo` = HTTP↔proto 联调示范；`add rpc` + `bind` = 业务模块内部编排调 gRPC。

### add db（库表驱动，postgres / mysql）

在项目根目录、且 `configs/config.yaml` 中 **`database.enabled: true` + 非空 `dsn`**，以及 **`redis.enabled: true` + 非空 `addr`** 时使用（CLI 执行前只校验配置，不 Ping；连库自省在 `loadTableMeta`）。详见 [库表驱动规划](../../docs/plans/goeasy-cli-database-first-codegen.md)。

| 子命令 | 说明 |
|--------|------|
| `add db module` | 按表生成 domain/app/持久化，不覆盖 HTTP CRUD 路由 |
| `add db crud` | 按表生成 domain/entity、goqu `repository_pg`、HTTP CRUD、`register_<module>` |
| `add db proto` | 标准 CRUD + `List` gRPC 契约；**默认自动尝试 `gen proto`**（无 protoc 时 warn 并提示手动执行） |
| `add db openapi` | 标准 REST OpenAPI 3（列/表注释 → `description`） |
| `add db all` | 批量 `crud`；`--with-proto` / `--with-openapi` 顺带契约 |

### gen（契约 → Go 代码）

| 子命令 | 说明 |
|--------|------|
| `gen http --from <openapi>` | 从 OpenAPI 3 生成 HTTP 层 + app/domain 桩（契约驱动）；**默认幂等**（已存在则 skip） |
| `gen http --dir-api api/contracts/openapi` | 批量 OpenAPI（SSOT） |
| `gen http --merge-http` | 库表先行后增量 HTTP：不碰 app/domain，不覆盖已有 HTTP 文件；OpenAPI 中非 CRUD 路径生成 `handler_openapi.go` / `router_openapi.go` |
| `gen http --allow-overwrite` | 与 `--force` 联用，允许覆盖 `add db crud` 产物（默认 `--force` 遇 `repository_pg` 会拒绝） |

`gen http` 与 `add db crud` 均**不生成** `internal/bootstrap/snippets/*_wire.md`（装配已自动化在 `register_<domain>.go`）。`snippets/` 仅保留 `add db proto` 的 `*_grpc.md` 与 `add mqdemo` 的调度说明。

**库表先行：** `add db crud` 后用 `add db openapi` 维护契约；避免 `gen http --force`。补自定义路由用 `--merge-http`。
| `gen grpc --from <proto>` | 从 `.proto` 生成 gRPC 桩（需已有 app 层） |
| `gen contract` | 批量：`api/contracts/openapi` + `api/proto`（默认 `--with-proto`） |
| `gen proto` | 对 `api/proto` 下 `.proto` 执行 protoc（`--file` 可指定单个） |
| `gen proto --from-url <url>` | 下载远程 `.proto` 到 `api/proto/imported/` 再 protoc |
| `gen proto --dir <root>` | 指定项目根，默认 `.` |

契约驱动完整说明见 [15 契约驱动生成](15-contract-first.md)。

依赖：本机 `protoc`、`protoc-gen-go`、`protoc-gen-go-grpc`。生成后实现 `internal/interface/grpc/<module>/server.go` 中的 RPC。跨服务联调见 [12 跨服务 gRPC](12-grpc-cross-service.md)。

与名字驱动：`add crud foo`（占位） vs `add db crud --table foo`（真实列）。

**选表方式（三选一）：**

| 参数 | 说明 |
|------|------|
| `--table sys_roles` | 单表（逻辑名；物理名 = `table_prefix` + 逻辑名） |
| `--tables a,b` | 多表逗号分隔 |
| `--all` | 当前 schema 下全部表；内置排除 `_sqlx_migrations` |

**常用参数：**

| 参数 | 默认 | 说明 |
|------|------|------|
| `-f` / `--config` | `configs/config.yaml` 或 `GOEASY_CONFIG` | 读取 `database.*` 与 `codegen.*`；通常可省略 |
| `--app-style` | `service` | 应用层风格，见 [18 应用层风格](18-app-style.md) |
| `--schema` | `public` | PG schema；MySQL 为 database 名（默认可从 DSN 解析） |
| `--module` | 空 | 覆盖模块名（默认由表名去 prefix 推导） |
| `--domain` | 空 | 限界上下文（如 `system`）；`--group` 为兼容别名 |
| `--resource` | 空 | 资源名 / Go 包名（如 `roles`） |
| `--dir` | `.` | 项目根 |
| `--force` | false | 覆盖已存在文件；未指定时跳过已有文件（幂等） |
| `--include-prefix` | 空 | `--all` 时只处理表名前缀 |
| `--exclude` | 空 | 额外排除表或 `prefix*` 模式 |
| `--skip-register` | false | 不写 `register_*.go` / `modules.go` |
| `--skip-proto` | false | `db all` 时不生成 proto |
| `--with-openapi` | false | `db crud` / `db all` 时顺带生成 OpenAPI |
| `--skip-openapi` | false | `db all` 时不生成 OpenAPI |

**领域布局（`config.yaml` → `codegen.domains`）：**

```yaml
codegen:
  layout: domain
  domains:
    system:
      table_prefix: sys_
      modules:
        sys_roles:
          resource: roles
        sys_menus:
          resource: menus
    iam:
      table_prefix: sys_
      modules:
        sys_apis:
          resource: apis
```

- `table_prefix` 最长匹配 → domain；`modules` 可显式指定 resource（**务必配置**，否则 resource 默认为表名如 `sys_apis`）。
- 生成目录：`internal/domain/system/roles/`、`internal/interface/http/admin/system/roles/`。
- `add` / `add db` 均支持 `--domain`、`--resource`。
- **同域多模块**：`register_<domain>.go` 按域**整文件重生成**；每模块独立 `register{Domain}{Module}()`，共享 `apiAdmin` 等 RouterGroup（如 `sys_roles` + `sys_menus` → `register_system.go`）。
- **HTTP 联调**：`add db crud` 同步生成 `api/examples/<client>/<domain>/<module_id>/crud.http`。
- **HTTP 与 app_style**：`handler.go` / `handler_crud.go` 为 HTTP 层按表拆分的 codegen 产物；`app_style: service` 表示应用层用 `Application` 方法，**不冲突**（`handler_crud` 仍调用 `h.app.List/Update/Delete`）。

**列裁剪规则（`config.yaml` → `codegen`）：**

- **Create 请求 / Insert 列**：默认省略 `id, created_at, updated_at, deleted_at`；时间戳由 SQL `CURRENT_TIMESTAMP`
- **Update 请求 / Set 列**：默认省略 `created_at, updated_at, deleted_at`；`updated_at` 在 PG 仓储中自动刷新
- **软删**：存在 `deleted_at` 列时 `Delete` 写软删而非物理删除
- **List 分页**：`GET /api/v1/<client>/<group>/<resource>`（或扁平 `/api/v1/<client>/<module>`）支持 JSON body 或 query `page` / `page_size`；成功时 `data` 为 `{ "list": [...], "pagination": { "page", "page_size", "total", "total_pages" } }`（依赖 `goeasy/pagination`）
- **HTTP 写请求**：按 `app_style` 绑定 Command——`service` 用 `app.<module>.CreateCommand`/`UpdateCommand`；`light_cqrs` 用 `command` 子包同名类型（均带 `json` 标签），不在 handler 内重复定义 `body` 结构体
- **import time**：`dto`、`repository_pg`、`entity` 等含时间列时自动生成 `import "time"`
- **Setter 去重**：`CreateCols` 与 `UpdateCols` 交集字段只生成一个 `SetXxx`
- **实体字段**：`entity` 结构体字段为导出（PascalCase，`id` 列生成 `ID`）；保留 `SetXxx`/`Rehydrate`，**不**生成与字段同名的 Getter（避免 `field and method with the same name`）；DTO/仓储使用 `root.Field` 访问
- **循环导入**：`query/list` 只返回 `[]*domain.Aggregate`；`app/<module>/list.go` 提供 `Application.List` → `ListResult`。若 `query/list.go` 仍含 `import .../internal/app/<module>` 或 `app.ListResult`，属于旧生成物；CLI 会在无 `--force` 时尝试自动覆盖该文件，并提示对 `app/list.go`、`handler_crud.go` 使用 `--force`
- **参数校验**：HTTP 在 `ShouldBindJSON` 后调用 `goeasy/validator.Validate`（Command 与分页 `Page` 带 `validate` 标签）
- **实体缓存（P1）**：`repository_pg` 的 `FindByID` 读 Redis（key `{key_prefix}:{module}:id:{id}`），`Update`/`Delete` 删缓存；需 `redis.enabled` + `cache.enabled`，见 [09 项目配置 P0/P1](09-project-config-p0-p1.md)

**若曾用名字驱动生成过同模块：** 再跑 `add db crud` 会出现 `info: skip existing` 并保留旧占位代码，必须加 `--force` 才会按库表覆盖。

**示例：**

```bat
cd mysvc
goeasy-cli migrate up
goeasy-cli add db crud --table sys_roles --force
goeasy-cli add db proto --table sys_roles
goeasy-cli gen proto
goeasy-cli add db openapi --table sys_roles --force
goeasy-cli add db crud --table sys_roles --force --with-openapi --with-proto
goeasy-cli add db all --all --include-prefix sys_ --with-proto --with-openapi
goeasy-cli add db module --table sys_roles
goeasy-cli add db crud --table sys_roles --app-style light_cqrs --force
go mod tidy
```

- `--table _sqlx_migrations` 会被拒绝（内置保留表）。
- 批量/重复执行时，未加 `--force` 会跳过已存在的 `*.proto`，但 **`add db proto` 仍会补齐缺失的 gRPC 桩**（`internal/interface/grpc/<module>/`、`register_<module>_grpc.go`）。
- **`add db proto` 必须先有 app 层**（先 `add db crud`）；否则会报错且不会写入 proto。
- **`add db proto` 后须有 `api/proto/gen/<module>/*.pb.go` 方可编译**（CLI 默认自动 protoc；`--skip-gen-proto` 跳过；失败时手动 `gen proto --file api/proto/<module>.proto`）。
- **`add db proto` / `gen grpc` 的 `handlers.go` 与 `codegen.app_style` 一致**（默认 `service`）；切换风格后请 `add db crud --force` 再 `add db proto --force`。

生成后请人工核对：复杂 RBAC 表、枚举约束、跨表关联等场景可能需要手改 `repository_pg.go`。

### add crud / add module 编译排错（app_style）

| 现象 | 原因 | 处理 |
|------|------|------|
| `assignment mismatch: Create returns 1 value` | 旧版名字驱动 app 层与 HTTP codegen 不一致 | 升级 CLI 后 `goeasy-cli add crud <m>`（会覆盖 handler 与 `application.go`） |
| `undefined: app.UpdateCommand` / `h.app.Update undefined` | 同上，service 风格缺 Update/Delete | 同上；或已有表时 `add db crud --table <m> --force` |
| `Commands().Create` 返回值不匹配 | light_cqrs 旧版 `command/create.go` 返回 `error` | 升级 CLI 后 `goeasy-cli add crud <m> --app-style light_cqrs` |
| `too many arguments in call to NewPGRepository` | 旧版占位 `repository_pg.go` 为 3 参数，register 为 8 参数 | 升级 CLI 后 `goeasy-cli add crud <m> --force` 重写 `repository_pg.go` |
| `redis.enabled is false` / `database.dsn is empty`（`add db crud`） | 库表驱动前置配置未就绪 | 见 [09 项目配置 P0/P1](09-project-config-p0-p1.md) 配齐 database + redis 后再执行 |

名字驱动（`add crud`）与库表驱动（`add db crud`）的 **service / light_cqrs** app 层现已与 HTTP、gRPC codegen 对齐；`add db proto` / `gen grpc` 依赖已有 app 层，请先修好 app 再生成 gRPC。

### add module 排错

| 现象 | 原因 | 处理 |
|------|------|------|
| `no required module provides package .../admin/<module>` | `register_<m>.go` 已生成，但 `internal/interface/http/admin/<module>/` 缺失 | 升级 CLI 后执行 `goeasy-cli add module <m> --force` 补全 HTTP 层 |
| 仅有 domain/app，无 `interface/http/admin` | 旧版 CLI 模板过滤 bug | 同上 |
| `h.repo.List undefined` | 旧版 `add module` 误生成 `query/list.go`，但无 `List` 仓储接口 | 升级 CLI 后 `goeasy-cli add module <m> --force`（会移除 list）；需 List 请用 `add crud` |
| `undefined: ListResult` / `repo.List undefined`（`application.go`） | service 风格 `add module` 误生成 `List()`，但 domain/dto 无 List | 升级 CLI 后 `goeasy-cli add module <m> --force`；需 List 请用 `add crud` |

`add module` 默认内存仓储（`NewRepository()`），不依赖数据库；`register` 仅在 `infra.DB.SQLX()!=nil` 时切换 PG。

默认 `--client admin`，无需手动指定；多客户端示例：`goeasy-cli add module orders --client admin --client h5`。公开 H5：`--client h5 --public h5`（可与 admin 组合）。

### add rpc / add rpc bind 排错

| 现象 | 处理 |
|------|------|
| `missing gateway` / `run add rpc first` | 先执行步骤 1：`add rpc <proto> --remote <r> --from-url <proto>` |
| `multiple gateways` | 多个 remote 下同 proto；bind 时显式 `--remote` |
| wire 失败 | 查看 `internal/bootstrap/snippets/<consumer>_rpc_<proto>.md` 手改 `register_*.go` |
| `declared and not used`（Gateway 变量） | 手改 `application.go` 注入 Gateway 后删除 `_ = xxxGW` |
| 编译缺 pb | 步骤 1 加 `--from-url` 或先 `gen proto --from-url` |

### add rpcdemo 排错

| 现象 | 原因 | 处理 |
|------|------|------|
| `import cycle not allowed`（`app/rpcdemo` ← `query`） | 旧版模板 `query` import 父包取 Port | 升级 CLI 后 `goeasy-cli add rpcdemo --remote user --force`（Port 在 `app/rpcdemo/port/`） |
| 编译缺 `sys_roles` / `sys_apis` pb | 未拉取对端 proto 或 `--proto` 不匹配 | `add rpcdemo --from-url <对端 proto> --proto <name>` 或先 `gen proto --from-url` |
| `Application` 用 `Queries()` 但 config 为 `service` | 旧版 rpcdemo 固定 light_cqrs | 升级 CLI 后 `add rpcdemo --remote <r> --force` |
| dial 超时 | 注册地址不可达 | 见 [14 RPC Gateway](14-rpc-gateway-integration.md) 配置与排错 |
| 启动 `register http routes: rpcdemo: grpc client lazy "<remote>": ...` | `discovery.services.<remote>` 未配或与 `--remote` 不一致 | 补 yaml，如 `demo1: "127.0.0.1:18083"`（gRPC 端口，非 etcd 2379） |
| 仅见 `/health`/`/healthz`、无 `http listening`、Gin 无 rpcdemo 路由 | **discovery/etcd 未配好**，`RegisterRPCDemo` 未完成 | 本地 dev：`mode: direct`、`etcd.enabled: false`，并配 `services.<remote>`；或修好 ETCD 集群与注册 |
| curl `(7) Could not connect` | HTTP 未监听 | 先解决上条，确认日志有 `http listening on` |
| GET/POST 仅 `id`/`active`，Create 只收 `id` | 旧版 rpcdemo 占位 DTO | `add rpcdemo --remote <r> --from-url <proto> --force` 按 proto 重生成 |

### 避免重复执行

| 错误做法 | 原因 |
|----------|------|
| `add crud` 后再 `add module` 同名 | domain/app 文件已存在 |
| 使用 `add aggregate` | 已废弃，请用 `add crud` |
| 对已有模块 `add repository --force` | 可能覆盖 domain/repository.go，破坏 CRUD 接口 |

## grpc（在项目根目录）

跨服务调用、ETCD 发现、Reflection 联调（无需复制 proto）。详见 [12 跨服务 gRPC](12-grpc-cross-service.md)。

```bat
goeasy-cli grpc resolve --service user
goeasy-cli grpc list --service user
goeasy-cli grpc call --service user --method sys_roles.SysRolesService/GetSysRoles --data "{\"id\":\"1\"}"
goeasy-cli grpc call --target 10.10.10.12:28021 --method sys_roles.SysRolesService/GetSysRoles --data "{\"id\":\"1\"}"
```

| 参数 | 说明 |
|------|------|
| `--dir` | 项目根，默认 `.` |
| `-f` / `--config` | 配置文件 |
| `--service` | 逻辑服务名（`app_name` / `discovery.services` key） |
| `--target` | 显式 `host:port`，跳过发现 |
| `--method` | `call` 必填，如 `sys_roles.SysRolesService/GetSysRoles`（先用 `grpc list` 确认方法名） |
| `--data` | `call` 请求 JSON，默认 `{}`；**须同一行**，CMD 示例见 [12 跨服务 gRPC](12-grpc-cross-service.md) |
| `--plaintext` | 默认 `true`（无 TLS） |
| `--rpc-service` | `list` 可选，列出某 Service 的方法 |

需本机 [grpcurl](https://github.com/fullstorydev/grpcurl/releases)；对端 goeasy 默认已开 Server Reflection。

## mq（在项目根目录）

```bat
goeasy-cli mq publish --event-type demo.message.published --payload "{\"text\":\"hello\"}"
```

| 参数 | 说明 |
|------|------|
| `--dir` | 项目根，默认 `.` |
| `-f` / `--config` | 配置文件，默认 `configs/config.yaml` |
| `--event-type` | 事件类型（默认同时作为 NSQ topic） |
| `--topic` | 覆盖 NSQ topic |
| `--payload` | JSON 对象，默认 `{}` |
| `--source` | `source_service`，默认 `goeasy-cli` |
| `--trace-id` | 可选 trace_id |

需 `mq.enabled: true` 且 `nsqd_addr` 可连。不要求已执行 `add mqdemo`。

## migrate（在项目根目录）

```bat
goeasy-cli migrate up
goeasy-cli migrate status
goeasy-cli migrate version
goeasy-cli migrate goto 3
goeasy-cli migrate force 3
goeasy-cli migrate down --steps 1
goeasy-cli migrate create add_users_table
```

| 参数 | 默认 | 说明 |
|------|------|------|
| `--dir` | `.` | 项目根 |
| `-f` / `--config` | `configs/config.yaml` | 读取 `database.dsn`、`database.driver` |
| `--migrations` | `migrations` | 迁移根目录；默认解析为 `migrations/postgres` 或 `migrations/mysql` |

目录结构：

```text
migrations/
├── postgres/   # database.driver: postgres（默认）
└── mysql/      # database.driver: mysql
```

`migrate create` 会在当前 driver 对应子目录下生成 up/down 文件。

需 `database.enabled: true`，`driver` 为 `postgres` 或 `mysql`。版本表 `_sqlx_migrations`（`version` + `dirty`）由 golang-migrate 自动创建。

**注意：** `migrate goto` 只表示 SQL 迁移版本，**不能**按表名生成 proto/OpenAPI；表契约请用 `add db proto` / `add db openapi`（见 [10 库表契约](10-db-openapi-proto.md)）。

## upgrade

- `upgrade template`：模板随 CLI 版本发布，升级 CLI 即升级内嵌模板
- `upgrade framework`：提示业务项目 bump `goeasy` 依赖版本

## Windows 注意事项

| 现象 | 处理 |
|------|------|
| `goeasy-cli.exe: command not found` | 使用 `.\goeasy-cli.exe` 或把 `go\bin` 加入 PATH 后用 `goeasy-cli` |
| 在 `goeasy-cli` 目录未编译 | 先 `go build -o goeasy-cli.exe .` |

## 下一步

[07 DDD Lite 实践](07-ddd-lite-practices.md) · [12 跨服务 gRPC](12-grpc-cross-service.md) · [14 RPC Gateway](14-rpc-gateway-integration.md) · [12 NSQ 消息示范](12-nsq-mqdemo.md) · [13 MQ 业务接入](13-mq-business-integration.md)
