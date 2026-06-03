# goesy-cli

GoEasy DDD Lite 官方脚手架（开发期工具，不参与运行时）。

**开发教程**：[docs/guide](docs/guide/README.md)

## 安装

```bat
go install github.com/txbao/goeasy-cli@latest
```

## 创建项目

生成项目直接依赖 **goesy 运行时**：

```bat
goesy new demo3 --module github.com/demo/demo3 --download=false
cd demo3
go mod tidy
go run ./cmd/service
```

monorepo 内会自动 `replace` 本地 `../goesy`。

别名：`goesy init demo ...`

### 模板变体

```bat
goesy new auth-svc --template auth --module github.com/demo/auth
goesy new pay-svc --template payment --module github.com/demo/pay
goesy new app --template monolith
goesy new sys-svc --template system
```

## 生成模块

在项目根目录执行：

```bat
goesy add module user
goesy add crud user --force
goesy add repository order
goesy add proto user
goesy add event user-created
goesy add aggregate order
```

## 升级

```bat
goesy upgrade template
goesy upgrade framework
```

## Flags（new/init）

| Flag | 默认 | 说明 |
|------|------|------|
| `--module` | 项目名 | Go module |
| `--template` | default | default / monolith / auth / system / payment |
| `--download` | false | 远端 zip（失败回退 embed） |
| `--output` | `.` | 输出父目录 |
| `--goesy-replace` | 自动 | 本地 goesy replace 路径 |

## 生成目录（DDD Lite）

```text
cmd/service/
internal/domain|app|interface/http|infrastructure|bootstrap|observer
configs/  api/  deploy/
```

## 开发指引

- [uidedocs/guide/README.md](docs/guide/README.md)
