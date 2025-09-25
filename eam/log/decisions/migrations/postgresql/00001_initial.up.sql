BEGIN;

CREATE TABLE IF NOT EXISTS request_type
(
    id   SMALLINT NOT NULL PRIMARY KEY,
    name VARCHAR(20)
);

INSERT INTO request_type (id,name) VALUES(1,'evaluation');
INSERT INTO request_type (id,name) VALUES(2,'evaluations');
INSERT INTO request_type (id,name) VALUES(3,'search_subject');
INSERT INTO request_type (id,name) VALUES(4,'search_action');
INSERT INTO request_type (id,name) VALUES(5,'search_resource');

CREATE TABLE IF NOT EXISTS decision
(
    id           BIGSERIAL PRIMARY KEY,
    created      TIMESTAMP NOT NULL,
    request_type SMALLINT REFERENCES request_type,
    policies     BIGINT,
    request      JSONB,
    response     JSONB,
    information  JSONB,
    engine       JSONB,
    trace_id     BYTEA,
    span_id      BYTEA
);

CREATE INDEX IF NOT EXISTS decision_ix1 ON decision (created);
CREATE INDEX IF NOT EXISTS decision_ix2 ON decision (trace_id, span_id);
CREATE INDEX IF NOT EXISTS decision_ix3 ON decision (request_type);
CREATE INDEX IF NOT EXISTS decision_ix4 ON decision (policies);

COMMIT;
