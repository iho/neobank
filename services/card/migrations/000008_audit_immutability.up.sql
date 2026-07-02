CREATE OR REPLACE FUNCTION card.audit_trail_immutable()
RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'card.% is append-only', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    tbl TEXT;
BEGIN
    FOREACH tbl IN ARRAY ARRAY['audit_log', 'fraud_decisions']
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS %I_no_update ON card.%I', tbl, tbl);
        EXECUTE format(
            'CREATE TRIGGER %I_no_update BEFORE UPDATE ON card.%I FOR EACH ROW EXECUTE FUNCTION card.audit_trail_immutable()',
            tbl, tbl
        );
        EXECUTE format('DROP TRIGGER IF EXISTS %I_no_delete ON card.%I', tbl, tbl);
        EXECUTE format(
            'CREATE TRIGGER %I_no_delete BEFORE DELETE ON card.%I FOR EACH ROW EXECUTE FUNCTION card.audit_trail_immutable()',
            tbl, tbl
        );
    END LOOP;
END $$;