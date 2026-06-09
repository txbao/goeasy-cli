# gRPC 与服务发现（P3）

> 本文描述 **goeasy 运行时库**的 gRPC 服务端、Reflection、服务发现与客户端能力。业务 proto、server 实现与 `goeasy gen proto` 见 [11 gRPC 项目集成](../guide/11-grpc-internal.md)。

## 配置

```yaml
grpc:
  enabled: true
  addr: "0.0.0.0:9001"
  timeout_sec: 5
  max_retries: 2

discovery:
  mode: direct          # direct | etcd
  etcd:
    enabled: false
    endpoints: ["127.0.0.1:2379"]
    lease_ttl_sec: 30
    advertise_addr: ""  # 注册可达地址，空则回退 services[app_name]
    prefix: /goeasy/services
  services:
    peer-svc: "127.0.0.1:9002"
```

| mode | 场景 |
|------|------|
| `direct` | 单机/开发，`discovery.services` 静态映射逻辑名 → `host:port` |
| `etcd` | 多实例；`etcd.enabled: true` 且 `mode: etcd` |

## 服务端启动与注册时机

```go
application := app.New(cfg)
application.RegisterHTTP(bootstrap.RegisterRoutes)
application.RegisterGRPC(bootstrap.RegisterGRPCServers)
application.Run()
```

`GRPCRegister` 签名为 `func(*grpcx.Server, HTTPInfra)`，与 HTTP 共用同一套 `HTTPInfra`（DB、Cache、JWT 等）。

`app.Run()` 在 `InitInfra()` 之后、开始 `Serve()` 之前调用 `grpcReg`：

```text
InitInfra → grpcx.NewServer（含 reflection）→ grpcReg(s) 注册业务 Service → Serve
```

| 组件 | 职责 |
|------|------|
| `grpcx.Server` | 监听 `grpc.addr`，`GRPC()` 返回 `*grpc.Server` 供 `pb.RegisterXxxServiceServer` |
| `grpcx.NewServer` | 默认 `reflection.Register`，供 `grpcurl list` |
| `discovery.Registry` | gRPC 监听绑定后，将 `app_name` → 可达地址注册到 etcd（KeepAlive 保活），退出时撤销租约 |

注册地址优先级：`discovery.etcd.advertise_addr` > `discovery.services[app_name]` > 监听地址（`0.0.0.0` 回退 `127.0.0.1`）。

etcd 验证（服务运行中）：

```bat
etcdctl get /goeasy/services --prefix
etcdctl lease list
```

key 为 `{prefix}/{app_name}`，例如 `/goeasy/services/my-app`。

**仅看到 Reflection、没有业务 Service：** Reflection 由框架注册；业务 Service 必须在 `RegisterGRPC` 链中调用 `pb.Register*ServiceServer`，且项目已 `protoc` 生成 `*.pb.go` 并实现 RPC（见 CLI 文档）。

## 客户端

### 长连接 Registry（推荐，bootstrap / Gateway）

```go
// app.Run → HTTPInfra.RPC 在 InitInfra 后可用
cli, err := infra.RPC.Client(ctx, "peer-svc") // 同 service 进程内复用连接
// pb.NewXxxServiceClient(cli.Conn()).Method(...)
// 勿在每次 RPC 后 cli.Close()；进程退出时 app.Run 会 RPC.Close()
```

`HTTPInfra.RPC` 为 `*grpcx.Registry`，按逻辑服务名缓存 `*grpcx.Client`，适合跨服务 Gateway（类似 go-zero zrpc）。

### 单次连接（低级 API）

```go
cli, err := app.NewGRPCClientForService(ctx, "peer-svc")
defer cli.Close()
conn := cli.Conn()
// pb.NewXxxServiceClient(conn).Method(...)
```

显式地址：`app.NewGRPCClient("127.0.0.1:9001")`。`grpcx.Client.Invoke` 包装熔断与重试。

## 包一览

| 包 | 职责 |
|----|------|
| `grpcx` | Server / Client / `Registry` / `ResolveService` |
| `discovery` | `Register` / `Resolve` / `Deregister` |

## 下一步

- [12 跨服务 gRPC](../guide/12-grpc-cross-service.md)（order 调 user、`goeasy-cli grpc resolve/call`）
- [11 gRPC 项目集成](../guide/11-grpc-internal.md)（proto、server、gen proto、grpcurl 排错）
- [05 运行时总览](../guide/05-goeasy-runtime.md)
