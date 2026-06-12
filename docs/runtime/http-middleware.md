# HTTP 中间件

> 本文描述 **goeasy 运行时库**能力；业务项目接入见 [guide/](../guide/) 对应章节。统一错误响应见 [HTTP 响应与错误日志](http-response.md)。

## 引擎级（goeasy/httpx）

`app.New(cfg)` 创建的 Gin 引擎默认挂载（按配置）：

| 中间件 | 配置 |
|--------|------|
| Recovery | 始终 |
| CORS | `http.cors.enabled` |
| RequestID | 始终 |
| Trace | `observability.trace` |
| 限流 | `governance.limiter`（local/redis/both，多维 ip/user） |
| Metrics | `observability.metrics` |
| AccessLog | slog JSON（`observability.logger`） |
| 健康探针 | `observability.health` → `/healthz` |

### CORS 配置

```yaml
http:
  cors:
    enabled: true
    allow_origins: ["https://admin.example.com"]
    allow_methods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"]
    allow_headers: ["Content-Type", "Authorization", "X-Request-ID"]
    allow_credentials: true
```

实现：`httpx.CORSMiddleware(cfg.HTTP.CORS)`。

## JWT（goeasy/jwt + httpx）

两套独立配置：

| 配置块 | 用途 | `app.App` 字段 |
|--------|------|----------------|
| `enterprise.jwt` | 管理后台 | `JWT` |
| `enterprise.member_jwt` | H5 会员 | `MemberJWT` |

`HTTPInfra` 同样携带 `JWT`、`MemberJWT`、`Casbin`，供 bootstrap 装配路由组。

校验中间件：

```go
httpx.RequireJWT(infra.JWT, "Authorization")
httpx.RequireJWT(infra.MemberJWT, "Authorization")
```

- Header：`Authorization: Bearer <token>`
- 成功后：`c.Set("jwt_subject", claims.Subject)`

`jwt` / `member_jwt` 的 `enabled: false` 时对应 `Token` 为 nil，中间件返回 503。

## 项目层

CLI 模板生成 `internal/interface/http/middleware`：

- `AdminAuth(infra)` — 管理端
- `MemberAuth(infra)` — H5（需登录）

公开 H5（不挂任何鉴权）由 CLI `--public h5` 生成：`engine.Group("/api/v1/h5")`，无 `MemberAuth`。与 `member_jwt.enabled: false`（503）不同。

```bat
goeasy-cli add crud products --client admin --client h5 --public h5
```

### Casbin RBAC

```go
httpx.RequireCasbin(infra.Casbin, "sys_roles", "read")
// 或项目模板 middleware.RequirePermission(infra, "sys_roles", "read")
```

### 开放平台 API 签名（RSA2）

```go
httpx.RequireAPISign(infra.APISign)
// 或 middleware.OpenPlatformAuth(infra)
```

规范见 monorepo `api-sign.md`；配置 `enterprise.api_sign`。

## 边界

goeasy **不**实现用户表、角色表、登录接口；仅提供 JWT 引擎与 Gin 中间件。

## 下一步

- [项目配置清单（P0/P1）](../guide/19-project-config-p0-p1.md)
- [05 运行时总览](../guide/05-goeasy-runtime.md)
