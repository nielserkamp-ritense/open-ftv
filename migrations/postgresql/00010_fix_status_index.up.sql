BEGIN;

DROP INDEX IF EXISTS policy_status_ix;
DROP INDEX IF EXISTS attribute_status_ix;
DROP INDEX IF EXISTS entity_status_ix;
DROP INDEX IF EXISTS relation_status_ix;

CREATE INDEX policy_status_ix ON policy (status);
CREATE INDEX attribute_status_ix ON attribute (status);
CREATE INDEX entity_status_ix ON entity (status);
CREATE INDEX relation_status_ix ON relation (status);

COMMIT;
