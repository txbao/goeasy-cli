# 实体缓存（Redis）

> 本文描述 **goeasy 运行时库**能力；业务项目接入见 [guide/](../guide/) 对应章节。

goeasy 提供仓储层 **cache-aside** 契约，不包含业务实体序列化规则以外的逻辑。

## 配置（`configs/config.yaml`）

```yaml
redis:
  enabled: true
  addr: "127.0.0.1:6379"
  key_prefix: mysvc      # 空则使用 app_name
  default_ttl: 168h

cache:
  enabled: true          # 必须同时 redis.enabled
  entity_ttl: 168h       # 空则用 redis.default_ttl
```

| 方法 | 说明 |
|------|------|
| `config.RedisKeyPrefix()` | 项目 key 前缀 |
| `config.EntityCacheTTL()` | 解析后的 TTL |
| `config.EntityCacheEnabled()` | `redis.enabled && cache.enabled` |

## Key 规范

逻辑**模块名**（如 `sys_roles`），非物理表名：

```text
{key_prefix}:{module}:id:{id}
示例：mysvc:sys_roles:id:1
```

包：`goeasy/cachekey` — `EntityKey`、`GetEntityBytes`、`SetEntityBytes`、`DeleteEntity`。

## HTTPInfra 注入

`app.Run()` 传入 bootstrap：

| 字段 | 含义 |
|------|------|
| `Cache` | `cache.Cache`（Redis 或 Noop） |
| `RedisKeyPrefix` | 前缀字符串 |
| `EntityCacheEnabled` | 是否读写缓存 |
| `EntityCacheTTL` | `time.Duration` |

## cache 接口（P1）

`goeasy/cache.Cache`：`Get`/`Set`、`GetBytes`/`SetBytes`、`Del`；未命中为 `cache.ErrNotFound`。

## 业务仓储

由 **goeasy-cli** `add db crud` 生成的 `repository_pg.go` 实现：

- `FindByID`：先 Redis JSON 行，miss 查库并回填
- `Update` / `Delete`：写库前 `DeleteEntity`
- `Create`：可选写入缓存（返回 id 后）

List 分页默认不缓存。

## 缓存治理（运行层）

| 能力 | API |
|------|-----|
| 多级缓存 L1+L2 | `cache.NewMultiLevel`（`cache.l1.enabled`） |
| 防穿透 | `cache.SetNull` + `null_ttl` |
| 防击穿 | `infra.GuardedGet.Do`（singleflight） |
| 防雪崩 | `cache.JitterTTL` + `ttl_jitter` |
| 分布式锁 | `infra.Locker` / `cache.NewRedisLocker` |

详见 [16 运行时能力](../guide/16-runtime-capabilities.md)。

## 下一步

- [HTTP 中间件](http-middleware.md)
- [05 运行时总览](../guide/05-goeasy-runtime.md)
