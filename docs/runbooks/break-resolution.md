# Reconciliation break resolution

## Detect

- Prometheus alert `NeobankReconciliationBreaks`
- CronJob logs: `payment-reconcile` / `card-reconcile` exit 1
- SQL: `SELECT * FROM payment.reconciliation_breaks WHERE status = 'open'`

## Resolve

```bash
make list-payment-breaks
make list-card-breaks

cd services/payment && go run ./cmd/resolve-break -id <break-uuid>
cd services/card && go run ./cmd/resolve-break -id <break-uuid>
```

Document the root cause in the break `resolution_notes` field. Re-run reconciliation to confirm zero open breaks.