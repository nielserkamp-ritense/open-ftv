BEGIN;

ALTER TABLE decision
    DROP COLUMN IF EXISTS timestamp_ms,
    DROP COLUMN IF EXISTS parent_span_id,
    DROP COLUMN IF EXISTS event_name,
    DROP COLUMN IF EXISTS status;

COMMIT;
