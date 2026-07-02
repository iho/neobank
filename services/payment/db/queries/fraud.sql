-- name: InsertFraudDecision :exec
INSERT INTO payment.fraud_decisions (
    id, entity_type, entity_id, user_id, transaction_type, amount, currency,
    decision, reason_code, risk_score, rule_set_version, correlation_id, created_at
) VALUES ($1, $2, $3, $4, $5, $6::numeric, $7, $8, $9, $10, $11, sqlc.narg(correlation_id), $12);

-- name: ListFraudDecisionsByEntity :many
SELECT id, entity_type, entity_id, user_id, transaction_type, amount::text AS amount, currency,
       decision, reason_code, risk_score, rule_set_version,
       COALESCE(correlation_id, '') AS correlation_id, created_at
FROM payment.fraud_decisions
WHERE entity_type = $1 AND entity_id = $2
ORDER BY created_at;
