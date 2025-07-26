-- +migrate Up
ALTER TABLE users
    ADD COLUMN oauth_provider VARCHAR(32),
    ADD COLUMN oauth_id VARCHAR(128);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth_provider_oauth_id ON users(oauth_provider, oauth_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_users_oauth_provider_oauth_id;
ALTER TABLE users
    DROP COLUMN oauth_provider,
    DROP COLUMN oauth_id;
