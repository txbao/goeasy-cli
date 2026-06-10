# HTTP 中间件（项目层）

框架已在 `goeasy/httpx` 注入：Recovery、CORS（`http.cors`）、RequestID、Trace、限流、Metrics、slog 访问日志。

本目录：

| 文件 | 用途 |
|------|------|
| `admin_auth.go` | 管理后台 JWT |
| `member_auth.go` | H5 会员 JWT |
| `require_permission.go` | Casbin RBAC |
| `open_platform_auth.go` | 开放平台 RSA2 验签 |

详见 `goeasy-cli/docs/guide/16-runtime-capabilities.md`（鉴权与 RBAC）。
JWT / CORS 配置见 `goeasy-cli/docs/guide/19-project-config-p0-p1.md`。
引擎级中间件见 `goeasy-cli/docs/runtime/http-middleware.md`。

本目录提供**业务路由组**鉴权工厂：

| 函数 | 配置 | 典型路由组 |
|------|------|------------|
| `AdminAuth(infra)` | `enterprise.jwt` | `/api/v1/admin` |
| `MemberAuth(infra)` | `enterprise.member_jwt` | `/api/v1/h5` |

示例（在 `internal/bootstrap/register_<module>.go` 中，`infra` 由 `RegisterRoutes(engine, infra)` 注入）：

```go
// add db crud 生成的 register_*.go 默认（--client admin）：
admin := engine.Group("/api/v1/admin", middleware.AdminAuth(infra))
// 代码在 internal/interface/http/admin/<module>/

// 同时生成 H5（--client h5，需会员登录）：
h5 := engine.Group("/api/v1/h5", middleware.MemberAuth(infra))

// 公开 H5（--client h5 --public h5，无鉴权中间件）：
h5 := engine.Group("/api/v1/h5")
// 生成：goeasy-cli add crud <name> --client admin --client h5 --public h5
```

**注意**：`member_jwt.enabled: false` 时 `MemberAuth` 返回 503，不等于公开 API；公开 H5 须用 `--public h5` 生成无中间件路由组。

勿在 `internal/interface/http/<module>/router.go` 使用 `infra`（该文件只有 `RegisterRoutes(r, h)`）。

Casbin 策略与菜单权限在业务层实现，不放在 goeasy 框架内。