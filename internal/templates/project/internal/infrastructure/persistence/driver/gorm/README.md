# 可选 ORM：GORM

默认规范为 **sqlx + goqu**（`internal/infrastructure/shared/dbx/`）。若引入 GORM，建议新建 `internal/infrastructure/persistence/driver/gorm/` 下的项目侧封装，勿与 domain 混放。
