BEGIN;

ALTER TABLE "public"."policy_audit" DROP CONSTRAINT "policy_audit_id_fkey";

COMMIT;
