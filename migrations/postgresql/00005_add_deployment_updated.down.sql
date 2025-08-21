BEGIN;

ALTER TABLE deployment DROP COLUMN updated;
ALTER TABLE deployment DROP COLUMN updated_by;

COMMIT;
