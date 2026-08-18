BEGIN;

ALTER TABLE language DROP CONSTRAINT IF EXISTS language_updated_by_fk;
ALTER TABLE tag DROP CONSTRAINT IF EXISTS tag_updated_by_fk;
ALTER TABLE deployment DROP CONSTRAINT IF EXISTS deployment_updated_by_fk;
ALTER TABLE policy DROP CONSTRAINT IF EXISTS policy_updated_by_fk;
ALTER TABLE attribute DROP CONSTRAINT IF EXISTS attribute_updated_by_fk;
ALTER TABLE entity DROP CONSTRAINT IF EXISTS entity_updated_by_fk;
ALTER TABLE relation DROP CONSTRAINT IF EXISTS relation_updated_by_fk;
ALTER TABLE settings DROP CONSTRAINT IF EXISTS settings_updated_by_fk;
ALTER TABLE language DROP CONSTRAINT IF EXISTS language_created_by_fk;
ALTER TABLE tag DROP CONSTRAINT IF EXISTS tag_created_by_fk;
ALTER TABLE deployment DROP CONSTRAINT IF EXISTS deployment_created_by_fk;
ALTER TABLE policy DROP CONSTRAINT IF EXISTS policy_created_by_fk;
ALTER TABLE attribute DROP CONSTRAINT IF EXISTS attribute_created_by_fk;
ALTER TABLE entity DROP CONSTRAINT IF EXISTS entity_created_by_fk;
ALTER TABLE relation DROP CONSTRAINT IF EXISTS relation_created_by_fk;
ALTER TABLE settings DROP CONSTRAINT IF EXISTS settings_created_by_fk;
ALTER TABLE deployment_bundle DROP CONSTRAINT IF EXISTS deployment_bundle_created_by_fk;

DROP TABLE IF EXISTS principal;

COMMIT;
