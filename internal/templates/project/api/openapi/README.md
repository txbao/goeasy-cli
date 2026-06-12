# api/openapi

OpenAPI 3 契约目录。`goeasy-cli gen http` / `gen contract` 默认从此处递归读取 `*.openapi.yaml`。

## 目录布局

```text
api/openapi/{client}/{domain}/{module_id}.openapi.yaml
```

示例：`api/openapi/admin/system/sys_roles.openapi.yaml`

- `{client}`：`admin` / `h5` / `app`（与 `--client` 一致）
- `{domain}` / `{module_id}`：由 `codegen.domains` 与表名解析

## 来源

- `goeasy add db openapi` 或 `add db crud --with-openapi`：库表自省生成
- 手写或 Apifox 导出：契约驱动（contract-first）

## 常用命令

```bat
goeasy-cli gen http --from api/openapi/admin/system/sys_roles.openapi.yaml
goeasy-cli gen contract
```

路径约定：`/api/v1/{client}/{domain}/{resource}`（如 `/api/v1/admin/system/roles`）。

详见 `goeasy-cli/docs/guide/20-db-openapi-proto.md`。
