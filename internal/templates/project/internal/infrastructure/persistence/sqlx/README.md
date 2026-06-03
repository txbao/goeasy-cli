# sqlx 持久化

在此包实现 `domain.Repository`，使用：

- `github.com/jmoiron/sqlx`（或通过 goeasy/database 暴露的 DB）
- `github.com/Masterminds/squirrel` 构建查询

示例：`repository_sqlx.go` 中 `Load` / `Save` 调用 squirrel 生成 SQL。
