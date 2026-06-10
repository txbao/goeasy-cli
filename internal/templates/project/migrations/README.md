# 数据库迁移

使用 **goeasy-cli migrate** 管理 SQL；目录按 `database.driver` 区分方言。

> **不要用** Rust 的 `sqlx migrate add`（sqlx-cli）：它会把文件生成在 `migrations/` 根目录且格式不兼容。请始终使用 `goeasy-cli migrate create`。

## 目录结构

```text
migrations/
├── postgres/     # configs: database.driver: postgres（默认）
│   ├── 000001_init.up.sql
│   └── <version>_<name>.up.sql
└── mysql/        # configs: database.driver: mysql
    ├── 000001_init.up.sql
    └── ...
```

CLI 会根据 `configs/config.yaml` 中的 `database.driver` 自动选择 `migrations/postgres` 或 `migrations/mysql`。

也可显式覆盖：`goeasy-cli migrate up --migrations migrations/postgres`

## 版本表

`_sqlx_migrations` 由 golang-migrate 自动维护，列为 `version`（bigint）与 `dirty`（boolean）。  
`000001_init` 仅为首版本占位，不要手写旧版 `id/applied_at` 结构。

## 命名规范

```text
<version>_<name>.up.sql
<version>_<name>.down.sql
```

示例：`20260603223555_create_sys_roles.up.sql`

`migrate create` 会自动生成成对的 up/down 文件并放入当前 driver 子目录。

## 常用命令

```bat
goeasy-cli migrate up
goeasy-cli migrate status
goeasy-cli migrate version
goeasy-cli migrate create add_users_table
goeasy-cli migrate down --steps 1
```

启用迁移前请设置 `database.enabled: true` 与有效 `database.dsn`。

## 新建迁移

```bat
goeasy-cli migrate create add_users_table
```

会在 `migrations/postgres/`（或 `migrations/mysql/`）下生成：

```text
<timestamp>_add_users_table.up.sql
<timestamp>_add_users_table.down.sql
```

## Outbox 表（可选）

启用 `mq.outbox.enabled: true` 时执行：

```bat
goeasy-cli migrate up
```

将应用 `000002_outbox.up.sql` 创建默认表 `goeasy_outbox`（支持 postgres / mysql）。

表名可通过配置修改：

```yaml
mq:
  outbox:
    table: goeasy_outbox   # 可改为 outbox 等；须与迁移 SQL 中 CREATE TABLE 一致
```

若设置了 `database.table_prefix`，运行时实际表名为 `{table_prefix}{table}`，迁移 SQL 也需使用相同表名。

## 排错

| 现象 | 处理 |
|------|------|
| 迁移文件在 `migrations/` 根目录 | 移到 `migrations/postgres/` 或 `mysql/`，并改为 `.up.sql` / `.down.sql` 成对 |
| 误用 `sqlx migrate add` | 删除错误文件，改用 `goeasy-cli migrate create` |
| `migrate status` 报路径语法错误 | 升级 goeasy-cli 到含修复的版本后重试 |
| Outbox 启用后表不存在 | 确认已 `migrate up` 且表名与 `mq.outbox.table` 一致 |
