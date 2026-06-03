# 数据库迁移

使用 **goeasy migrate** 管理本目录下的 SQL 文件。

## 命名规范

```text
<version>_<name>.up.sql
<version>_<name>.down.sql
```

示例：`000001_init.up.sql` / `000001_init.down.sql`

## 版本表

首次迁移会创建 `_sqlx_migrations`，记录已应用的 `version`。

## 常用命令

在项目根目录：

```bat
goeasy migrate up -f configs/config.yaml
goeasy migrate status
goeasy migrate down
goeasy migrate create add_users_table
```

或使用 Makefile：

```bat
make migrate-up CONFIG=configs/config.example.yaml
```

启用迁移前请在配置中设置 `database.enabled: true` 与有效 `database.dsn`。
