CREATE TABLE IF NOT EXISTS payment.reconciliation_breaks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id          UUID NOT NULL REFERENCES payment.reconciliation_runs(id),
    entity_type     TEXT NOT NULL,
    entity_id       TEXT NOT NULL,
    reason          TEXT NOT NULL,
    local_status    TEXT,
    ledger_ref      TEXT,
    status          TEXT NOT NULL DEFAULT 'open',
    resolved_by     TEXT,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ,
    CONSTRAINT payment_reconciliation_breaks_status_check
        CHECK (status IN ('open', 'investigated', 'closed'))
);

-- One active break per entity+reason; closed breaks may re-open as new rows.
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_reconciliation_breaks_active
    ON payment.reconciliation_breaks (entity_type, entity_id, reason)
    WHERE status IN ('open', 'investigated');

CREATE INDEX IF NOT EXISTS idx_payment_reconciliation_breaks_status
    ON payment.reconciliation_breaks (status, created_at DESC);