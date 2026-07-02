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

-- name: UpsertReconciliationBreak :exec
INSERT INTO payment.reconciliation_breaks (
    id, run_id, entity_type, entity_id, reason, local_status, ledger_ref, status, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, sqlc.narg(local_status), sqlc.narg(ledger_ref), 'open', $6, $6)
ON CONFLICT (entity_type, entity_id, reason) WHERE (status IN ('open', 'investigated'))
DO UPDATE SET
    run_id = EXCLUDED.run_id,
    local_status = EXCLUDED.local_status,
    ledger_ref = EXCLUDED.ledger_ref,
    updated_at = EXCLUDED.updated_at;

-- name: ResolveReconciliationBreak :execrows
UPDATE payment.reconciliation_breaks
SET status = $2,
    resolved_by = $3,
    notes = sqlc.narg(notes),
    resolved_at = CASE WHEN $2 = 'closed' THEN $4 ELSE resolved_at END,
    updated_at = $4
WHERE id = $1 AND status != 'closed';

-- name: ListOpenReconciliationBreaks :many
SELECT id, run_id, entity_type, entity_id, reason, local_status, ledger_ref, status,
       COALESCE(resolved_by, '') AS resolved_by, COALESCE(notes, '') AS notes,
       created_at, updated_at, resolved_at
FROM payment.reconciliation_breaks
WHERE status IN ('open', 'investigated')
ORDER BY created_at DESC
LIMIT $1;
