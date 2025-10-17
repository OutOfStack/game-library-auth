-- +migrate Up
CREATE TABLE email_unsubscribes (
    id              UUID DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    date_created    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id)
);

-- +migrate Down
DROP TABLE IF EXISTS email_unsubscribes;
