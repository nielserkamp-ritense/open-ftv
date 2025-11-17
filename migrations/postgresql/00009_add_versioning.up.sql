BEGIN;

ALTER TABLE policy ADD COLUMN status VARCHAR(40);

CREATE UNIQUE INDEX IF NOT EXISTS policy_status_ix ON policy (status);

ALTER TABLE attribute ADD COLUMN status VARCHAR(40);

CREATE UNIQUE INDEX IF NOT EXISTS attribute_status_ix ON attribute (status);

ALTER TABLE entity ADD COLUMN status VARCHAR(40);

CREATE UNIQUE INDEX IF NOT EXISTS entity_status_ix ON entity (status);

ALTER TABLE relation ADD COLUMN status VARCHAR(40);

CREATE UNIQUE INDEX IF NOT EXISTS relation_status_ix ON relation (status);

CREATE TABLE IF NOT EXISTS policy_version
(
    version     INTEGER NOT NULL,
    id          UUID    NOT NULL,
    language    VARCHAR(40) NOT NULL REFERENCES language,
    title       VARCHAR(80) NOT NULL,
    rvva_id     VARCHAR(80),
    uri         VARCHAR(400),
    tags        VARCHAR(40) ARRAY,
    description TEXT,
    content     TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS policy_version_pk ON policy_version (version, id);

CREATE TABLE IF NOT EXISTS attribute_version
(
    version     INTEGER      NOT NULL,
    key         VARCHAR(200) NOT NULL,
    type        VARCHAR(40)  NOT NULL,
    title       VARCHAR(80),
    value       BYTEA,
    original    BYTEA,
    tags        VARCHAR(40) ARRAY,
    description TEXT,
    content     TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS attribute_version_pk ON attribute_version (version, key);

CREATE TABLE IF NOT EXISTS entity_version
(
    version     INTEGER      NOT NULL,
    type        VARCHAR(80)  NOT NULL,
    id          VARCHAR(200) NOT NULL,
    title       VARCHAR(80),
    value       BYTEA,
    original    BYTEA,
    tags        VARCHAR(40)  ARRAY,
    parents     VARCHAR(200) ARRAY,
    description TEXT,
    content     TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS entity_version_pk ON entity_version (version, type, id);

CREATE TABLE IF NOT EXISTS relation_version
(
    version      INTEGER      NOT NULL,
    id           UUID         NOT NULL,
    subject_type VARCHAR(80)  NOT NULL,
    subject_id   VARCHAR(200) NOT NULL,
    predicate    VARCHAR(80)  NOT NULL,
    object_type  VARCHAR(80)  NOT NULL,
    object_id    VARCHAR(200) NOT NULL,
    title        VARCHAR(80),
    tags         VARCHAR(40) ARRAY,
    description  TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS relation_version_pk ON relation_version (version, id);

COMMIT;
