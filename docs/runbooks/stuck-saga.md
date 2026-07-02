# Stuck saga response

## Detect

- `neobank_saga_alerts_open` metric or alert `NeobankSagaAlertsOpen`
- `make list-saga-alerts` / `make saga-watchdog`

## Response

Sagas are **not** auto-resumed. Inspect `*.saga_instances` and `*.saga_alerts` for the `saga_instance_id`.

1. Identify whether the client can safely retry with the **same** `Idempotency-Key`.
2. If compensated, close the alert after verifying ledger/outbox state.
3. If a step is stuck mid-flight, escalate — manual compensation may be required.

Never delete saga rows; update `saga_alerts.alert_status` to `resolved` only after the business outcome is confirmed.