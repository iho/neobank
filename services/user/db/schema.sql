CREATE SCHEMA "user";

CREATE TABLE "user".users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    phone         TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE "user".profiles (
    user_id       UUID PRIMARY KEY REFERENCES "user".users(id),
    full_name     TEXT,
    date_of_birth DATE,
    country_code  CHAR(2)
);

CREATE TABLE "user".kyc_cases (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES "user".users(id),
    status           TEXT NOT NULL DEFAULT 'pending',
    submitted_at     TIMESTAMPTZ,
    decided_at       TIMESTAMPTZ,
    rejection_reason TEXT,
    decided_by       TEXT
);

CREATE TABLE "user".kyc_submissions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kyc_case_id        UUID NOT NULL REFERENCES "user".kyc_cases(id),
    user_id            UUID NOT NULL REFERENCES "user".users(id),
    document_type      TEXT,
    document_number    TEXT,
    provider           TEXT NOT NULL,
    provider_reference TEXT,
    provider_response  JSONB NOT NULL DEFAULT '{}',
    screening_decision TEXT NOT NULL,
    screening_reason   TEXT,
    correlation_id     TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE "user".screening_checks (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    check_type         TEXT NOT NULL,
    subject_user_id    UUID NOT NULL,
    entity_type        TEXT NOT NULL,
    entity_id          TEXT NOT NULL,
    decision           TEXT NOT NULL,
    reason_code        TEXT NOT NULL,
    provider           TEXT NOT NULL,
    provider_reference TEXT,
    raw_response       JSONB NOT NULL DEFAULT '{}',
    correlation_id     TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE "user".wallets (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES "user".users(id),
    currency          CHAR(3) NOT NULL DEFAULT 'USD',
    ledger_account_id TEXT NOT NULL UNIQUE,
    status            TEXT NOT NULL DEFAULT 'active',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, currency)
);

CREATE TABLE "user".outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  TEXT NOT NULL,
    aggregate_id    TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    event_version   INT NOT NULL DEFAULT 1,
    payload         JSONB NOT NULL,
    correlation_id  TEXT,
    causation_id    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE "user".outbox_publications (
    event_id      UUID PRIMARY KEY REFERENCES "user".outbox_events(id),
    published_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_outbox_publications_published_at
    ON "user".outbox_publications (published_at DESC);

CREATE TABLE "user".saga_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'running',
    current_step    TEXT NOT NULL DEFAULT '',
    context         JSONB NOT NULL DEFAULT '{}',
    completed_steps JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE "user".saga_alerts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_instance_id  UUID NOT NULL REFERENCES "user".saga_instances(id),
    saga_type         TEXT NOT NULL,
    idempotency_key   TEXT NOT NULL,
    instance_status   TEXT NOT NULL,
    alert_status      TEXT NOT NULL DEFAULT 'open',
    stuck_since       TIMESTAMPTZ NOT NULL,
    last_seen_at      TIMESTAMPTZ NOT NULL,
    completed_steps   JSONB NOT NULL DEFAULT '{}',
    context           JSONB NOT NULL DEFAULT '{}',
    resolved_by       TEXT,
    notes             TEXT,
    resolved_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_saga_alerts_status_check
        CHECK (alert_status IN ('open', 'investigating', 'resolved'))
);

CREATE UNIQUE INDEX idx_user_saga_alerts_active
    ON "user".saga_alerts (saga_instance_id)
    WHERE alert_status IN ('open', 'investigating');

-- Append-only lifecycle trail for KYC decisions and wallet provisioning.
CREATE TABLE "user".audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type     TEXT NOT NULL,
    entity_id       TEXT NOT NULL,
    action          TEXT NOT NULL,
    from_status     TEXT,
    to_status       TEXT,
    actor           TEXT NOT NULL DEFAULT 'system',
    correlation_id  TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_audit_log_entity
    ON "user".audit_log (entity_type, entity_id, created_at);

-- Append-only trail for reads of customer PII (profile, KYC, wallet lookups).
CREATE TABLE "user".pii_access_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_user_id UUID NOT NULL,
    resource        TEXT NOT NULL,
    actor           TEXT NOT NULL DEFAULT 'system',
    correlation_id  TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_pii_access_log_subject
    ON "user".pii_access_log (subject_user_id, created_at);

CREATE INDEX idx_user_pii_access_log_actor
    ON "user".pii_access_log (actor, created_at);

CREATE TABLE "user".wallet_transactions (
    user_id         UUID NOT NULL REFERENCES "user".users(id),
    id              TEXT NOT NULL,
    source_event_id UUID NOT NULL,
    tx_type         TEXT NOT NULL,
    amount          TEXT NOT NULL,
    currency        TEXT NOT NULL,
    direction       TEXT NOT NULL,
    status          TEXT NOT NULL,
    counterparty    TEXT,
    memo            TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, id),
    CONSTRAINT user_wallet_tx_type_check
        CHECK (tx_type IN ('p2p_out', 'p2p_in', 'card_hold', 'card_purchase')),
    CONSTRAINT user_wallet_tx_direction_check
        CHECK (direction IN ('debit', 'credit'))
);

CREATE UNIQUE INDEX idx_user_wallet_tx_event_dedup
    ON "user".wallet_transactions (user_id, source_event_id);

CREATE INDEX idx_user_wallet_tx_user_created
    ON "user".wallet_transactions (user_id, created_at DESC);

CREATE TABLE "user".consumer_inbox (
    event_id      UUID PRIMARY KEY,
    event_type    TEXT NOT NULL,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_consumer_inbox_type
    ON "user".consumer_inbox (event_type, processed_at DESC);
