# 业务操作日志（audit.Recorder）

> 框架提供 Port 与工具，**业务表与 List API 由业务服务实现**（如 `base_operation_logs`）。

## 两条审计线

| 能力 | 配置 | 用途 |
|------|------|------|
| `audit.Logger` | `observability.audit.enabled` | 运维 JSON stdout |
| `audit.Recorder` | 业务注入 | DB 持久化操作日志 |

## 配置

```yaml
observability:
  audit:
    enabled: false          # JSON 运维审计
    async_enabled: false    # Recorder 异步写入
    buffer_size: 256
    mask_phone: true        # 省略时默认 true
    mask_login_id: true
    sensitive_keys: []      # 额外敏感字段
```

## 注入 Recorder

```go
application := app.New(cfg)
application.SetAuditRecorder(operationlogs.NewPGRecorder(dbx))
application.RegisterHTTP(bootstrap.RegisterRoutes)
application.Run()
```

bootstrap 中使用 `infra.AuditRecorder`：

```go
func RegisterRoutes(e *gin.Engine, infra app.HTTPInfra) error {
    admin := e.Group("/api/v1/admin")
    admin.Use(middleware.AdminAuth(infra))
    admin.Use(httpx.InjectOperatorContext())
    // ...
}
```

## Application 埋点

```go
opts := audit.RedactOptionsFromCfg(cfg.Observability.Audit)
before, after, _ := audit.BuildChangeSummary(old, new, []string{"customerName", "status"}, audit.SummaryOptions{Redact: opts})

_ = a.recorder.Record(ctx, contextx.OperatorFrom(ctx), audit.Entry{
    ModuleCode:    "customer",
    ActionType:    "update",
    ObjectType:    "customer",
    ObjectID:      strconv.FormatInt(id, 10),
    ObjectName:    name,
    CustomerID:    customerID,
    Result:        "1",
    BeforeSummary: before,
    AfterSummary:  after,
})
```

## 中间件顺序

```
requestID → trace → RequireJWT → InjectOperatorContext → 业务路由
```

引擎级已挂载 `requestID` 与 `trace`；JWT 与 `InjectOperatorContext` 在**鉴权路由组**上挂载。

## 相关文档

- [HTTP 中间件](http-middleware.md)
- [goeasy/audit/README.md](../../../goeasy/audit/README.md)

## CLI 生成

对已有模块启用操作日志埋点桩：

```bat
goeasy-cli add crud customers --audit --force
```

生成内容：
- `Application` 构造函数注入 `audit.Recorder`（bootstrap 传入 `infra.AuditRecorder`）
- `Create` / `Update` / `Delete` 方法内注释桩
- `register_<domain>.go` 鉴权路由组挂载 `httpx.InjectOperatorContext()`
