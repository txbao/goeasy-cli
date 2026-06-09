# api/proto/gen

protoc 生成的 Go 桩代码目录（`*.pb.go` / `*_grpc.pb.go`），由 `goeasy gen proto` 产出。

- **勿手改**：修改契约请编辑上级目录的 `api/proto/*.proto`，再执行 `goeasy gen proto`。
- **需提交**：gRPC 编译依赖此目录下的生成代码，请与 `.proto` 一并纳入版本控制。
