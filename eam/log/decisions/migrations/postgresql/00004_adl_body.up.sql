BEGIN;

ALTER TABLE decision
    ADD COLUMN IF NOT EXISTS body JSONB;

UPDATE decision
SET body = jsonb_build_object('request', request, 'response', response)
WHERE body IS NULL;

ALTER TABLE decision
    DROP COLUMN IF EXISTS request,
    DROP COLUMN IF EXISTS response;

COMMIT;
