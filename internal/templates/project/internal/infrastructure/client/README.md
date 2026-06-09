# 外部 gRPC 客户端

- `grpc_client.go`：低级 dial 示例（联调/测试）。
- **业务跨服务调用**：使用 `internal/infrastructure/rpc/` Gateway + `HTTPInfra.RPC`（长连接）。
- 示范模块：`goeasy-cli add rpcdemo --remote <app_name>`
