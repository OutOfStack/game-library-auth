-- +migrate Up
ALTER TABLE email_verifications
    ADD COLUMN unsubscribe_token VARCHAR(255),
    DROP COLUMN expires_at,
    ALTER COLUMN verified_at SET DATA TYPE TIMESTAMPTZ USING verified_at AT TIME ZONE 'UTC',
    ALTER COLUMN date_created SET DATA TYPE TIMESTAMPTZ USING date_created AT TIME ZONE 'UTC';

CREATE INDEX email_verifications_unsubscribe_token_idx 
    ON email_verifications (unsubscribe_token) WHERE unsubscribe_token IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS email_verifications_unsubscribe_token_idx;

ALTER TABLE email_verifications
    DROP COLUMN unsubscribe_token,
    ADD COLUMN expires_at TIMESTAMP,
    ALTER COLUMN verified_at SET DATA TYPE TIMESTAMP,
    ALTER COLUMN date_created SET DATA TYPE TIMESTAMP;
