BEGIN;

ALTER TABLE language ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';
ALTER TABLE language ALTER COLUMN created SET DEFAULT now();
ALTER TABLE language ALTER COLUMN updated TYPE TIMESTAMPTZ USING updated AT TIME ZONE 'UTC';

ALTER TABLE tag ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';
ALTER TABLE tag ALTER COLUMN created SET DEFAULT now();
ALTER TABLE tag ALTER COLUMN updated TYPE TIMESTAMPTZ USING updated AT TIME ZONE 'UTC';

ALTER TABLE deployment ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';
ALTER TABLE deployment ALTER COLUMN created SET DEFAULT now();
ALTER TABLE deployment ALTER COLUMN updated TYPE TIMESTAMPTZ USING updated AT TIME ZONE 'UTC';

ALTER TABLE policy ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';
ALTER TABLE policy ALTER COLUMN created SET DEFAULT now();
ALTER TABLE policy ALTER COLUMN updated TYPE TIMESTAMPTZ USING updated AT TIME ZONE 'UTC';

ALTER TABLE policy_audit ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';

ALTER TABLE attribute ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';
ALTER TABLE attribute ALTER COLUMN created SET DEFAULT now();
ALTER TABLE attribute ALTER COLUMN updated TYPE TIMESTAMPTZ USING updated AT TIME ZONE 'UTC';

ALTER TABLE attribute_audit ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';

ALTER TABLE entity ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';
ALTER TABLE entity ALTER COLUMN created SET DEFAULT now();
ALTER TABLE entity ALTER COLUMN updated TYPE TIMESTAMPTZ USING updated AT TIME ZONE 'UTC';

ALTER TABLE entity_audit ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';

ALTER TABLE relation ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';
ALTER TABLE relation ALTER COLUMN created SET DEFAULT now();
ALTER TABLE relation ALTER COLUMN updated TYPE TIMESTAMPTZ USING updated AT TIME ZONE 'UTC';

ALTER TABLE relation_audit ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';

ALTER TABLE deployment_bundle ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';
ALTER TABLE deployment_bundle ALTER COLUMN created SET DEFAULT now();

ALTER TABLE settings ALTER COLUMN created TYPE TIMESTAMPTZ USING created AT TIME ZONE 'UTC';
ALTER TABLE settings ALTER COLUMN created SET DEFAULT now();
ALTER TABLE settings ALTER COLUMN updated TYPE TIMESTAMPTZ USING updated AT TIME ZONE 'UTC';

CREATE OR REPLACE FUNCTION policy_logger() RETURNS TRIGGER AS $BODY$
DECLARE
    updated TIMESTAMPTZ = now();
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO policy_audit(id,created,operation,user_id)
        VALUES (NEW.id,updated,'C',current_setting('openftv.user'));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO policy_audit (id,created,operation,user_id)
        VALUES (NEW.id,updated,'U',current_setting('openftv.user'));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO policy_audit(id,created,operation,user_id)
        VALUES (OLD.id,updated,'D',current_setting('openftv.user'));
        RETURN NEW;
    END IF;
END;
$BODY$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION attribute_logger() RETURNS TRIGGER AS $BODY$
DECLARE
    updated TIMESTAMPTZ = now();
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO attribute_audit(key,created,operation,user_id)
        VALUES (NEW.key,updated,'C',current_setting('openftv.user'));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO attribute_audit (key,created,operation,user_id)
        VALUES (NEW.key,updated,'U',current_setting('openftv.user'));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO attribute_audit(key,created,operation,user_id)
        VALUES (OLD.key,updated,'D',current_setting('openftv.user'));
        RETURN NEW;
    END IF;
END;
$BODY$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION entity_logger() RETURNS TRIGGER AS $BODY$
DECLARE
    updated TIMESTAMPTZ = now();
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO entity_audit(type,id,created,operation,user_id)
        VALUES (NEW.type, NEW.id,updated,'C',current_setting('openftv.user'));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO entity_audit (type,id,created,operation,user_id)
        VALUES (NEW.type, NEW.id,updated,'U',current_setting('openftv.user'));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO entity_audit(type,id,created,operation,user_id)
        VALUES (OLD.type, OLD.id,updated,'D',current_setting('openftv.user'));
        RETURN NEW;
    END IF;
END;
$BODY$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION relation_logger() RETURNS TRIGGER AS $BODY$
DECLARE
    updated TIMESTAMPTZ = now();
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO relation_audit(id,created,operation,user_id)
        VALUES (NEW.id,updated,'C',current_setting('openftv.user'));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO relation_audit (id,created,operation,user_id)
        VALUES (NEW.id,updated,'U',current_setting('openftv.user'));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO relation_audit(id,created,operation,user_id)
        VALUES (OLD.id,updated,'D',current_setting('openftv.user'));
        RETURN NEW;
    END IF;
END;
$BODY$ LANGUAGE plpgsql;

COMMIT;
