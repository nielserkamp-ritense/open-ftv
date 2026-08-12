BEGIN;

ALTER TABLE decision
    ALTER COLUMN trace_id TYPE BYTEA USING decode(trace_id, 'hex'),
    ALTER COLUMN span_id TYPE BYTEA USING decode(span_id, 'hex');

COMMIT;
