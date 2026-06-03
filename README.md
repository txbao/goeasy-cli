# goeasy-cli

GoEasy DDD Lite 官方脚手架（开发期工具，不参与运行时）。

**开发教程**：[docs/guide](docs/guide/README.md)

## 安装

```bat
# 1. 安装 goeasy-cli（代码生成工具）
go install github.com/txbao/goeasy-cli@latest


```

## 创建项目

生成项目直接依赖 **goeasy 运行时**：

```bat
goeasy-cli new demo --module github.com/demo/demo --download=false
cd demo
go mod tidy
go run ./cmd/service
```

monorepo 内会自动 `replace` 本地 `../goeasy`。

别名：`goeasy-cli init demo ...`

### 模板变体

```bat
goeasy-cli new auth-svc --template auth --module github.com/demo/auth
goeasy-cli new pay-svc --template payment --module github.com/demo/pay
goeasy-cli new app --template monolith
goeasy-cli new sys-svc --template system
```

## 生成模块

在项目根目录执行：

```bat
goeasy-cli add module user
goeasy-cli add crud user --force
goeasy-cli add repository order
goeasy-cli add proto user
goeasy-cli add event user-created
goeasy-cli add aggregate order
```

## 升级

```bat
goeasy-cli upgrade template
goeasy-cli upgrade framework
```

## Flags（new/init）

| Flag | 默认 | 说明 |
|------|------|------|
| `--module` | 项目名 | Go module |
| `--template` | default | default / monolith / auth / system / payment |
| `--download` | false | 远端 zip（失败回退 embed） |
| `--output` | `.` | 输出父目录 |
| `--goeasy-replace` | 自动 | 本地 goeasy replace 路径 |

## 生成目录（DDD Lite）

```text
cmd/service/
internal/domain|app|interface/http|infrastructure|bootstrap|observer
configs/  api/  deploy/
```

## 开发指引

- [uidedocs/guide/README.md](docs/guide/README.md)
