-- +migrate Up
ALTER TABLE users
    DROP COLUMN avatar_url;

-- +migrate Down
ALTER TABLE users
    ADD COLUMN avatar_url VARCHAR(120);
