-- Outbox 事务发件箱（mq.outbox.enabled=true 时需执行）
CREATE TABLE IF NOT EXISTS _outbox (
    id          BIGSERIAL PRIMARY KEY,
    event_id    VARCHAR(64) NOT NULL UNIQUE,
    topic       VARCHAR(128) NOT NULL,
    payload     BYTEA NOT NULL,
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at     TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx__outbox_status ON _outbox (status, id);
