-- +migrate Up
CREATE TABLE user_oauth_links (
    id              UUID         DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    oauth_provider  VARCHAR(32)  NOT NULL,
    oauth_id        VARCHAR(128) NOT NULL,
    
    PRIMARY KEY(id),
    UNIQUE(oauth_provider, oauth_id)
);

INSERT INTO user_oauth_links (id, user_id, oauth_provider, oauth_id)
    SELECT gen_random_uuid(), id, oauth_provider, oauth_id
    FROM users
    WHERE oauth_provider IS NOT NULL AND oauth_id IS NOT NULL;

DROP INDEX IF EXISTS idx_users_oauth_provider_oauth_id;
ALTER TABLE users 
    DROP COLUMN oauth_provider, 
    DROP COLUMN oauth_id;

-- +migrate Down
ALTER TABLE users
    ADD COLUMN oauth_provider VARCHAR(32),
    ADD COLUMN oauth_id VARCHAR(128);

UPDATE users 
SET oauth_provider = uol.oauth_provider, 
    oauth_id = uol.oauth_id
FROM user_oauth_links uol 
WHERE users.id = uol.user_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth_provider_oauth_id ON users(oauth_provider, oauth_id);

DROP TABLE IF EXISTS user_oauth_links;
