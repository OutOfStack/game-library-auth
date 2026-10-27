-- +migrate Up
CREATE TABLE notification_outbox (
    id UUID PRIMARY KEY,
    reference_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    payload JSONB,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'cancelled', 'expired', 'failed')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    next_attempt_at TIMESTAMPTZ NOT NULL,
    locked_by TEXT,
    locked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    provider_message_id TEXT,
    last_error TEXT,
    UNIQUE (kind, reference_id),
    CHECK ((locked_by IS NULL) = (locked_at IS NULL)),
    CHECK ((status = 'pending' AND payload IS NOT NULL)
        OR (status <> 'pending' AND payload IS NULL AND locked_by IS NULL))
);
CREATE INDEX notification_outbox_due_idx ON notification_outbox (next_attempt_at, created_at)
    WHERE status = 'pending';
CREATE INDEX notification_outbox_expiry_idx ON notification_outbox (expires_at)
    WHERE status = 'pending';
CREATE INDEX notification_outbox_retention_idx ON notification_outbox (updated_at)
    WHERE status <> 'pending';

-- +migrate Down
DROP TABLE notification_outbox;
