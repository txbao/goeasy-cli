# api/contracts/openapi

契约驱动（SSOT）的 OpenAPI 3 文档目录。`goeasy-cli gen http` / `gen contract` 默认从此处读取。

建议工作流：

1. 手写或从 Apifox 导出 `*.openapi.yaml`
2. `goeasy-cli gen http --from api/contracts/openapi/<module>.openapi.yaml`
3. 实现 handler / app 业务逻辑

路径约定：`/api/v1/{client}/{domain}/{resource}`（如 `/api/v1/admin/system/roles`）。
