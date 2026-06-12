# HTTP 统一响应与错误日志

> 本文描述 **goeasy 运行时** `response` 包能力；业务项目接入见 [19 项目配置](../guide/19-project-config-p0-p1.md)。

对齐 CSDS `10-api-error-code-spec.md` §9：**HTTP Status** 与 **body.code（6 位业务码）** 分离；服务端错误自动结构化日志。

## 响应结构

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

| 场景 | HTTP Status | body.code |
|------|-------------|-----------|
| 成功 | 200 | `0` |
| 业务错误 | 4xx/5xx（按语义） | 6 位业务码（如 `100001`） |
| 未映射内部错误 | 500 | `500001` |

## API 速查

```go
import (
    zresp "github.com/txbao/goeasy/response"
    zerr  "github.com/txbao/goeasy/errors"
)

// 成功
zresp.Success(c, data)

// 标准业务错误（推荐）
zresp.FailBiz(c, http.StatusBadRequest, int(zerr.BizCodeParamInvalid), "参数错误")
zresp.FailBiz(c, http.StatusBadRequest, int(zerr.BizCodeParamFormat), zvalid.Format(err))

// 未映射内部错误（prod 脱敏，日志保留完整 err）
zresp.FailInternal(c, err)

// 从 error 推断：BizError → CodedError → FailInternal
zresp.FailErr(c, err)

// 兼容旧代码（deprecated：body.code = httpCode）
zresp.Fail(c, http.StatusNotFound, "not found")
```

## 全局保留业务码（框架）

| 码 | 常量 | 用途 |
|----|------|------|
| 100001 | `BizCodeParamInvalid` | 参数错误 |
| 100002 | `BizCodeParamFormat` | 校验失败 |
| 200002 | `BizCodeAuthMissing` | 缺少/无效凭证 |
| 210001 | `BizCodeForbidden` | 无权限 |
| 500001 | `BizCodeInternal` | 内部错误兜底 |
| 503001 | `BizCodeServiceUnavail` | JWT/Casbin 等未启用 |

领域模块码（如订单 `301xxx`）在业务项目 `mapErr` 中维护，不下沉到 goeasy。

## 500 错误日志

当 `httpCode >= 500` 或 `bizCode >= 500000` 时，`Fail` / `FailBiz` / `FailInternal` 自动 `slog.Error`，字段含：

`method`、`path`、`request_id`、`trace_id`、`user_id`、`biz_code`、`err`

由 `httpx.NewEngineWith` 启动时注入 `observability.logger` 与 `env`。

访问日志（`http_access`）在 `status >= 500` 时升为 `Warn`，与 Error 日志互补。

## FailInternal 脱敏

| 环境 | 响应 msg | 日志 |
|------|----------|------|
| `env=dev` | `err.Error()` | 完整 err |
| `env=prod` | `服务内部错误` | 完整 err |

可选配置 `observability.http.expose_error_detail: true` 覆盖 prod 脱敏（仅联调）。

## 配置

```yaml
observability:
  logger:
    level: info
    format: json
    output: stdout
  http:                          # 可选，整段可省略
    log_server_errors: true      # default: true
    expose_error_detail: false   # nil=按 env
```

现有项目仅需 `observability.logger` + `env` 即可工作，**不强制**改 `config.yaml`。

## 业务项目迁移

1. `go get github.com/txbao/goeasy@<新版本>`
2. 删除临时 `internal/interface/http/respond/` 包装（若有）
3. Handler：`mapErr` 后 `FailBiz`；default 分支 `FailInternal`
4. 参数校验：`FailBiz(c, 400, 100001/100002, ...)`
