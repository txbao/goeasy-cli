# dbx（sqlx + goqu）

| 组件 | 职责 |
|------|------|
| **sqlx** | 连接执行、`GetContext` / `ExecContext`、事务（由 goeasy/database 提供池） |
| **goqu** | 类型安全 SQL 构建（`Select` / `Insert` / `Update` / `Delete`） |

## 约定

- 表名使用 `dbx.TableName(config.database.table_prefix, "<module>")`，勿拼接用户输入。
- 仓储内用 `dbx.New(sqlxDB, driver)`，`driver` 与 `configs/config.yaml` 中 `database.driver` 一致。
- V1 占位 UPSERT 仅保证 **postgres**；mysql 需按方言调整 `OnConflict` / `OnDuplicate`。

## 依赖

```bat
go get github.com/doug-martin/goqu/v9 github.com/jmoiron/sqlx
```
