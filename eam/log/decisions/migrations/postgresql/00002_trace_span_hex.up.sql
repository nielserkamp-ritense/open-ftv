BEGIN;

-- Store W3C trace/span identifiers as readable lowercase hex text, matching Logius
-- LDV/ADL JSON serialization (32- and 16-character hex strings).
ALTER TABLE decision
    ALTER COLUMN trace_id TYPE CHAR(32) USING NULLIF(encode(trace_id, 'hex'), ''),
    ALTER COLUMN span_id TYPE CHAR(16) USING NULLIF(encode(span_id, 'hex'), '');

COMMIT;
