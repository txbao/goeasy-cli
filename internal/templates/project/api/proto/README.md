# gRPC Proto 契约

由 `goeasy add db proto` 或 `add db crud --with-proto` 生成 `*.proto`。

## 目录约定

```text
api/proto/
├── *.proto           # 契约源文件（手改）
├── imported/         # gen proto --from-url 拉取的远程契约
└── gen/              # protoc 生成的 *.pb.go（需提交，勿手改）
```

## 生成 Go 代码

```bat
goeasy gen proto
```

需本机 protoc 与 `protoc-gen-go` / `protoc-gen-go-grpc`。详见 `api/proto/README.md`（init 后由模板生成）或 `goeasy-cli/docs/guide/11-grpc-internal.md`。
