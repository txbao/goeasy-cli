# Grafana 仪表盘

基于 goeasy Prometheus 指标，可直接导入 Grafana。

## 文件

| 文件 | 说明 |
|------|------|
| `goeasy-http-overview.json` | HTTP QPS、P95 延迟、状态码、5xx 率 |
| `goeasy-sql-cache.json` | SQL QPS、慢查询、缓存命中率 |

## 导入步骤

1. Grafana → Dashboards → Import
2. Upload JSON 或粘贴内容
3. 选择 Prometheus 数据源
4. 变量 `service` 选择你的 `app_name`

## 前置条件

```yaml
observability:
  metrics:
    enabled: true
    path: /metrics
```

Prometheus scrape 示例见 [17 观测栈指南](../../guide/17-observability-stack.md)。
