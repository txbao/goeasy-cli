# 常见问题

## CLI

### `goeasy.exe: command not found`

在 `goeasy-cli` 目录未加入 PATH。请使用：

```bat
go build -o goeasy.exe .
.\goeasy.exe version
```

或使用 `go install` 后的全局命令 `goeasy`。

### 未指定 `--module` 警告

重新创建时显式传入：

```bat
.\goeasy.exe new mysvc --module github.com/org/mysvc
```

已生成项目可手工修改 `go.mod` 第一行并全局替换 import 前缀。

### `go mod tidy` 拉取 goeasy 失败

- 检查 `GOPRIVATE` 与企业 Git 权限
- monorepo 内确认 `replace` 指向存在的 `goeasy` 目录

## 编译

### `go build` 找不到 goeasy 包

确认 `go.mod` 中 `require` 与 `replace` 正确；在业务项目根执行 `go mod tidy`。

### interface 层 import infrastructure 报错（架构守卫）

这是预期约束。将 `new Repository` 移到 `bootstrap/wire.go`，Handler 只接收 Application。

### 路由 404：期望 `/api/v1/admin/system/roles` 但仍是 `/api/v1/admin/sys_roles`

| 现象 | 处理 |
|------|------|
| 旧扁平路由仍可访问，新分组路由 404 | 在 `configs/config.yaml` 增加 `codegen.group_prefixes`，对目标表 `add db crud --force` 重生成 HTTP 层 |
| `register_*.go` import 路径与 handler 目录不一致 | 升级 CLI 后 `add db crud --table <m> --force` 或 `add module <m> --group <g> --resource <r> --force` |
| 无分组配置 | 扁平路由为预期行为；见 [19 项目配置 §7](../19-project-config-p0-p1.md) |

## 运行

### 端口占用

修改 `configs/config.yaml` 中 `server.port`。

### `/health` 404

确认 `bootstrap.RegisterRoutes` 已注册 health 路由，且服务已 `RegisterHTTP`。

### `/healthz` 无响应

在配置中启用 `observability.health.enabled`（参见 `config.example.yaml`）。

### grpcurl list 只有 Reflection、没有业务 Service

| 现象 | 处理 |
|------|------|
| 仅有 `grpc.reflection.v1*` | gRPC 端口正常；业务 `Register*ServiceServer` 未生效或未实现 |
| `go build` 报找不到 `api/proto/gen/...` | `add db proto` 后缺 `*.pb.go`；执行 `goeasy-cli gen proto --file api/proto/<module>.proto`（需 protoc 与 go 插件） |
| `register_*.go: syntax error: unexpected .` | 旧版 CLI `ImportAlias` 生成 bug；升级 CLI 后 `add db crud --table <m> --force` 重生成 register，或手改 HTTP import 别名 |
| 编译通过但 list 仍无 Service | 确认 `internal/bootstrap/grpc.go` 含 `Register<Module>GRPC`；存在 `register_<module>_grpc.go` 与 `internal/interface/grpc/<domain>/<resource>/server.go` |
| `no required module .../grpc/sys_roles` | register import 与 domain 布局不一致 | `add db proto --table <m> --force` 或手改 import 为 `.../grpc/<domain>/<resource>`；见 [11 gRPC](../11-grpc-internal.md) |
| 有 `api/proto/*.proto` 和 `api/proto/gen/*.pb.go`，但无 gRPC 桩目录 | 命令顺序错误或旧版跳过桩生成 | 先 `add db crud --table <m>`，再 `add db proto --table <m>`（自动补齐）；见 [11 gRPC 项目集成](../11-grpc-internal.md) |
| 老项目 `server.go` 注册仍被注释 | `goeasy-cli add db proto --force` 覆盖桩，或手改与 [11 gRPC 项目集成](../11-grpc-internal.md) 一致 |
| `no required module .../command` | `add db proto` / `gen grpc` 生成的 handlers 与 `codegen.app_style` 不一致 | 升级 CLI 后 `add db proto --table <m> --force`；见 [11 gRPC](../11-grpc-internal.md) |

## 模板与升级

### 远端模板下载失败

CLI 会自动回退内嵌模板；日志中会有 fallback 提示。开发期建议 `--download=false`。

### 旧 demo 目录结构不一致

使用最新 CLI 重新 `goeasy new`，或对照 `goeasy-cli/demo6` 与 [04 项目结构](../04-project-structure.md) 手工迁移。

## 获取更多帮助

- [开发教程目录](../README.md)
- [goeasy 运行时规范](../../spec/goeasy-runtime-spec.md)
