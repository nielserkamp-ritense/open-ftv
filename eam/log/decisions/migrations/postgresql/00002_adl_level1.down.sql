BEGIN;

ALTER TABLE decision
    ADD COLUMN IF NOT EXISTS request JSONB,
    ADD COLUMN IF NOT EXISTS response JSONB,
    ADD COLUMN IF NOT EXISTS created TIMESTAMP;

ALTER TABLE decision
    DROP COLUMN IF EXISTS body,
    DROP COLUMN IF EXISTS attributes,
    DROP COLUMN IF EXISTS resource,
    DROP COLUMN IF EXISTS timestamp,
    DROP COLUMN IF EXISTS parent_span_id,
    DROP COLUMN IF EXISTS event_name,
    DROP COLUMN IF EXISTS status;

ALTER TABLE decision
    ALTER COLUMN trace_id TYPE BYTEA USING decode(trace_id, 'hex'),
    ALTER COLUMN span_id TYPE BYTEA USING decode(span_id, 'hex'),
    ALTER COLUMN trace_id DROP NOT NULL,
    ALTER COLUMN span_id DROP NOT NULL;

COMMIT;
