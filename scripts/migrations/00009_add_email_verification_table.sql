-- +migrate Up
CREATE TABLE email_verifications (
    id                  UUID DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL,
    email               VARCHAR(255) NOT NULL,
    verification_code   VARCHAR(64) UNIQUE,
    expires_at          TIMESTAMP NOT NULL,
    message_id          VARCHAR(64),
    verified_at         TIMESTAMP,
    date_created        TIMESTAMP DEFAULT NOW(),
    
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_email_verifications_code ON email_verifications(verification_code);
CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);

-- +migrate Down
DROP TABLE IF EXISTS email_verifications;