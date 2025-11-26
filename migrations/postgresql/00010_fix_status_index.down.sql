BEGIN;

DROP INDEX IF EXISTS policy_status_ix;
DROP INDEX IF EXISTS attribute_status_ix;
DROP INDEX IF EXISTS entity_status_ix;
DROP INDEX IF EXISTS relation_status_ix;

CREATE UNIQUE INDEX policy_status_ix ON policy (status);
CREATE UNIQUE INDEX attribute_status_ix ON attribute (status);
CREATE UNIQUE INDEX entity_status_ix ON entity (status);
CREATE UNIQUE INDEX relation_status_ix ON relation (status);

COMMIT;
