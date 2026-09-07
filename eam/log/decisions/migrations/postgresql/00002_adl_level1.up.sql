BEGIN;

ALTER TABLE decision
    ALTER COLUMN trace_id TYPE VARCHAR(32) USING NULLIF(encode(trace_id, 'hex'), ''),
    ALTER COLUMN span_id TYPE VARCHAR(16) USING NULLIF(encode(span_id, 'hex'), '');

ALTER TABLE decision
    ADD COLUMN IF NOT EXISTS timestamp BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS parent_span_id VARCHAR(16),
    ADD COLUMN IF NOT EXISTS event_name VARCHAR(40),
    ADD COLUMN IF NOT EXISTS status VARCHAR(10) NOT NULL DEFAULT 'Unset',
    ADD COLUMN IF NOT EXISTS body JSONB,
    -- attributes: Logius ADL source references. Empty at Level 1, since
    -- request/response live in body instead.
    ADD COLUMN IF NOT EXISTS attributes JSONB,
    -- resource: identifies the producer of the log record.
    ADD COLUMN IF NOT EXISTS resource JSONB;

DELETE FROM decision WHERE request_type IS NULL OR request_type NOT IN (1, 2, 3, 4, 5);

UPDATE decision SET
    timestamp = (extract(epoch FROM created) * 1000)::bigint,
    event_name = CASE request_type
        WHEN 1 THEN 'adl.access_evaluation'
        WHEN 2 THEN 'adl.access_evaluations'
        WHEN 3 THEN 'adl.search_subject'
        WHEN 4 THEN 'adl.search_action'
        WHEN 5 THEN 'adl.search_resource'
    END,
    status = 'Unset';

ALTER TABLE decision
    DROP COLUMN IF EXISTS request,
    DROP COLUMN IF EXISTS response,
    DROP COLUMN IF EXISTS created;

ALTER TABLE decision
    ALTER COLUMN event_name SET NOT NULL;

-- (trace_id, span_id) is the idempotency key.
DROP INDEX IF EXISTS decision_ix2;

CREATE UNIQUE INDEX IF NOT EXISTS decision_ux_trace_span ON decision (trace_id, span_id);
CREATE INDEX IF NOT EXISTS decision_ix1 ON decision (timestamp);

COMMIT;
