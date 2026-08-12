BEGIN;

ALTER TABLE decision
    ADD COLUMN IF NOT EXISTS timestamp_ms BIGINT,
    ADD COLUMN IF NOT EXISTS parent_span_id CHAR(16),
    ADD COLUMN IF NOT EXISTS event_name VARCHAR(40),
    ADD COLUMN IF NOT EXISTS status VARCHAR(10);

UPDATE decision
SET timestamp_ms = (EXTRACT(EPOCH FROM created) * 1000)::BIGINT
WHERE timestamp_ms IS NULL;

UPDATE decision d
SET event_name = rt.name
FROM request_type rt
WHERE d.event_name IS NULL AND d.request_type = rt.id;

UPDATE decision
SET event_name = 'adl.access_evaluation'
WHERE event_name IS NULL;

UPDATE decision
SET status = 'Unset'
WHERE status IS NULL;

ALTER TABLE decision
    ALTER COLUMN timestamp_ms SET NOT NULL,
    ALTER COLUMN event_name SET NOT NULL,
    ALTER COLUMN status SET NOT NULL;

COMMIT;
