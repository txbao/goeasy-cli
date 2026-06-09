# 数据库迁移

使用 **goeasy migrate** 管理 SQL；目录按 `database.driver` 区分方言。

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

也可显式覆盖：`goeasy migrate up --migrations migrations/postgres`

## 版本表

`_sqlx_migrations` 由 golang-migrate 自动维护，列为 `version`（bigint）与 `dirty`（boolean）。  
`000001_init` 仅为首版本占位，不要手写旧版 `id/applied_at` 结构。

## 命名规范

```text
<version>_<name>.up.sql
<version>_<name>.down.sql
```

示例：`20260603223555_create_sys_roles.up.sql`

## 常用命令

```bat
goeasy migrate up
goeasy migrate status
goeasy migrate create add_users_table
```

启用迁移前请设置 `database.enabled: true` 与有效 `database.dsn`。

## Outbox 表（可选）

启用 `mq.outbox.enabled: true` 时执行：

```bat
goeasy migrate up
```

将应用 `000002_outbox.up.sql` 创建 `goeasy_outbox` 表（支持 postgres / mysql）。
