-- name: StartReconciliationRun :one
INSERT INTO card.reconciliation_runs (id, started_at, status)
VALUES ($1, $2, 'running')
RETURNING id;

-- name: FinishReconciliationRun :exec
UPDATE card.reconciliation_runs
SET finished_at = $2, checked_count = $3, break_count = $4, breaks = $5, status = $6
WHERE id = $1;

-- name: ListAuthorizationsForReconciliation :many
SELECT id, card_id, user_id, currency, ledger_hold_id, ledger_transfer_id, status
FROM card.authorizations
WHERE status IN ('authorized', 'captured', 'declined') AND ledger_hold_id IS NOT NULL
ORDER BY created_at DESC
LIMIT $1;
