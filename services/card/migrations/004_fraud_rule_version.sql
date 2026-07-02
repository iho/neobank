ALTER TABLE card.fraud_decisions
    ADD COLUMN IF NOT EXISTS rule_set_version TEXT NOT NULL DEFAULT 'mvp-1.0.0';