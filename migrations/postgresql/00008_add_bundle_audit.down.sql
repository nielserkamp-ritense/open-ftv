BEGIN;

ALTER TABLE policy_deployment ADD COLUMN version INTEGER REFERENCES deployment;
ALTER TABLE policy_deployment DROP COLUMN bundle_id;

ALTER TABLE attribute_deployment ADD COLUMN version INTEGER REFERENCES deployment;
ALTER TABLE attribute_deployment DROP COLUMN bundle_id;

ALTER TABLE entity_deployment ADD COLUMN version INTEGER REFERENCES deployment;
ALTER TABLE entity_deployment DROP COLUMN bundle_id;

ALTER TABLE relation_deployment ADD COLUMN version INTEGER REFERENCES deployment;
ALTER TABLE relation_deployment DROP COLUMN bundle_id;

DROP TABLE deployment_bundle;

COMMIT;
