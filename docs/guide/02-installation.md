# 02 安装

## 前置条件

| 工具 | 最低版本 | 用途 |
|------|----------|------|
| Go | 1.22+ | 编译与运行 |
| Git | 任意近期版本 | 拉取框架与业务仓库 |

可选（生成 gRPC 契约时）：

| 工具 | 用途 |
|------|------|
| protoc | 编译 `.proto` |

## 方式一：本地编译 goeasy-cli（推荐 monorepo 开发）

在框架仓库中进入脚手架目录并编译：

```bat
cd goeasy-cli
go build -o goeasy.exe .
```

验证：

```bat
.\goeasy.exe version
```

在 Git Bash 中必须使用 `./goeasy.exe`，不能直接输入 `goeasy.exe`（当前目录不在 PATH 中）。

## 方式二：go install 安装

```bat
go install github.com/txbao/goeasy-cli@latest
```

安装后命令名为 **`goeasy`**（不是 `goeasy-cli`）：

```bat
goeasy version
```

确保 `%USERPROFILE%\go\bin` 已加入系统 PATH。

## 业务项目引用 goeasy 运行时

创建项目时 CLI 会执行 `go mod init <你的 --module>`；在 monorepo 内还会自动执行：

```text
go mod edit -replace=github.com/txbao/goeasy=<本地 goeasy 路径>
```

进入项目目录后执行 `go mod tidy`，会根据源码 import 写入 `require`（默认运行时模块为 `github.com/txbao/goeasy`，可通过 `--goeasy-module` 或环境变量 `GOEASY_MODULE` 覆盖）。

详见 [09 Monorepo 与 Module](09-monorepo-and-modules.md)。

## IDE 建议

- 使用 Go 官方插件或 Cursor / VS Code Go 扩展
- 将 `framework` 根目录或业务服务目录作为工作区打开

## 下一步

[03 60 秒上手](03-quickstart-60s.md)
