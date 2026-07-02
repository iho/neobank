# Regulatory & Traceability Review — TODO

Analysis of the neobank monorepo from the point of view of a regulator/auditor asking:
*"Show me the complete lifecycle of this transaction, who initiated it, what decisions
were made along the way, and prove the record hasn't been tampered with."*

## Verdict on CQRS and Event Sourcing

**Event sourcing: no (not as an application pattern).** The financial source of truth is
already event-sourced — goledger is an immutable double-entry ledger, and only it mutates
balances. Rebuilding Payment/Card/User as event-sourced aggregates would duplicate that
guarantee at high cost (snapshots, upcasters, replay infrastructure, harder onboarding)
without adding regulatory value. What regulators actually require is **auditability**:
an append-only, tamper-evident record of state transitions and decisions. You can get
that with append-only history tables + a durable outbox, which this codebase is 80% of
the way to already.

**CQRS: yes, but the lightweight form only.** A single read model for unified transaction
history (currently the gateway fans out to Payment + Card and merges in memory) fed by
the existing outbox events is worth doing — it's already on the roadmap. Do **not** split
into separate command/query services or introduce a separate query database per service;
at this scale it adds operational surface without benefit.

## What is already good

- Double-entry ledger as the only balance mutator (`ledger_transfer_id` /
  `ledger_hold_id` foreign references give financial traceability).
- Transactional outbox in every service — events commit atomically with domain changes.
- Persisted saga state (`saga_instances.completed_steps` + `context`) — partial evidence
  of multi-step flows.
- Idempotency keys end-to-end with DB uniqueness constraints.
- Schema-per-service isolation; `pkg/otel` bootstrap exists.

## Gaps (ordered by regulatory severity)

### 1. No correlation/causation trail — HIGH — ✅ DONE
`events.Envelope` defines `CorrelationID`/`CausationID`, but nothing populates them:
`outbox.Record` and the `outbox_events` tables have no columns for them, and
`outbox.Worker.flush` builds envelopes without them. There is no request-ID middleware
in the gateway. A regulator cannot trace "this API call → this fraud decision → this
ledger transfer → this notification".

- [x] Add gateway middleware generating/propagating a correlation ID — `pkg/reqctx`
      (`Middleware`, `Transport` for outbound HTTP, gRPC client interceptor in
      `pkg/grpcutil`), wired into gateway + user/payment/card/notification routers.
- [x] Add `correlation_id`, `causation_id` columns to all `outbox_events` tables and
      thread them through `BuildRecord` → `Worker.flush` → `Envelope`
      (`services/*/migrations/00*_traceability.sql`).
- [x] Include correlation ID in goledger calls (gRPC metadata via `grpcutil.Dial`'s
      interceptor) and in the outbox `Envelope`. Structured logs still only carry it
      where handlers explicitly log — no blanket logging middleware yet (see #9).

### 2. State transitions are destructive UPDATEs — HIGH — ✅ DONE (app-level; DB grants still open)
`transfers.status`, `cards.status`, `kyc_cases.status`, `authorizations.status` are
updated in place. Only the final state survives; there is no record of *when* a transfer
went pending→completed, *who* froze a card, or *why* KYC flipped. Auditors ask for
lifecycle, not snapshots.

- [x] Add a generic `audit_log(entity_type, entity_id, action, from_status, to_status,
      actor, correlation_id, metadata, created_at)` per schema (`pkg/audit` +
      `services/*/internal/adapter/sqlc/audit_repository.go`), written in the same tx as
      every status-changing mutation (transfer create/complete/fail, card issue/freeze/
      unfreeze, authorization create/hold/capture/fail, KYC submit/approve, wallet
      provision).
- [ ] Still open: enforce append-only at the DB level (separate role with INSERT-only
      grants, no UPDATE/DELETE on audit tables) — requires a deployment/ops decision on
      DB roles, not just application code.
- [x] Record the actor on every mutation — `audit.Resolve` defaults `Actor` from
      `reqctx.Actor(ctx)` (falls back to `"system"`); nothing currently sets an actor
      into context from the JWT claims, so today every row records `"system"`. Wiring
      the authenticated user ID into `reqctx.WithActor` in gateway/service middleware is
      a small follow-up.

### 3. Fraud decisions are not persisted — HIGH — ✅ DONE
`pkg/fraud.Checker.Evaluate` returns a decision that is acted on and discarded. Regulators
(and disputes) require the decision record: inputs, rule versions, outcome, reason code.

- [x] Persist every fraud evaluation (allow *and* deny) with input snapshot, decision,
      reason code, risk score, correlation ID — `payment.fraud_decisions` /
      `card.fraud_decisions`, written from the `fraud_check` saga step in
      `p2p_transfer.go`, `issue_card.go`, `authorize_transaction.go`.
- [ ] Still open: version the rule configuration so past decisions can be explained
      (`pkg/fraud.Checker` has no rule-set version today).

### 4. KYC has no evidence trail — HIGH — partially done
KYC still auto-approves via stub screening, but submission evidence and screening
audit rows are now persisted.

- [x] Store KYC submission artifacts — `user.kyc_submissions` (document type/number,
      provider, provider_reference, provider_response JSON, screening decision/reason,
      correlation_id); `kyc_cases.decided_by` set on approve/reject.
- [x] Sanctions/PEP screening hook — `pkg/screening` stub at KYC onboarding and P2P
      counterparty checks; persisted to `user.screening_checks` and
      `payment.screening_checks`.
- [x] AML transaction-monitoring layer distinct from fraud (structuring/threshold rules,
      case creation, SAR/CTR export format) — `pkg/amlmonitor`, `payment.aml_evaluations` /
      `payment.aml_cases`, post-ledger hook in P2P transfer, `cmd/aml-export`.

### 5. Outbox is not an archive — MEDIUM — still open
`MarkPublished` mutates rows; nothing forbids deletion; retention is undefined. If the
outbox doubles as the domain event history, it must be durable and tamper-evident.

- [ ] Define retention: keep events ≥ 5–7 years (typical financial retention); partition
      `outbox_events` by month and archive partitions to object storage with WORM/object-lock.
- [ ] Never `DELETE`; move `published_at` tracking to a separate table or accept the
      single mutable column but revoke broader UPDATE.
- [ ] Optional tamper evidence: per-stream hash chain or periodic Merkle root anchoring.

This needs an infra/vendor decision (object storage + WORM policy, retention period per
jurisdiction) that isn't ours to make in code — flagging for a product/compliance call.

### 6. No reconciliation — MEDIUM — ✅ DONE
Service tables and goledger can drift (saga compensation failures, crashes between steps).

- [x] Reconciliation job: `payment.transfers` ↔ goledger transfers (`ledger.GetTransfer`),
      `card.authorizations` ↔ goledger holds (`ledger.ListHoldsByAccount`, since goledger
      has no `GetHold`-by-ID — the card job resolves each authorization's ledger account
      via the user service and caches per user+currency for the run).
      `services/payment/cmd/reconcile`, `services/card/cmd/reconcile`; `make
      reconcile-payment` / `make reconcile-card`. Not yet scheduled — needs a cron/k8s
      CronJob entry in deployment config.
- [x] Persist reconciliation runs — `payment.reconciliation_runs` /
      `card.reconciliation_runs` (started_at, finished_at, checked_count, break_count,
      breaks JSON, status). Break *resolution* tracking (marking a break as
      investigated/closed) is not built — today it's read-only evidence.
- [x] Also fixed in passing: `MarkFailed` in `p2p_transfer.go` (and the equivalent in
      `authorize_transaction.go`) now runs inside the same tx as its audit-log insert,
      and the error is propagated instead of discarded.

### 7. PII and data protection — MEDIUM — partially done
- [ ] Field-level encryption (or pgcrypto/KMS envelope) for `document_number`, DOB, phone —
      still open; needs a KMS/key-management decision, not just an app-code change.
- [ ] Audit access to PII (who read which customer record) — still open; the `audit_log`
      table built for #2 only covers writes, not reads.
- [ ] GDPR export/delete workflows — still open; document that deletion must be *masking*,
      not row deletion, because financial records must be retained.
- [x] Dev-auth bypasses hardened — see #7b below.

### 7b. Dev auth bypass reachable in production — HIGH — ✅ DONE
The `X-User-Id` header and legacy `access.<user-id>.*` token both skip real JWT
validation with no environment guard.

- [x] Gated behind `cfg.AllowDevAuth` (`services/gateway/internal/config`), true only when
      `APP_ENV` is `development`/`local`/`dev` (default `development` — **set `APP_ENV=production`
      in any real deployment**). `resolveUserID` in `server.go` now rejects both paths
      when `allowDevAuth` is false. Covered by
      `TestResolveUserIDLegacyDevTokenBlockedOutsideDev` /
      `TestResolveUserIDXUserIDBlockedOutsideDev`.

### 8. Event hygiene — LOW — mostly done
- [x] `EventVersion` is no longer hardcoded — `outbox.BuildRecord` reads `Event.Version()`,
      persists it, and `Worker.flush` uses the stored value (falling back to 1 for rows
      written before the migration).
- [ ] Still open: document the event catalog as a contract (schema registry or versioned
      JSON schemas); regulators care that replayed history is interpretable years later.
- [x] Consumer-side inbox/dedup for at-least-once delivery (notification service) —
      `notification.consumer_inbox` event-level dedup in `IngestEventUseCase`; per-user
      notification rows still use `ON CONFLICT (event_id, user_id) DO NOTHING`.

### 9. Observability wiring — LOW — partially done
- [x] Wire `pkg/otel` through gateway → services → ledger client — gated on
      `OTEL_EXPORTER_OTLP_ENDPOINT`; `otel.HTTPMiddleware` on all services,
      `otel.OutboundTransport` on HTTP clients, `otelgrpc` on gRPC ledger dial.
      Collector deployment still open (needs infra).
- [x] Structured HTTP access logs — `pkg/sloghttp` middleware on gateway and all
      services; logs `correlation_id`, `user_id`, `idempotency_key`, status, duration.
      `sloghttp.Logger(ctx)` helper for handler/worker logs. Retained log shipping
      still open (needs infra).
- [x] The outbox worker no longer swallows flush errors — `Worker.flush` errors are now
      logged via an injected `*slog.Logger` (`Worker.WithLogger`).

## Suggested order of work — status

1. ~~Correlation ID propagation + outbox columns (#1)~~ — done.
2. ~~Append-only audit/history tables + actor recording (#2)~~ — done (actor defaults to
   `"system"` until JWT identity is threaded into `reqctx.WithActor`).
3. ~~Persist fraud decisions (#3)~~ — done.
4. ~~Reconciliation job (#6)~~ — done, not yet scheduled.
5. Outbox retention/archival (#5) — open, needs an infra/compliance decision.
6. KYC/AML evidence model (#4) — KYC evidence + screening done; AML txn-monitoring stub done;
   real vendor integration still needed.
7. PII encryption (#7) — open, needs a KMS decision. Dev-auth hardening (#7b) — done.
8. Light CQRS read model for `/v1/wallet/transactions` fed from outbox events — ✅ DONE
   (`user.wallet_transactions`, `pkg/walletprojection`, User service ingest + list API,
   gateway reads User service; payment/card outbox fan-out via `ProjectionURLs`).
