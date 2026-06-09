# 持久化实现（Infrastructure）

领域仓储接口在 `internal/domain/<module>/repository.go`，**仓储实现**在 `repository/<module>/`（如 `repository/sys_roles/repository_pg.go`）。

## 默认技术栈（sqlx + goqu）

| 组件 | 职责 |
|------|------|
| **sqlx** | 连接执行、`GetContext` / `ExecContext`（池由 goeasy/database 管理） |
| **goqu** | SQL 构建（`Select` / `Insert` / `Update` / `Delete`） |

公共包：`persistence/dbx`（方言、执行辅助、表名 `TableName(prefix, module)`）。

配置见 `configs/config.yaml`：

```yaml
database:
  orm: sqlx
  driver: postgres
  table_prefix: ""   # 可选，如 ge_
```

## 扩展 ORM

| 目录 | 用途 |
|------|------|
| `dbx/` | 默认 sqlx + goqu |
| `repository/<module>/` | 各模块 PG / 内存仓储 |
| `driver/sqlx/` | sqlx 说明（已统一到 dbx） |
| `driver/gorm/` | 可选 GORM |
| `driver/ent/` | 可选 Ent |

实体缓存（P1）：`redis.enabled` + `cache.enabled` 后，`FindByID` 使用 key `{key_prefix}:{module}:id:{id}`（逻辑模块名）；`Update`/`Delete` 失效缓存。配置说明见 monorepo `goeasy-cli/docs/guide/09-project-config-p0-p1.md`。

## 依赖方向

- 实现可依赖 goeasy、sqlx、goqu、dbx
- **不得** 在 domain 层 import 本目录
