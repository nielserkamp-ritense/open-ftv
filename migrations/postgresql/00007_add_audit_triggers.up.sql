BEGIN;

CREATE OR REPLACE FUNCTION policy_logger() RETURNS TRIGGER AS $BODY$
DECLARE
    updated TIMESTAMP = now() AT TIME ZONE 'UTC';
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

CREATE TRIGGER policy_trigger
    AFTER INSERT OR UPDATE OR DELETE ON policy
    FOR EACH ROW EXECUTE FUNCTION policy_logger();

CREATE OR REPLACE FUNCTION attribute_logger() RETURNS TRIGGER AS $BODY$
DECLARE
    updated TIMESTAMP = now() AT TIME ZONE 'UTC';
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

CREATE TRIGGER attribute_trigger
    AFTER INSERT OR UPDATE OR DELETE ON attribute
    FOR EACH ROW EXECUTE FUNCTION attribute_logger();

CREATE OR REPLACE FUNCTION entity_logger() RETURNS TRIGGER AS $BODY$
DECLARE
    updated TIMESTAMP = now() AT TIME ZONE 'UTC';
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

CREATE TRIGGER entity_trigger
    AFTER INSERT OR UPDATE OR DELETE ON entity
    FOR EACH ROW EXECUTE FUNCTION entity_logger();

CREATE OR REPLACE FUNCTION relation_logger() RETURNS TRIGGER AS $BODY$
DECLARE
    updated TIMESTAMP = now() AT TIME ZONE 'UTC';
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

CREATE TRIGGER relation_trigger
    AFTER INSERT OR UPDATE OR DELETE ON relation
    FOR EACH ROW EXECUTE FUNCTION relation_logger();

COMMIT;
