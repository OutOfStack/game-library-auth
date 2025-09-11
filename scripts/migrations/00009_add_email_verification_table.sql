-- +migrate Up
CREATE TABLE email_verifications (
    id                  UUID DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL,
    verification_code   VARCHAR(64),
    expires_at          TIMESTAMP NOT NULL,
    message_id          VARCHAR(64),
    verified_at         TIMESTAMP,
    date_created        TIMESTAMP DEFAULT NOW(),
    
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX email_verifications_user_id_pending_verification_idx
    ON email_verifications (user_id, date_created DESC)
    WHERE verified_at IS NULL AND verification_code IS NOT NULL;

-- +migrate Down
DROP TABLE IF EXISTS email_verifications;