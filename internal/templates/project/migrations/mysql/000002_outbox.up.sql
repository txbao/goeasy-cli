-- Outbox 事务发件箱（mq.outbox.enabled=true 时需执行）
CREATE TABLE IF NOT EXISTS _outbox (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id    VARCHAR(64) NOT NULL UNIQUE,
    topic       VARCHAR(128) NOT NULL,
    payload     BLOB NOT NULL,
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at     TIMESTAMP NULL,
    INDEX idx__outbox_status (status, id)
);
