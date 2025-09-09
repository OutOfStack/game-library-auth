-- +migrate Up
ALTER TABLE users
    ADD COLUMN email VARCHAR(255),
    ADD COLUMN email_verified BOOLEAN DEFAULT FALSE NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;

UPDATE users
SET email_verified = true
WHERE oauth_id IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_users_email;
ALTER TABLE users
    DROP COLUMN email,
    DROP COLUMN email_verified;