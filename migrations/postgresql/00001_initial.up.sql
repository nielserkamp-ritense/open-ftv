BEGIN;

CREATE TABLE IF NOT EXISTS language
(
    language    VARCHAR(40) PRIMARY KEY,
    title       VARCHAR(80),
    description TEXT,
    created     TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    created_by  VARCHAR(80) NOT NULL DEFAULT '*SYSTEM*',
    updated     TIMESTAMP,
    updated_by  VARCHAR(80)
);

CREATE TABLE IF NOT EXISTS tag
(
    tag         VARCHAR(40) PRIMARY KEY,
    title       VARCHAR(80),
    description TEXT,
    created     TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    created_by  VARCHAR(80) NOT NULL DEFAULT '*SYSTEM*',
    updated     TIMESTAMP,
    updated_by  VARCHAR(80)
);

CREATE TABLE IF NOT EXISTS deployment
(
    version     INTEGER PRIMARY KEY,
    status      INTEGER NOT NULL DEFAULT 0,
    title       VARCHAR(80),
    description TEXT,
    created     TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    created_by  VARCHAR(80) NOT NULL DEFAULT '*SYSTEM*'
);

CREATE TABLE IF NOT EXISTS policy
(
    id          UUID PRIMARY KEY,
    language    VARCHAR(40) NOT NULL REFERENCES language,
    title       VARCHAR(80) NOT NULL,
    rvva_id     VARCHAR(80),
    uri         VARCHAR(400),
    tags        VARCHAR(40) ARRAY,
    description TEXT,
    content     TEXT,
    created     TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    created_by  VARCHAR(80) NOT NULL DEFAULT '*SYSTEM*',
    updated     TIMESTAMP,
    updated_by  VARCHAR(80)
);

CREATE UNIQUE INDEX IF NOT EXISTS policy_ix1 ON policy (language, title);

CREATE TABLE IF NOT EXISTS policy_audit
(
    id        UUID        NOT NULL REFERENCES policy,
    created   TIMESTAMP NOT NULL,
    operation CHAR(1)     NOT NULL,
    user_id   VARCHAR(80) NOT NULL
);

CREATE INDEX IF NOT EXISTS policy_audit_ix1 ON policy_audit (id);
CREATE INDEX IF NOT EXISTS policy_audit_ix2 ON policy_audit (created);

CREATE TABLE IF NOT EXISTS policy_deployment
(
    id      UUID    NOT NULL REFERENCES policy,
    version INTEGER NOT NULL REFERENCES deployment
);

CREATE INDEX IF NOT EXISTS policy_deployment_ix1 ON policy_deployment (id);
CREATE INDEX IF NOT EXISTS policy_deployment_ix2 ON policy_deployment (version);

CREATE TABLE IF NOT EXISTS attribute
(
    key         VARCHAR(200) PRIMARY KEY,
    type        VARCHAR(40)  NOT NULL,
    title       VARCHAR(80),
    value       BYTEA,
    original    BYTEA,
    tags        VARCHAR(40) ARRAY,
    description TEXT,
    content     TEXT,
    created     TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    created_by  VARCHAR(80) NOT NULL DEFAULT '*SYSTEM*',
    updated     TIMESTAMP,
    updated_by  VARCHAR(80)
);

CREATE TABLE IF NOT EXISTS attribute_audit
(
    key       VARCHAR(200) NOT NULL REFERENCES attribute,
    created   TIMESTAMP  NOT NULL,
    operation CHAR(1)      NOT NULL,
    user_id   VARCHAR(80)  NOT NULL
);

CREATE INDEX IF NOT EXISTS attribute_audit_ix1 ON attribute_audit (key);
CREATE INDEX IF NOT EXISTS attribute_audit_ix2 ON attribute_audit (created);
CREATE INDEX IF NOT EXISTS attribute_audit_ix3 ON attribute_audit (user_id);

CREATE TABLE IF NOT EXISTS attribute_deployment
(
    key     VARCHAR(200) NOT NULL REFERENCES attribute,
    version INTEGER      NOT NULL REFERENCES deployment
);

CREATE INDEX IF NOT EXISTS attribute_deployment_ix1 ON attribute_deployment (key);
CREATE INDEX IF NOT EXISTS attribute_deployment_ix2 ON attribute_deployment (version);

CREATE TABLE IF NOT EXISTS entity
(
    type        VARCHAR(80)  NOT NULL,
    id          VARCHAR(200) NOT NULL,
    title       VARCHAR(80),
    value       BYTEA,
    original    BYTEA,
    tags        VARCHAR(40)  ARRAY,
    parents     VARCHAR(200) ARRAY,
    description TEXT,
    content     TEXT,
    created     TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    created_by  VARCHAR(80) NOT NULL DEFAULT '*SYSTEM*',
    updated     TIMESTAMP,
    updated_by  VARCHAR(80)
);

CREATE UNIQUE INDEX IF NOT EXISTS entity_ix1 ON entity (type, id);
ALTER TABLE entity DROP CONSTRAINT IF EXISTS entity_pk CASCADE;
ALTER TABLE entity ADD CONSTRAINT entity_pk PRIMARY KEY USING INDEX entity_ix1;

CREATE TABLE IF NOT EXISTS entity_audit
(
    type      VARCHAR(80)  NOT NULL,
    id        VARCHAR(200) NOT NULL,
    created   TIMESTAMP  NOT NULL,
    operation CHAR(1)      NOT NULL,
    user_id   VARCHAR(80)  NOT NULL,
    FOREIGN KEY (type, id) REFERENCES entity (type, id)
);

CREATE INDEX IF NOT EXISTS entity_audit_ix1 ON entity_audit (type, id);
CREATE INDEX IF NOT EXISTS entity_audit_ix2 ON entity_audit (created);
CREATE INDEX IF NOT EXISTS entity_audit_ix3 ON entity_audit (user_id);

CREATE TABLE IF NOT EXISTS entity_deployment
(
    type    VARCHAR(80)  NOT NULL,
    id      VARCHAR(200) NOT NULL,
    version INTEGER      NOT NULL REFERENCES deployment
);

CREATE INDEX IF NOT EXISTS entity_deployment_ix1 ON entity_deployment (type, id, version);
CREATE INDEX IF NOT EXISTS entity_deployment_ix2 ON entity_deployment (type, id, version);

CREATE TABLE IF NOT EXISTS relation
(
    id           UUID PRIMARY KEY,
    subject_type VARCHAR(80)  NOT NULL,
    subject_id   VARCHAR(200) NOT NULL,
    predicate    VARCHAR(80)  NOT NULL,
    object_type  VARCHAR(80)  NOT NULL,
    object_id    VARCHAR(200) NOT NULL,
    title        VARCHAR(80),
    tags         VARCHAR(40) ARRAY,
    description  TEXT,
    created      TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    created_by   VARCHAR(80) NOT NULL DEFAULT '*SYSTEM*',
    updated      TIMESTAMP,
    updated_by   VARCHAR(80)
);

CREATE UNIQUE INDEX IF NOT EXISTS relation_ix1 ON relation (subject_type, subject_id, predicate, object_type, object_id);

CREATE TABLE IF NOT EXISTS relation_audit
(
    id        UUID         NOT NULL REFERENCES relation,
    created   TIMESTAMP  NOT NULL,
    operation CHAR(1)      NOT NULL,
    user_id   VARCHAR(80)  NOT NULL
);

CREATE INDEX IF NOT EXISTS relation_audit_ix1 ON relation_audit (id);
CREATE INDEX IF NOT EXISTS relation_audit_ix2 ON relation_audit (created);
CREATE INDEX IF NOT EXISTS relation_audit_ix3 ON relation_audit (user_id);

CREATE TABLE IF NOT EXISTS relation_deployment
(
    id      UUID    NOT NULL REFERENCES relation,
    version INTEGER NOT NULL REFERENCES deployment
);

CREATE INDEX IF NOT EXISTS relation_deployment_ix1 ON relation_deployment (id);
CREATE INDEX IF NOT EXISTS relation_deployment_ix2 ON relation_deployment (version);

COMMIT;
