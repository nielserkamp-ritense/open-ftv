BEGIN;

DROP TRIGGER IF EXISTS policy_trigger ON policy;
DROP FUNCTION IF EXISTS policy_logger;

DROP TRIGGER IF EXISTS attribute_trigger ON attribute;
DROP FUNCTION IF EXISTS attribute_logger;

DROP TRIGGER IF EXISTS entity_trigger ON entity;
DROP FUNCTION IF EXISTS entity_logger;

DROP TRIGGER IF EXISTS relation_trigger ON relation;
DROP FUNCTION IF EXISTS relation_logger;

COMMIT;
