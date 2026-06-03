# 持久化实现（Infrastructure）

领域仓储接口定义在 `internal/domain/<module>/repository.go`，实现放在本目录子包。

## 默认技术栈

| 组件 | 说明 |
|------|------|
| **sqlx** | 数据库访问（连接由 goeasy/database 管理） |
| **squirrel** | SQL 构建（推荐，避免拼接字符串） |

配置项 `database.orm: sqlx`（见 `configs/config.yaml`）。

## 扩展 ORM

| 目录 | 用途 |
|------|------|
| `sqlx/` | 默认实现占位 |
| `gorm/` | 可选 GORM 适配 |
| `ent/` | 可选 Ent 适配 |

业务模块示例：`persistence/health/repository_memory.go`（无 DB 时内存实现）。

## 依赖方向

- 实现类型可依赖 goeasy、sqlx、squirrel
- **不得** 在 domain 层 import 本目录
