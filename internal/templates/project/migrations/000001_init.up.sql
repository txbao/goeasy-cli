-- 迁移版本表（goeasy migrate 管理）
CREATE TABLE IF NOT EXISTS _sqlx_migrations (
    id          SERIAL PRIMARY KEY,
    version     VARCHAR(64) NOT NULL UNIQUE,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
