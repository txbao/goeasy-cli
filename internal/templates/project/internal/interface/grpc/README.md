# gRPC 接口层

- `register.go`：占位；业务 Service 由 `bootstrap/register_<module>_grpc.go` 注册。
- `<module>/server.go`、`handlers.go`、`convert.go`：由 `add db proto` / `add db crud --with-proto` 生成。

流程：`add db proto` → `goeasy gen proto` → `bootstrap/grpc.go` 调用 `Register<Module>GRPC`。

详见 `goeasy-cli/docs/guide/11-grpc-internal.md`。
