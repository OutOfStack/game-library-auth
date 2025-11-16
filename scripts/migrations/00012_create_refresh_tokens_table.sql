-- +migrate Up
CREATE TABLE refresh_tokens (
    id              UUID            DEFAULT gen_random_uuid(),
    user_id         UUID            NOT NULL,
    token_hash      VARCHAR(64)     NOT NULL    UNIQUE,
    expires_at      TIMESTAMPTZ     NOT NULL,
    date_created    TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX refresh_tokens_user_id_idx ON refresh_tokens (user_id);
CREATE INDEX refresh_tokens_token_idx ON refresh_tokens (token_hash);
CREATE INDEX refresh_tokens_expires_at_idx ON refresh_tokens (expires_at);

-- +migrate Down
DROP TABLE IF EXISTS refresh_tokens;
