# 20 库表契约：OpenAPI 与 Proto（P2）

从数据库自省生成 `api/openapi/<client>/<domain>/*.openapi.yaml` 与 `api/proto/*.proto`，列/表**注释**来自 PG `col_description` / MySQL `COLUMN_COMMENT`。

## 命令

| 命令 | 输出 |
|------|------|
| `goeasy add db openapi --table <m>` | `api/openapi/<client>/<domain>/<module>.openapi.yaml` |
| `goeasy add db proto --table <m>` | `api/proto/<module>.proto` |
| `goeasy add db crud --table <m> --with-openapi` | CRUD 代码 + OpenAPI |
| `goeasy add db crud --table <m> --with-proto` | CRUD 代码 + Proto |
| `goeasy add db all --all --with-proto --with-openapi` | 批量 CRUD + 契约 |

选表参数与 [06 命令](06-goeasy-cli-commands.md) 中 `add db crud` 相同（`--table` / `--tables` / `--all`）。

## migrate 与契约（勿混淆）

| 工具 | 作用 |
|------|------|
| `goeasy-cli migrate goto <version>` | **迁移版本号**，与表无关 |
| `goeasy add db proto --table sys_roles` | 按**当前表结构**生成 gRPC 契约 |
| `goeasy add db openapi --table sys_roles` | 按表结构生成 OpenAPI 3 |

建表请用 `migrate create` + `migrate up`，再用 `add db` 生成代码与契约。

## OpenAPI 内容

- 路径：`/api/v1/<client>/<domain>/<resource>`（GET 列表、POST 创建）、`.../{id}`（GET/PUT/DELETE）
- 示例：`/api/v1/admin/system/roles`（`sys_roles` + `codegen.domains.system`）
- 列表：query `page`、`page_size`（与 HTTP CRUD 一致）
- Schemas：`Create*Request`、`Update*Request`、`*DTO`、`ListResponse` + `PaginationMeta`
- `description`：优先列/表数据库注释

## Proto 内容

- Service：`List/Get/Create/Update/Delete`
- `List*Request`：`page`、`page_size`
- `List*Response`：`repeated <Entity> list`、`total`、`total_pages` 等
- 字段上方 `//` 注释：列备注

生成 `*.pb.go` 与 gRPC 服务注册：

| 方式 | 说明 |
|------|------|
| `add db proto`（默认） | 写完 `.proto` 后**自动尝试** `gen proto`；无 protoc 时 warn，需手动补跑 |
| `goeasy gen proto` | 对 `api/proto` 下全部 `.proto` 调用 protoc（需本机 protoc 与 go 插件） |
| `--skip-gen-proto` | 仅生成 `.proto` + gRPC 桩，不自动 protoc |
| `protoc` / buf | 见项目 `api/proto/README.md` |

生成 pb 后，`internal/interface/grpc/<domain>/<resource>/server.go` 方可编译（如 `system/roles`）；`api/proto/<module_id>.proto` 与 bootstrap `register_<module_id>_grpc.go` 仍用模块 ID。实现 RPC 后 `grpcurl list` 可见业务 Service。完整步骤见 [11 gRPC 项目集成](11-grpc-internal.md)。

## 示例（demo）

```bat
cd demo
goeasy-cli add db crud --table sys_roles --force --with-openapi --with-proto
```

## 下一步

- [11 gRPC 项目集成](11-grpc-internal.md)
- [19 项目配置 P0/P1](19-project-config-p0-p1.md)
- [06 CLI 命令](06-goeasy-cli-commands.md)
