-- name: InsertAMLEvaluation :one
INSERT INTO payment.aml_evaluations (
    id, entity_type, entity_id, user_id, transaction_type, amount, currency,
    disposition, reason_code, risk_score, rule_set_version, correlation_id, created_at
) VALUES ($1, $2, $3, $4, $5, $6::numeric, $7, $8, $9, $10, $11, sqlc.narg(correlation_id), $12)
RETURNING id;

-- name: InsertAMLCase :exec
INSERT INTO payment.aml_cases (
    id, evaluation_id, user_id, entity_type, entity_id, case_type, status,
    reason_code, correlation_id, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, 'open', $7, sqlc.narg(correlation_id), $8, $8);

-- name: ListOpenAMLCasesForExport :many
SELECT
    c.id,
    c.evaluation_id,
    c.user_id,
    c.entity_type,
    c.entity_id,
    c.case_type,
    c.status,
    c.reason_code,
    c.filing_reference,
    COALESCE(c.correlation_id, '') AS correlation_id,
    c.created_at,
    e.transaction_type,
    e.amount::text AS amount,
    e.currency,
    e.disposition,
    e.rule_set_version
FROM payment.aml_cases c
JOIN payment.aml_evaluations e ON e.id = c.evaluation_id
WHERE c.status = 'open'
  AND c.case_type = ANY($1::text[])
ORDER BY c.created_at;

-- name: CountAMLCasesByEntity :one
SELECT COUNT(*)::bigint AS count
FROM payment.aml_cases
WHERE entity_type = $1 AND entity_id = $2;