BEGIN;

ALTER TABLE policy DROP COLUMN status;

DROP INDEX IF EXISTS policy_status_ix;

ALTER TABLE attribute DROP COLUMN status;

DROP INDEX IF EXISTS attribute_status_ix;

ALTER TABLE entity DROP COLUMN status;

DROP INDEX IF EXISTS entity_status_ix;

ALTER TABLE relation DROP COLUMN status;

DROP INDEX IF EXISTS relation_status_ix;

DROP TABLE IF EXISTS policy_version;

DROP INDEX IF EXISTS policy_version_pk;

DROP TABLE IF EXISTS attribute_version;

DROP INDEX IF EXISTS attribute_version_pk;

DROP TABLE IF EXISTS entity_version;

DROP INDEX IF EXISTS entity_version_pk;

DROP TABLE IF EXISTS relation_version;

DROP INDEX IF EXISTS relation_version_pk;

COMMIT;
