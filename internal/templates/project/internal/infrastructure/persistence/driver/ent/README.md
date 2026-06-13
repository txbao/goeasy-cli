# 可选 ORM：Ent

默认规范为 **sqlx + goqu**（`internal/infrastructure/shared/dbx/`）。若引入 Ent，建议新建项目侧 schema 与生成物目录，勿与 domain 混放。
