BEGIN;

-- The principal table is the manager's local record of every subject it has seen act on the
-- management plane, so a stored attribution string can be rendered as a name and linked to.
--
-- NEVER ADD A roles COLUMN. Roles stay in the token and the PDP remains authoritative; a
-- persisted role set would be a second authorization model that silently diverges from the
-- real one. See docs/adr/0002-roles-as-open-data-flat-claim.md.
--
-- id is the attribution string itself, not a generated key: a JWT sub for users, a *SYSTEM*-
-- style sentinel for the manager's own actions, and a display name for rows carried over from
-- before this table existed. That is what lets created_by/updated_by gain a foreign key
-- without rewriting a single attribution value. There is deliberately no separate subject
-- column: it would duplicate the primary key exactly. issuer is kept as provenance so a
-- future second IdP can be migrated to generated ids if it ever becomes necessary.
--
-- active currently has no writer: the manager has no deprovisioning signal from the IdP and
-- only ever learns of a user by seeing one succeed. It exists because ON DELETE RESTRICT
-- below leaves deactivation as the only way to retire a principal.
-- See docs/adr/0004-principal-attribution-and-retention.md.
CREATE TABLE IF NOT EXISTS principal
(
    id           VARCHAR(80) PRIMARY KEY,
    kind         VARCHAR(40)  NOT NULL,
    issuer       VARCHAR(400) NOT NULL DEFAULT '',
    display_name VARCHAR(200),
    email        VARCHAR(320),
    first_seen   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    last_seen    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    active       BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS principal_ix1 ON principal (kind);

-- Pass 0: the sentinels the manager attributes its own actions to. These must exist before
-- any other pass, so the passes below skip them rather than classifying them as users.
INSERT INTO principal (id, kind, display_name)
VALUES
    ('*SYSTEM*', 'system', '*SYSTEM*'),
    ('*SEED*', 'system', '*SEED*'),
    ('*BUNDLE*', 'system', '*BUNDLE*'),
    ('*BOOTSTRAP*', 'system', '*BOOTSTRAP*'),
    ('*LOADER*', 'system', '*LOADER*'),
    ('*STORAGE*', 'system', '*STORAGE*'),
    ('*UNKNOWN*', 'system', '*UNKNOWN*')
ON CONFLICT (id) DO NOTHING;

-- Pass 1: real principal ids, taken from the audit tables, which have always stored the id
-- (the trigger writes current_setting('openftv.user')). display_name is left NULL so the UI
-- falls back to the raw id until the owner next logs in and the upsert fills in their name.
--
-- This pass MUST precede pass 2. eam/pip/attributes_pg.go already stored user.ID, so
-- attribute.created_by holds ids while policy/entity/tag/settings/deployment hold names, and
-- SQL cannot tell the two apart by inspection. Every attribute write also fired the audit
-- trigger with the identical value, so claiming the ids here leaves pass 2 with only names.
INSERT INTO principal (id, kind)
SELECT DISTINCT policy_audit.user_id, 'user' FROM policy_audit WHERE policy_audit.user_id <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind)
SELECT DISTINCT attribute_audit.user_id, 'user' FROM attribute_audit WHERE attribute_audit.user_id <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind)
SELECT DISTINCT entity_audit.user_id, 'user' FROM entity_audit WHERE entity_audit.user_id <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind)
SELECT DISTINCT relation_audit.user_id, 'user' FROM relation_audit WHERE relation_audit.user_id <> ''
ON CONFLICT (id) DO NOTHING;

-- Pass 2: values carried over from before this table existed. These are display names, not
-- ids, so they are marked 'legacy': the UI shows them but must not offer a link, and the same
-- human gets a second, sub-keyed row the next time they log in. The attribution columns
-- themselves are left untouched -- this only makes the rows they reference exist.
INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT language.created_by, 'legacy', language.created_by FROM language WHERE language.created_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT tag.created_by, 'legacy', tag.created_by FROM tag WHERE tag.created_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT deployment.created_by, 'legacy', deployment.created_by FROM deployment WHERE deployment.created_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT policy.created_by, 'legacy', policy.created_by FROM policy WHERE policy.created_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT attribute.created_by, 'legacy', attribute.created_by FROM attribute WHERE attribute.created_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT entity.created_by, 'legacy', entity.created_by FROM entity WHERE entity.created_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT relation.created_by, 'legacy', relation.created_by FROM relation WHERE relation.created_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT settings.created_by, 'legacy', settings.created_by FROM settings WHERE settings.created_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT deployment_bundle.created_by, 'legacy', deployment_bundle.created_by FROM deployment_bundle WHERE deployment_bundle.created_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT language.updated_by, 'legacy', language.updated_by FROM language WHERE language.updated_by IS NOT NULL AND language.updated_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT tag.updated_by, 'legacy', tag.updated_by FROM tag WHERE tag.updated_by IS NOT NULL AND tag.updated_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT deployment.updated_by, 'legacy', deployment.updated_by FROM deployment WHERE deployment.updated_by IS NOT NULL AND deployment.updated_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT policy.updated_by, 'legacy', policy.updated_by FROM policy WHERE policy.updated_by IS NOT NULL AND policy.updated_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT attribute.updated_by, 'legacy', attribute.updated_by FROM attribute WHERE attribute.updated_by IS NOT NULL AND attribute.updated_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT entity.updated_by, 'legacy', entity.updated_by FROM entity WHERE entity.updated_by IS NOT NULL AND entity.updated_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT relation.updated_by, 'legacy', relation.updated_by FROM relation WHERE relation.updated_by IS NOT NULL AND relation.updated_by <> ''
ON CONFLICT (id) DO NOTHING;

INSERT INTO principal (id, kind, display_name)
SELECT DISTINCT settings.updated_by, 'legacy', settings.updated_by FROM settings WHERE settings.updated_by IS NOT NULL AND settings.updated_by <> ''
ON CONFLICT (id) DO NOTHING;

-- Empty attribution values predate the *UNKNOWN* sentinel and reference nothing. Normalize
-- them so the constraints below can be validated; this is the only place any attribution
-- column is rewritten, and only where it held no information to begin with.
UPDATE language SET created_by = '*UNKNOWN*' WHERE created_by = '';
UPDATE tag SET created_by = '*UNKNOWN*' WHERE created_by = '';
UPDATE deployment SET created_by = '*UNKNOWN*' WHERE created_by = '';
UPDATE policy SET created_by = '*UNKNOWN*' WHERE created_by = '';
UPDATE attribute SET created_by = '*UNKNOWN*' WHERE created_by = '';
UPDATE entity SET created_by = '*UNKNOWN*' WHERE created_by = '';
UPDATE relation SET created_by = '*UNKNOWN*' WHERE created_by = '';
UPDATE settings SET created_by = '*UNKNOWN*' WHERE created_by = '';
UPDATE deployment_bundle SET created_by = '*UNKNOWN*' WHERE created_by = '';
UPDATE language SET updated_by = NULL WHERE updated_by = '';
UPDATE tag SET updated_by = NULL WHERE updated_by = '';
UPDATE deployment SET updated_by = NULL WHERE updated_by = '';
UPDATE policy SET updated_by = NULL WHERE updated_by = '';
UPDATE attribute SET updated_by = NULL WHERE updated_by = '';
UPDATE entity SET updated_by = NULL WHERE updated_by = '';
UPDATE relation SET updated_by = NULL WHERE updated_by = '';
UPDATE settings SET updated_by = NULL WHERE updated_by = '';

-- Attribution is an active relationship, so the database enforces that the subject exists.
-- ON DELETE RESTRICT is what implements 'principals are never deleted': retire one by setting
-- active = false, because a retired owner must still be visible as the owner. Never CASCADE.
--
-- The *_audit tables deliberately get NO foreign key: an audit write must never fail because
-- a principal row is missing, and audit history must outlive deprovisioning.
ALTER TABLE language ADD CONSTRAINT language_created_by_fk
    FOREIGN KEY (created_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE tag ADD CONSTRAINT tag_created_by_fk
    FOREIGN KEY (created_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE deployment ADD CONSTRAINT deployment_created_by_fk
    FOREIGN KEY (created_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE policy ADD CONSTRAINT policy_created_by_fk
    FOREIGN KEY (created_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE attribute ADD CONSTRAINT attribute_created_by_fk
    FOREIGN KEY (created_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE entity ADD CONSTRAINT entity_created_by_fk
    FOREIGN KEY (created_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE relation ADD CONSTRAINT relation_created_by_fk
    FOREIGN KEY (created_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE settings ADD CONSTRAINT settings_created_by_fk
    FOREIGN KEY (created_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE deployment_bundle ADD CONSTRAINT deployment_bundle_created_by_fk
    FOREIGN KEY (created_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE language ADD CONSTRAINT language_updated_by_fk
    FOREIGN KEY (updated_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE tag ADD CONSTRAINT tag_updated_by_fk
    FOREIGN KEY (updated_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE deployment ADD CONSTRAINT deployment_updated_by_fk
    FOREIGN KEY (updated_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE policy ADD CONSTRAINT policy_updated_by_fk
    FOREIGN KEY (updated_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE attribute ADD CONSTRAINT attribute_updated_by_fk
    FOREIGN KEY (updated_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE entity ADD CONSTRAINT entity_updated_by_fk
    FOREIGN KEY (updated_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE relation ADD CONSTRAINT relation_updated_by_fk
    FOREIGN KEY (updated_by) REFERENCES principal (id) ON DELETE RESTRICT;
ALTER TABLE settings ADD CONSTRAINT settings_updated_by_fk
    FOREIGN KEY (updated_by) REFERENCES principal (id) ON DELETE RESTRICT;

COMMIT;
