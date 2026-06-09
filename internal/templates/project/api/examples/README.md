# api/examples

`goeasy-cli add db crud` 按模块自动生成的 HTTP 联调示例（与路由 `/api/v1/{client}/{domain}/{resource}` 一致）。

```text
api/examples/
└── admin/
    └── system/
        └── sys_roles/
            └── crud.http   # VS Code / IDEA REST Client
```

- 目录第三段为 **模块 ID**（表名，如 `sys_roles`），第四段资源路径仍由 `codegen.domains` 的 `resource` 决定。
- 鉴权端点需在 `crud.http` 填写 `@token`。
