# 15 契约驱动生成（contract-first）

在「库表驱动」之外，支持 **先写 OpenAPI / Proto，再生成接口层与应用层桩**。

## 推荐流程

```text
① SSOT 契约（手写或 Apifox 导出）
   api/openapi/<client>/<domain>/<module_id>.openapi.yaml
   api/proto/<module>.proto

② 生成代码桩
   goeasy-cli gen http --from api/openapi/admin/system/sys_roles.openapi.yaml
   goeasy-cli gen proto
   goeasy-cli gen grpc --from api/proto/sys_roles.proto

   或一步：goeasy-cli gen contract

③ 手改 domain 规则、repository_pg、复杂校验
④ migrate up → go run ./cmd/service
```

库表自省同样写入 `api/openapi/<client>/<domain>/`（`add db openapi` / `--with-openapi`）。

## 命令

| 命令 | 说明 |
|------|------|
| `gen http --from <openapi>` | 解析 OpenAPI 3 → HTTP + app/domain 桩 |
| `gen http --dir-api api/openapi` | 批量（默认目录，递归扫描） |
| `gen grpc --from <proto>` | gRPC 桩（**需已有 app 层**） |
| `gen contract` | 批量 HTTP + gRPC；默认 `--with-proto` |

**常用参数：** `--force`、`--allow-overwrite`、`--merge-http`、`--domain`、`--resource`、`--client`、`--skip-app`

**幂等：** 未加 `--force` 时，已存在文件跳过（`info: skip existing`），可重复执行 `gen http`；不会写入 `internal/bootstrap/snippets/*_wire.md`。

### 与库表驱动（add db crud）的边界

| 场景 | 推荐做法 |
|------|----------|
| 已 `add db crud` 生成模块 | 用 `add db openapi` 同步 REST 契约；**勿** `gen http --force` |
| 需在已有 CRUD 上补 OpenAPI 自定义路由 | `gen http --merge-http --from <openapi>`（不覆盖 handler/router，仅生成 `handler_openapi.go` / `router_openapi.go`） |
| 确需用契约重生成全部代码 | `gen http --force --allow-overwrite`（显式确认覆盖 `repository_pg` 等产物） |

## OpenAPI 约定

- 路径：`/api/v1/{client}/{domain}/{resource}`（如 `/api/v1/admin/system/roles`）
- 目录：`api/openapi/{client}/{domain}/{module_id}.openapi.yaml`
- 布局由 OpenAPI 路径 + `codegen.domains` 解析为 `domain/system/roles`

## 示例

```bat
REM 契约先行（无 db crud 产物时）
goeasy-cli gen http --from api/openapi/admin/system/sys_roles.openapi.yaml --force

REM 库表先行后补自定义路由
goeasy-cli gen http --merge-http --from api/openapi/admin/system/sys_roles.openapi.yaml

goeasy-cli gen contract --force
go mod tidy
go run ./cmd/service
```

## 下一步

- [20 库表契约](20-db-openapi-proto.md)
- [11 gRPC 项目集成](11-grpc-internal.md)
- [06 CLI 命令](06-goeasy-cli-commands.md)
