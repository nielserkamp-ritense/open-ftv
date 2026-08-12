BEGIN;

-- Store W3C trace/span identifiers as readable lowercase hex text, matching Logius
-- LDV/ADL JSON serialization (32- and 16-character hex strings). VARCHAR, not
ALTER TABLE decision
    ALTER COLUMN trace_id TYPE VARCHAR(32) USING NULLIF(encode(trace_id, 'hex'), ''),
    ALTER COLUMN span_id TYPE VARCHAR(16) USING NULLIF(encode(span_id, 'hex'), '');

ALTER TABLE decision
    ADD COLUMN IF NOT EXISTS timestamp BIGINT,
    ADD COLUMN IF NOT EXISTS parent_span_id VARCHAR(16),
    ADD COLUMN IF NOT EXISTS event_name VARCHAR(40),
    ADD COLUMN IF NOT EXISTS status VARCHAR(10),
    ADD COLUMN IF NOT EXISTS body JSONB,
    -- attributes: Logius ADL source references. Empty at Level 1, since
    -- request/response live in body instead.
    ADD COLUMN IF NOT EXISTS attributes JSONB,
    -- resource: identifies the producer of the log record.
    ADD COLUMN IF NOT EXISTS resource JSONB;

ALTER TABLE decision
    DROP COLUMN IF EXISTS request,
    DROP COLUMN IF EXISTS response,
    DROP COLUMN IF EXISTS created;

ALTER TABLE decision
    ALTER COLUMN timestamp SET NOT NULL,
    ALTER COLUMN event_name SET NOT NULL,
    ALTER COLUMN status SET NOT NULL,
    ALTER COLUMN trace_id SET NOT NULL,
    ALTER COLUMN span_id SET NOT NULL;

COMMIT;
