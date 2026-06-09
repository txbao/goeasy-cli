# 17 观测栈（Prometheus + Loki + Tempo + Grafana）

goeasy 观测能力下沉在运行层，项目只需改 `configs/config.yaml`。

## 1. Metrics（Prometheus）

```yaml
observability:
  metrics:
    enabled: true
    path: /metrics
```

**Prometheus scrape 示例：**

```yaml
scrape_configs:
  - job_name: goeasy
    metrics_path: /metrics
    static_configs:
      - targets: ["127.0.0.1:8080"]
        labels:
          service: demo
          env: dev
```

**指标：**

| 指标 | 说明 |
|------|------|
| `goeasy_http_requests_total` | HTTP 请求计数 |
| `goeasy_http_duration_ms` | HTTP 延迟直方图 |
| `goeasy_sql_queries_total` | SQL 执行计数 |
| `goeasy_sql_slow_total` | 慢 SQL 计数 |
| `goeasy_cache_hit_total` / `miss_total` | 缓存命中/未命中 |

Grafana 仪表盘：[`docs/observability/grafana`](../observability/grafana/README.md)

## 2. Logs（Loki）

```yaml
observability:
  logger:
    level: info
    format: json
    output: stdout
```

访问日志字段：`service`、`env`、`method`、`path`、`status`、`latency_ms`、`request_id`、`trace_id`。

Promtail / Alloy 采集 stdout JSON 到 Loki，按 `service` 过滤。

## 3. Tracing（OTLP → Tempo/Jaeger）

```yaml
observability:
  trace:
    enabled: true
    exporter: otlp        # noop=仅本地 trace_id；otlp=导出
    protocol: grpc        # grpc | http
    endpoint: "127.0.0.1:4317"
    insecure: true
    sample_ratio: 1.0
```

**OTel Collector 最小配置：**

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
exporters:
  otlp/tempo:
    endpoint: tempo:4317
    tls:
      insecure: true
service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [otlp/tempo]
```

HTTP 中间件自动创建 span；日志 `trace_id` 与 Tempo 一致，可在 Grafana 做 trace ↔ log 关联。

## 4. Grafana 联调清单

| 步骤 | 操作 |
|------|------|
| 1 | 开启 `metrics.enabled`，确认 `GET /metrics` 有数据 |
| 2 | Prometheus scrape 业务端口 |
| 3 | 导入 `goeasy-http-overview.json`、`goeasy-sql-cache.json` |
| 4 | 开启 `trace.exporter=otlp`，Collector 转发到 Tempo |
| 5 | Loki 采集 JSON 日志，Explore 用 `trace_id` 跳转 Tempo |

## 下一步

- [16 运行时能力清单](16-runtime-capabilities.md)
- [13 MQ 业务接入](13-mq-business-integration.md)
