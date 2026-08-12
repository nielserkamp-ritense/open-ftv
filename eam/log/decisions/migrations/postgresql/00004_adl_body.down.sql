BEGIN;

ALTER TABLE decision
    ADD COLUMN IF NOT EXISTS request JSONB,
    ADD COLUMN IF NOT EXISTS response JSONB;

UPDATE decision
SET request = body -> 'request',
    response = body -> 'response'
WHERE body IS NOT NULL;

ALTER TABLE decision
    DROP COLUMN IF EXISTS body;

COMMIT;
