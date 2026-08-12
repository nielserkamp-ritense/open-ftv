BEGIN;

ALTER TABLE decision
    DROP COLUMN IF EXISTS attributes;

ALTER TABLE decision
    RENAME COLUMN timestamp TO timestamp_ms;

COMMIT;
