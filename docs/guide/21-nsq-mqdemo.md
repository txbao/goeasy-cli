# 21 NSQ 消息示范（mqdemo）

基于 NSQ 的消息生产与消费示例，符合 DDD Lite 分层。按需生成，不随 `goeasy new` 默认内置。

## 前置条件

1. 本地或 Docker 运行 NSQ（nsqlookupd + nsqd）
2. `configs/config.yaml` 中启用 MQ 与 consumer 健康探针：

```yaml
mq:
  enabled: true
  type: nsq
  nsqd_addr: "127.0.0.1:4150"
  lookupd_addr: "127.0.0.1:4161"
  channel: "default"

consumer_http:
  enabled: true
  host: "0.0.0.0"
  port: 18080          # 自行配置，勿与 http.port / grpc.addr 冲突

observability:
  health:
    enabled: true
    path: /healthz
  mq:
    log_payload: false           # true 时日志附带 payload_preview（截断）
    log_payload_max_bytes: 512
```

## Payload 类型（`json.RawMessage`）

信封 `goeasy/mq.Envelope` 与各层 Command 的 `Payload` 统一为 **`json.RawMessage`**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `Envelope.Payload` | `json.RawMessage` | 线上原样透传，避免 `map[string]any` 数字精度问题 |
| `PublishCommand.Payload` | `json.RawMessage` | HTTP/CLI 传入任意合法 JSON（object/array 均可） |
| `ConsumeCommand.Payload` | `json.RawMessage` | 消费侧可按 `event_type` 解码为业务 struct |

业务强类型解码：`env.PayloadInto(&YourPayload{})`（见 [13 MQ 业务接入](13-mq-business-integration.md)）。

## MQ 日志

消费日志默认输出：`event_id`、`event_type`、`trace_id`、`payload_bytes`。

开启 `observability.mq.log_payload: true` 后额外输出 `payload_preview`（按 `log_payload_max_bytes` 截断）。生产环境建议保持 `false`，避免 PII 泄露。

## 端口规划

| 配置段 | 进程 | 用途 |
|--------|------|------|
| `http.port` | `cmd/service` | 业务 REST API |
| `grpc.addr` | `cmd/service` | 业务 gRPC（仅 service 会 `Serve`） |
| `consumer_http.port` | `cmd/consumer` | 健康探针 `/healthz`（可选 `/metrics`） |

**说明：** gRPC 采用延迟监听（`Serve()` 时才占用端口），consumer 与 service 可共用同一份 `config.yaml` 且 `grpc.enabled: true`，consumer 不会因初始化而抢 gRPC 端口。consumer **不提供** gRPC，仅订阅 MQ + 可选 HTTP 探针。

## 生成示范模块

```bat
cd goeasy-cli
go build -mod=mod -o goeasy-cli.exe .

cd mysvc
goeasy-cli add mqdemo
go mod tidy
```

**更新模板后重新生成：** 修改 `goeasy-cli/internal/templates/mqdemo` 后必须先 **重新编译** `goeasy-cli.exe`，再执行 `add mqdemo --force`。未重编译会继续使用旧内嵌模板（例如 `Payload` 仍为 `map[string]any`）。

生成内容：

- `internal/domain|app|interface|infrastructure` 下的 `mqdemo` 模块
- `cmd/consumer/main.go` 独立消费进程
- `internal/bootstrap/register_mqdemo.go` 并更新 `modules.go`

## 运行

```bat
REM 终端 A：HTTP 服务（含发布 API）
go run cmd\service\main.go

REM 终端 B：消费者（MQ 订阅 + consumer_http 探针）
go run cmd\consumer\main.go

REM 探活
curl http://localhost:18080/healthz

REM 终端 C：CLI 直发（无需 HTTP）
goeasy-cli mq publish ^
  --event-type demo.message.published ^
  --payload "{\"text\":\"hello\"}"

REM 或 HTTP 发布
curl -X POST http://localhost:8080/api/v1/demo/mq/publish ^
  -H "Content-Type: application/json" ^
  -d "{\"event_type\":\"demo.message.published\",\"payload\":{\"text\":\"hello\"}}"
```

## 分层说明

| 层 | 目录 | 职责 |
|----|------|------|
| domain | `internal/domain/mqdemo` | 消息实体校验、默认 topic |
| app | `internal/app/mqdemo` | Publish/Consume 用例、`MessageDispatch` 端口 |
| interface | `http/mqdemo`、`mq/mqdemo` | HTTP 与 NSQ 协议适配 |
| infrastructure | `infrastructure/mq` | NSQ 发布/订阅实现 |

## 业务模块接入

在其它接口（如创建订单）后发 MQ，见 [13 MQ 业务接入](13-mq-business-integration.md)。

## 调度扩展（预留）

见 `internal/bootstrap/snippets/mqdemo_scheduler.md`：`Application.Dispatch` 可供未来 `goeasy/scheduler` 与 HTTP、CLI 共用。

## 相关命令

- `goeasy-cli add mqdemo` — 生成模块
- `goeasy-cli mq publish` — CLI 直发 NSQ 信封消息

见 [06 goeasy-cli 命令](06-goeasy-cli-commands.md)。
