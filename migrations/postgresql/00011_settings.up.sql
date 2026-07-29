BEGIN;

CREATE TABLE IF NOT EXISTS settings
(
    singleton_id    BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton_id),
    header_title    VARCHAR(200) NOT NULL,
    header_color    VARCHAR(20)  NOT NULL,
    title_color     VARCHAR(20)  NOT NULL,
    logo            BYTEA,
    logo_media_type VARCHAR(80),
    created         TIMESTAMP    NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    created_by      VARCHAR(80)  NOT NULL DEFAULT '*SYSTEM*',
    updated         TIMESTAMP,
    updated_by      VARCHAR(80)
);

COMMIT;
