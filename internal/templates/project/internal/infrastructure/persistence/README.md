# Infrastructure 持久化说明

> 业务仓储按 **限界上下文** 组织，技术底座在 `shared/`。详细布局见 goeasy-cli `docs/guide/04-project-structure.md`。

## 目录约定

```text
internal/infrastructure/
├── shared/dbx/                         # sqlx + goqu 公共包（唯一 dbx 入口）
├── <domain>/persistence/<resource>/      # 业务仓储（repository_pg / repository_memory）
├── health/persistence/health/          # 健康检查示范（内存仓储）
├── rpc/                                # 跨服务 gRPC ACL
├── mq/、cache/、client/                # 横切集成
└── persistence/driver/                 # 可选 ORM 扩展说明（gorm/ent）
```

## 默认技术栈

| 组件 | 职责 |
|------|------|
| **sqlx** | 连接执行（池由 goeasy/database 管理） |
| **goqu** | SQL 构建 |
| **shared/dbx** | 方言、`TableName(prefix, module)`、执行辅助 |

领域仓储接口在 `internal/domain/<domain>/<resource>/repository.go`，实现在 `internal/infrastructure/<domain>/persistence/<resource>/`。

## 依赖方向

- 仓储实现可依赖 goeasy、sqlx、goqu、`shared/dbx`
- **不得** 在 domain 层 import infrastructure
