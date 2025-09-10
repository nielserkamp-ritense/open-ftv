BEGIN;

CREATE TABLE IF NOT EXISTS deployment_bundle
(
    id         UUID PRIMARY KEY,
    version    INTEGER NOT NULL REFERENCES deployment,
    bundle_id  VARCHAR(40) NOT NULL,
    config     JSONB NOT NULL,
    created    TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    created_by VARCHAR(80) NOT NULL DEFAULT '*SYSTEM*'
);

CREATE INDEX IF NOT EXISTS deployment_bundle_ix1 ON deployment_bundle (bundle_id);
CREATE INDEX IF NOT EXISTS deployment_bundle_ix2 ON deployment_bundle (version);

ALTER TABLE policy_deployment ADD COLUMN bundle_id UUID REFERENCES deployment_bundle;
ALTER TABLE policy_deployment DROP COLUMN version;

ALTER TABLE attribute_deployment ADD COLUMN bundle_id UUID REFERENCES deployment_bundle;
ALTER TABLE attribute_deployment DROP COLUMN version;

ALTER TABLE entity_deployment ADD COLUMN bundle_id UUID REFERENCES deployment_bundle;
ALTER TABLE entity_deployment DROP COLUMN version;

ALTER TABLE relation_deployment ADD COLUMN bundle_id UUID REFERENCES deployment_bundle;
ALTER TABLE relation_deployment DROP COLUMN version;

COMMIT;
