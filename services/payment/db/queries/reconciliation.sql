-- name: StartReconciliationRun :one
INSERT INTO payment.reconciliation_runs (id, started_at, status)
VALUES ($1, $2, 'running')
RETURNING id;

-- name: FinishReconciliationRun :exec
UPDATE payment.reconciliation_runs
SET finished_at = $2, checked_count = $3, break_count = $4, breaks = $5, status = $6
WHERE id = $1;

-- name: ListTransfersForReconciliation :many
SELECT id, ledger_transfer_id, status
FROM payment.transfers
WHERE status IN ('completed', 'failed') AND ledger_transfer_id IS NOT NULL
ORDER BY created_at DESC
LIMIT $1;
