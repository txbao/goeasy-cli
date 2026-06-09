# bootstrap/snippets

本目录仅存放**可选运维说明**（非编译依赖）：

| 文件 | 来源 | 用途 |
|------|------|------|
| `<module>_grpc.md` | `add db proto` | grpcurl 联调示例 |
| `mqdemo_scheduler.md` | `add mqdemo` | 调度器接入说明 |

标准业务模块（`add module` / `add db crud` / `gen http`）**不再**生成 `*_wire.md`；装配说明见 [06 CLI 命令](../../../../docs/guide/06-goeasy-cli-commands.md) 与 `register_<domain>.go`。
