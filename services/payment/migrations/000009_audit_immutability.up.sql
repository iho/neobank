CREATE OR REPLACE FUNCTION payment.audit_trail_immutable()
RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'payment.% is append-only', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    tbl TEXT;
BEGIN
    FOREACH tbl IN ARRAY ARRAY['audit_log', 'fraud_decisions', 'aml_evaluations', 'screening_checks']
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS %I_no_update ON payment.%I', tbl, tbl);
        EXECUTE format(
            'CREATE TRIGGER %I_no_update BEFORE UPDATE ON payment.%I FOR EACH ROW EXECUTE FUNCTION payment.audit_trail_immutable()',
            tbl, tbl
        );
        EXECUTE format('DROP TRIGGER IF EXISTS %I_no_delete ON payment.%I', tbl, tbl);
        EXECUTE format(
            'CREATE TRIGGER %I_no_delete BEFORE DELETE ON payment.%I FOR EACH ROW EXECUTE FUNCTION payment.audit_trail_immutable()',
            tbl, tbl
        );
    END LOOP;
END $$;