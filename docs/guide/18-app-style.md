# 18 应用层风格（app_style）

GoEasy 支持三种应用层组织方式，通过 `codegen.app_style` 或 CLI `--app-style` 选择。

## 三种风格

| 值 | 说明 | CLI 生成 |
|----|------|----------|
| `service`（**默认**） | DDD Lite + `Application` 方法：`Create/Get/List/Update/Delete` | ✅ |
| `light_cqrs` | `command/` + `query/` 子包 + `Queries()`/`Commands()` 门面 | ✅ |
| `full_cqrs` | 独立读模型、投影、事件溯源 | ❌ 仅文档引导 |

CLI 别名：`light` → `light_cqrs`，`full` → `full_cqrs`（`full` 会报错并提示文档）。

## 配置

```yaml
codegen:
  app_style: service
  domains:
    system:
      app_style: light_cqrs      # 域级覆盖
      modules:
        sys_roles:
          app_style: service     # 模块级覆盖（优先级最高）
```

解析优先级：`--app-style` > 模块 > 域 > 项目默认 > 命令内置默认（`add db crud` 等为 `service`）。

`goeasy new` 生成的 **health** 示范模块固定为 `light_cqrs`，不受上述配置影响。

`add rpcdemo` 遵循 `codegen.app_style`（默认 `service`）：`service` 时 `Application.Get/Create` + HTTP `h.app.Get`；`light_cqrs` 时保留 `command/`、`query/` 与 `Queries()`/`Commands()`。

## CLI 示例

```bat
REM 默认 service，无需 -f（默认 configs/config.yaml，或环境变量 GOEASY_CONFIG）
goeasy-cli add db crud --table sys_roles --force

REM 显式 light_cqrs
goeasy-cli add db crud --table sys_roles --app-style light_cqrs --force

REM 非默认配置文件
goeasy-cli migrate up -f configs\config.prod.yaml
set GOEASY_CONFIG=configs\config.prod.yaml
goeasy-cli add db crud --table sys_roles
```

## 名字驱动 vs 库表驱动

| 命令 | app 层 | HTTP / gRPC handler |
|------|--------|---------------------|
| `add crud` / `add module` | 按 `app_style` 生成 | HTTP **codegen**（`handler.go` / `handler_crud.go`） |
| `add db crud` | 按 `app_style` 生成 | HTTP **codegen**（同上） |
| `add db proto` / `gen grpc` / `gen contract` | 需已有 app 层 | gRPC **codegen**（`handlers.go`），与 app 层一致 |

`add crud sys_roles` 在 `codegen.domains` 下会生成 `internal/interface/http/admin/system/roles/`，默认 `service` 写法为 `h.app.Get` / `h.app.Create` 等。

名字驱动与库表驱动在 **service / light_cqrs** 下共用同一套 HTTP、gRPC codegen 约定：

- **service**：`Application` 上 `Create(ctx, cmd) (string, error)`、`Update`、`Delete`、`List`
- **light_cqrs**：`Commands().Create` 返回 `(string, error)`；`add crud` 会同步写入 `application.go` 与 `command/`、`query/`

## 存量项目

- domain/app/infrastructure 等：**不加 `--force` 时跳过已有文件**。
- HTTP `handler.go` / `handler_crud.go`：再次执行 `add crud` / `add module` 时会按当前 `app_style` **覆盖 handler**（无需 `--force`）。
- 切换 app 层风格：对目标模块 `add db crud --app-style service --force` 或 `add crud <name> --app-style service --force`；若已生成 gRPC 桩，再执行 `add db proto --table <m> --force`。

## 目录对比

**service：**

```text
app/<domain>/<resource>/
├── application.go
└── dto.go
```

**light_cqrs：**

```text
app/<domain>/<resource>/
├── application.go
├── command/
├── query/
├── list.go
└── dto.go
```

## Full CQRS

CLI 不生成 `full_cqrs` 模板。可在业务项目中基于 Outbox（`migrations/000002_outbox`）与读模型表手工演进，详见架构规则。

## 相关文档

- [07 DDD Lite 实践](07-ddd-lite-practices.md)
- [06 goeasy-cli 命令](06-goeasy-cli-commands.md)
