# 13 MQ 业务接入（DDD Lite）

在业务接口（如「创建订单」）成功后向 NSQ 投递消息的标准做法。以 `order` 模块、事件 `order.created` 为例。

## 原则

- **domain** 不依赖 NSQ / `goeasy/mq`
- **interface** 不直接 `mq.Publish`
- **写用例** 在 `app` 层编排：持久化成功后再发布
- 消息体使用 `goeasy/mq.Envelope`（含 `event_id`、`trace_id` 等）

## 接入步骤

### 1. 领域事件（domain）

`internal/domain/order/event.go`：

```go
const TopicOrderCreated = "order.created"

type OrderCreated struct {
    OrderID string
    // ...
}

func (e OrderCreated) Name() string { return TopicOrderCreated }
```

### 2. 发布端口（app）

`internal/app/order/port.go`：

```go
type OrderEventPublisher interface {
    PublishOrderCreated(ctx context.Context, evt domain.OrderCreated) error
}
```

### 3. 基础设施实现（infrastructure）

`internal/infrastructure/mq/order_publisher.go`：

```go
type orderCreatedPayload struct {
    OrderID string `json:"order_id"`
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, evt domain.OrderCreated) error {
    raw, _ := json.Marshal(orderCreatedPayload{OrderID: evt.OrderID})
    env, err := goeasymq.NewEnvelope(evt.Name(), p.source, traceIDFromCtx(ctx), raw)
    if err != nil {
        return err
    }
    body, _ := env.Marshal()
    return p.mq.Publish(ctx, evt.Name(), body)
}
```

消费侧解码：

```go
var p orderCreatedPayload
if err := env.PayloadInto(&p); err != nil { ... }
```

`bootstrap/register_order.go` 注入 `infra.MQ` 构造 `Publisher`。

### 4. 写用例中发布（app/command）

`internal/app/order/command/create.go`：

```go
func (h *Handler) Create(ctx context.Context, cmd CreateCommand) error {
    agg := domain.NewOrder(cmd.ID)
    if err := h.repo.Save(ctx, agg); err != nil {
        return err
    }
    return h.publisher.PublishOrderCreated(ctx, domain.NewOrderCreated(agg.ID()))
}
```

**顺序：** `Save` 成功 → 再 `Publish`。

### Outbox 模式（可选，`mq.outbox.enabled: true`）

保证 DB 与 MQ 最终一致：事务内写 outbox 表，系统 Cron `outbox_relay` 异步投递。

```go
return infra.DB.Transaction(ctx, func(txCtx context.Context) error {
    if err := h.repo.Save(txCtx, agg); err != nil {
        return err
    }
    body, _ := env.Marshal()
    return infra.Outbox.PublishInTx(txCtx, domain.TopicOrderCreated, body)
})
```

前置：`database.enabled`、`mq.enabled`、`mq.outbox.enabled`，并执行 `migrations/*/000002_outbox.up.sql`。
默认表名 `goeasy_outbox`（`mq.outbox.table` 可改，迁移 SQL 须与运行时 `OutboxTable()` 一致，含 `database.table_prefix`）。
默认 `outbox.enabled: false` 时，`infra.Outbox.Publish` 直发 MQ。

### 5. 消费者（独立进程）

- 另一 topic 订阅者在 `cmd/consumer` 中注册
- `interface/mq/<module>` 反序列化 `Envelope` → 调 `app.CommandHandler.Consume`
- 用 `event_id` 做幂等（参考 `mqdemo` 内存仓储）

## 数据流

```text
POST /orders
  → interface/http/order.Handler.Create
  → app/order.CommandHandler.Create
  → domain + Repository.Save
  → app/port OrderEventPublisher
  → infrastructure/mq → goeasy/mq → NSQ

NSQ → cmd/consumer
  → interface/mq/order.Handler
  → app/order.Consume（幂等）
```

## 与脚手架命令关系

| 命令 | 用途 |
|------|------|
| `add mqdemo` | 完整发布/消费示范，可复制 `infrastructure/mq` |
| `add event order-created --domain order` | 生成 `domain/order/event/order_created/` + publisher 桩，需手改接入 `goeasy/mq`；模块内事件请改 `domain/<bc>/<resource>/event.go` |
| `goeasy-cli mq publish` | 运维联调，不经业务 CommandHandler |

## 事件命名

采用过去式或已完成式，例如：

- `order.created`
- `payment.order.paid`

信封字段规范见项目 `docs/AI生成提示语.md` §12。

## 配置

与 [21 NSQ 消息示范](21-nsq-mqdemo.md) 相同：`mq.enabled`、`nsqd_addr`、`lookupd_addr`；consumer 使用 `consumer_http` 做探活。
