# 16 运行时能力清单（治理 / 缓存 / 观测 / 安全）

本文汇总 goeasy 运行层已下沉的能力及项目侧调用方式。配置模板见 `configs/config.yaml`（`goeasy new` 生成）。

## 1. 限流与熔断

### 限流（`goeasy/limiter`）

| 配置 | 说明 |
|------|------|
| `governance.limiter.enabled` | 开关 |
| `mode` | `local` / `redis` / `both` |
| `dimensions` | `global`、`ip`、`user`（user 取 `jwt_subject`） |
| `qps` / `burst` | 令牌桶参数 |

HTTP 由 `httpx` 自动挂载；Redis 模式需 `redis.enabled`，`app.InitInfra` 后自动 `BindRedis`。

### 熔断（`goeasy/breaker`）

- 配置：`governance.breaker`
- gRPC 出站：`grpcx.Client.Invoke` 已包装熔断 + 重试
- 业务可注入：`infra.RPC` / `app.Breaker().Execute(...)`

## 2. 缓存治理

### 实体 cache-aside

见 [实体缓存](../runtime/entity-cache.md)。

### 多级缓存

```yaml
cache:
  l1:
    enabled: true
    max_entries: 10000
    ttl: 1m
  null_ttl: 5m      # 空值缓存，防穿透
  ttl_jitter: 0.1   # TTL 抖动，防雪崩
```

| API | 用途 |
|-----|------|
| `cache.NewMultiLevel` | L1 内存 + L2 Redis |
| `cache.SetNull` | 写入空值标记 |
| `cache.JitterTTL` | 随机 TTL |
| `cache.GuardedGet` | singleflight，防击穿 |
| `infra.GuardedGet` | HTTPInfra 注入 |
| `infra.Locker` | Redis 分布式锁 |

仓储示例（防击穿）：

```go
raw, err := infra.GuardedGet.Do(ctx, infra.Cache, key, ttl, func(ctx context.Context) ([]byte, error) {
    // 查库并序列化
})
```

## 3. 数据库治理

| 能力 | 包 | 说明 |
|------|-----|------|
| 自动分页 | `pagination` | `Parse` / `MetaFrom` / `SQLLimitOffset` |
| SQL 日志 | `sqllog` | `dbx` 模板自动 `LogDuration` |
| 慢查询 | `observability.sql` | `slow_ms` 阈值，slog warn + Prometheus |

```yaml
observability:
  sql:
    enabled: true
    slow_ms: 200
```

N+1：列表接口保持单次 `COUNT` + `SELECT`（CLI 生成）；关联查询需在应用层 Batch Load（业务规范）。

## 4. 日志

```yaml
observability:
  logger:
    level: info
    format: json
    output: stdout
```

- 框架日志：`logger.New(cfg)` → slog JSON
- 访问日志：`httpx.SlogAccessLog`（含 `request_id`、`trace_id`、`client_ip`）
- Loki：采集 stdout JSON，按 `service` / `env` 标签过滤

## 5. 任务调度

```yaml
scheduler:
  enabled: true
  timezone: Asia/Shanghai
```

```go
// main 或 bootstrap
application.RegisterCron(scheduler.TaskSystem, "health_ping", "0 */5 * * * *", func(ctx context.Context) error {
    return nil
})
application.RegisterCron(scheduler.TaskBusiness, "sync_order", "0 0 2 * * *", syncOrders)
application.Run()
```

`infra.Cron` 也可在 bootstrap 注册业务任务。

## 6. 监控（Prometheus / Grafana）

```yaml
observability:
  metrics:
    enabled: true
    path: /metrics
```

指标：`goeasy_http_*`、`goeasy_sql_*`、`goeasy_cache_*`。

Tracing（OTLP 完整导出）：

```yaml
observability:
  trace:
    enabled: true
    exporter: otlp
    protocol: grpc
    endpoint: "127.0.0.1:4317"
    insecure: true
    sample_ratio: 1.0
```

详见 [17 观测栈](17-observability-stack.md)；Grafana 仪表盘见 [`docs/observability/grafana`](../observability/grafana/README.md)。

## 7. Outbox（MQ 事务一致性，默认关闭）

```yaml
mq:
  enabled: true
  outbox:
    enabled: false    # 默认关闭；开启需 database + migrate 000002_outbox
    poll_interval: 5s
```

**直发模式（默认）：**

```go
return infra.Outbox.Publish(ctx, topic, body)
```

**Outbox 模式（同事务）：**

```go
return infra.DB.Transaction(ctx, func(txCtx context.Context) error {
    if err := repo.Save(txCtx, agg); err != nil { return err }
    return infra.Outbox.PublishInTx(txCtx, topic, body)
})
```

框架自动注册系统 Cron `outbox_relay` 轮询投递。迁移：`migrations/*/000002_outbox.up.sql`。

## 8. 事件总线

| 类型 | 实现 |
|------|------|
| 进程内域事件 | `eventbus.Bus` → `infra.EventBus` |
| 跨服务集成事件 | `mq.Envelope` + NSQ，见 [13 MQ 业务接入](13-mq-business-integration.md) |

```go
infra.EventBus.Subscribe("order.created", func(ctx context.Context, evt eventbus.Event) error {
    return nil
})
_ = infra.EventBus.Publish(ctx, eventbus.Event{Name: "order.created", Payload: payload})
```

## 9. JWT / RBAC / API 签名

### JWT

`middleware.AdminAuth(infra)` / `MemberAuth(infra)`（模板已生成）。

### RBAC

```yaml
enterprise:
  casbin:
    enabled: true
    model: configs/casbin_model.conf
    policy: configs/casbin_policy.csv
```

```go
middleware.RequirePermission(infra, "sys_roles", "read")
```

### API 签名（RSA2）

规范见 monorepo 根目录 `api-sign.md`。

```yaml
enterprise:
  api_sign:
    enabled: true
    apps:
      my-app:
        public_key_pem: |
          -----BEGIN PUBLIC KEY-----
          ...
          -----END PUBLIC KEY-----
```

```go
open := engine.Group("/open/v1", middleware.OpenPlatformAuth(infra))
```

## 10. HTTPInfra 字段速查

| 字段 | 用途 |
|------|------|
| `Cache` / `Locker` / `GuardedGet` | 缓存与锁 |
| `JWT` / `MemberJWT` / `Casbin` / `APISign` | 安全 |
| `MQ` / `Outbox` | 消息发布（直发或事务发件箱） |
| `EventBus` / `Cron` | 进程内事件与定时任务 |
| `RPC` | gRPC 客户端池 |

## 下一步

- [17 观测栈](17-observability-stack.md)
- [19 项目配置 P0/P1](19-project-config-p0-p1.md)
- [HTTP 中间件](../runtime/http-middleware.md)
- [13 MQ 业务接入](13-mq-business-integration.md)
