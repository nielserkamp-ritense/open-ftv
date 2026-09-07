BEGIN;

DROP INDEX IF EXISTS decision_ix1;

DROP INDEX IF EXISTS decision_ux_trace_span;

CREATE INDEX IF NOT EXISTS decision_ix2 ON decision (trace_id, span_id);

ALTER TABLE decision
    ADD COLUMN IF NOT EXISTS request JSONB,
    ADD COLUMN IF NOT EXISTS response JSONB,
    ADD COLUMN IF NOT EXISTS created TIMESTAMP;

UPDATE decision SET
    created = (to_timestamp(timestamp / 1000.0) AT TIME ZONE 'UTC'),
    request = body -> 'adl.core.request',
    response = body -> 'adl.core.response';

ALTER TABLE decision
    ALTER COLUMN created SET NOT NULL;

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
