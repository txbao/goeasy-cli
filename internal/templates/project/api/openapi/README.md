# OpenAPI 契约

由 `goeasy add db openapi` 或 `add db crud --with-openapi` 生成，源文件为 `*.openapi.yaml`。

- 列/表说明来自数据库注释（PG/MySQL 自省）
- 与 `internal/interface/http/<client>/<module>` REST 路径对齐：`/api/v1/admin/<module>`（默认 `--client admin`）

可用 Swagger UI、Redoc 或 CI 契约测试加载本目录 YAML。

详见 `goeasy-cli/docs/guide/20-db-openapi-proto.md`。
