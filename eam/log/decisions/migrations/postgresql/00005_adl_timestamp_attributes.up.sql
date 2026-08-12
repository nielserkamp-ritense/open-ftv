BEGIN;

ALTER TABLE decision
    RENAME COLUMN timestamp_ms TO timestamp;

ALTER TABLE decision
    ADD COLUMN IF NOT EXISTS attributes JSONB;

UPDATE decision
SET attributes = jsonb_build_object(
    'adl.core.request', jsonb_build_object('ref', 'body.request'),
    'adl.core.response', jsonb_build_object('ref', 'body.response')
)
WHERE body IS NOT NULL AND attributes IS NULL;

COMMIT;
