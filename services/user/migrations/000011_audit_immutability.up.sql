-- Append-only enforcement for compliance audit/evidence tables (DB-level, not app-only).
CREATE OR REPLACE FUNCTION "user".audit_trail_immutable()
RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'user.% is append-only', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    tbl TEXT;
BEGIN
    FOREACH tbl IN ARRAY ARRAY['audit_log', 'pii_access_log', 'gdpr_requests', 'kyc_submissions', 'screening_checks']
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS %I_no_update ON "user".%I', tbl, tbl);
        EXECUTE format(
            'CREATE TRIGGER %I_no_update BEFORE UPDATE ON "user".%I FOR EACH ROW EXECUTE FUNCTION "user".audit_trail_immutable()',
            tbl, tbl
        );
        EXECUTE format('DROP TRIGGER IF EXISTS %I_no_delete ON "user".%I', tbl, tbl);
        EXECUTE format(
            'CREATE TRIGGER %I_no_delete BEFORE DELETE ON "user".%I FOR EACH ROW EXECUTE FUNCTION "user".audit_trail_immutable()',
            tbl, tbl
        );
    END LOOP;
END $$;