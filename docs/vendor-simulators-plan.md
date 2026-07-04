# Vendor Simulators Plan

Make the neobank a complete "bank in a box" by simulating the four external vendor
surfaces a real neobank integrates with: payment rails, card processor, KYC vendor,
and FX rates. Each simulator speaks the same protocol shape a real vendor would, so
swapping to a real vendor later is a config change, not a rewrite.

## Principles

1. **Simulators are outsiders.** Each one is a separate service under
   `services/simulators/` with its own DB schema (or in-memory state). It talks to
   domain services only via REST/webhooks — never shared Go packages for domain
   logic, never a shared database.
2. **Domain services keep ports.** Every call to a simulator goes through an
   interface in `internal/port/` with the simulator client as one adapter. The real
   vendor becomes a second adapter later.
3. **Webhooks are realistic.** HMAC-signed payloads, at-least-once delivery with
   retries and backoff, occasional out-of-order delivery. Consumers must be
   idempotent (reuse `pkg/idempotency`).
4. **Magic values drive outcomes.** Like Stripe test cards: specific inputs
   deterministically trigger rejections, delays, returns, chargebacks. Documented
   per simulator, used heavily by integration tests.
5. **Money only moves through goledger.** Every simulated external flow maps to
   double-entry postings against a system account (nostro/settlement, fees, FX
   position). No simulator ever writes ledger rows directly.

## Current state (what gets replaced)

| Today | Where | Replaced by |
|-------|-------|-------------|
| Synchronous demo deposit minting from `DEPOSIT_SOURCE_LEDGER_ACCOUNT_ID` | `services/user/internal/usecase/deposit_wallet.go` | Rails simulator webhook → payment service |
| Card auth/capture invoked by our own API | `services/card/internal/usecase/authorize_transaction.go`, `capture_authorization.go` | Card processor simulator originating webhooks |
| KYC auto-approve | `services/user/internal/usecase/submit_kyc.go` | KYC simulator with async verdicts |
| No FX / single-currency wallets | — | FX simulator + per-currency ledger accounts (goledger already validates ISO 4217 per account) |

## Phase 0 — Shared infrastructure: `pkg/vendorsim`

**Status: core package done.** Implemented in `pkg/vendorsim/` (8 files, 21 tests,
lint-clean):

- `sign.go` — HMAC-SHA256 signing/verification over `<timestamp>.<body>`,
  standard headers (`X-Vendorsim-Timestamp`, `-Signature`, `-Event`, `-Delivery-Id`).
- `delivery.go` / `memory_store.go` — `Delivery` model and `DeliveryStore`
  interface (`Enqueue`/`ClaimDue`/`MarkDelivered`/`MarkFailed`/`List`/`Get`),
  with an in-memory implementation for local dev and tests. Each simulator
  backs this with Postgres when it needs durability (Phase 1+).
- `dispatcher.go` — `Dispatcher` delivers signed webhooks with exponential
  backoff (`BackoffConfig`, default 2s→5m, 10 retries) and a background
  `Run` loop; `Enqueue` is what simulators call to schedule a webhook.
- `chaos.go` — `ChaosConfig` (delay range, duplicate probability, reorder
  probability) plus `ChaosConfigFromEnv`; off by default, opt-in per test/env.
- `middleware.go` — consumer-side `VerifyWebhook` HTTP middleware: signature
  check + `NonceStore`-backed replay short-circuit (in-memory impl provided;
  domain services can back it with Redis the same way `pkg/idempotency` does).
- `magicvalues.go` — shared `ContainsToken` / `AmountEndsInCents` helpers and
  the documented convention (REJECT/REVIEW/RETURN tokens, `.13`/`.99` cent
  suffixes) every simulator and integration test should reuse.

**Not yet done** (deferred to when the first simulator needs them, to avoid
building against a guessed shape):

- `deployments/docker-compose.simulators.yml` — held back until Phase 1
  (rails) exists; an empty compose file has nothing to run.
- Per-simulator service skeletons (`cmd/server`, `internal/{config,domain,
  usecase,adapter,port}`, `db/`, `sqlc.yaml`) — each phase creates its own,
  following the existing `services/card` layout, wired to `pkg/vendorsim`.
- Postgres-backed `DeliveryStore` — first simulator to need cross-restart
  durability implements it against its own schema (see Principle 1: no shared
  DB), reusing the `MemoryDeliveryStore` tests as a spec.

The rest of this phase's original scope (below) still describes intent for
what each simulator will build on top of the package above.

- **Webhook delivery**: signed (HMAC-SHA256 over timestamp + body, `X-Sim-Signature`
  header), persistent outbox per simulator (reuse `pkg/outbox` pattern), retries
  with exponential backoff, delivery log queryable via admin API.
- **Webhook verification** middleware for consumers: signature check, timestamp
  skew window, replay protection.
- **Scenario controls**: every simulator exposes an admin API (`/sim/...`) for
  tests and manual poking — inject events, list pending deliveries, force
  redelivery, reset state.
- **Magic value registry**: shared conventions, e.g. amount cents `.13` → decline,
  reference containing `RETURN` → bounce, surname `REVIEW` → manual review.
- **Chaos knobs** (env-configurable): delivery delay range, duplicate-delivery
  probability, out-of-order probability. Default off; integration tests turn them on.

Also in this phase:
- `deployments/docker-compose.simulators.yml` running all simulators.
- Skeleton generator: each simulator follows the existing service layout
  (`cmd/server`, `internal/{config,domain,usecase,adapter,port}`, `api/` OpenAPI).

## Phase 1 — Payment rails simulator (`services/simulators/rails`)

Simulates a SEPA/ACH-style sponsor-bank connection. Biggest realism gain: removes
the "mint money out of thin air" deposit.

**Simulator responsibilities**
- Issue a virtual IBAN per wallet on request (`POST /v1/accounts` → IBAN).
- Admin API to inject an inbound transfer: `POST /sim/inbound-transfers`
  (IBAN, amount, currency, sender name, reference). Delivers
  `rails.transfer.received` webhook after a configurable delay.
- Accept outbound payment orders: `POST /v1/payments` → `accepted`, then async
  `rails.payment.settled` or `rails.payment.returned` webhook (magic values:
  reference `RETURN` bounces after settlement, amount `.99` fails validation
  asynchronously).
- End-of-day statement endpoint: `GET /v1/statements/{date}` listing all movements
  — this is the reconciliation source of truth, feeding the existing recon jobs.

**Neobank-side changes**
- Payment service gains a webhook consumer (`POST /webhooks/rails`) and two new
  sagas:
  - *Inbound*: webhook → idempotency check → fraud/AML screen → ledger transfer
    `rails settlement account → user wallet` → outbox event → notification.
  - *Outbound*: user request → fraud screen → ledger hold on wallet →
    `POST /v1/payments` to simulator → on `settled` webhook: capture hold into
    settlement account; on `returned`: release hold (or reverse if already
    captured — the return-after-settlement case is the interesting saga).
- Ledger: one settlement account per rail/currency (e.g. `rails:sepa:EUR`),
  created by bootstrap (see operator section). Existing
  `SETTLEMENT_LEDGER_ACCOUNT_ID` generalizes to per-rail config.
- User service: `deposit_wallet.go` demo endpoint becomes dev-only or is removed;
  gateway exposes the user's virtual IBAN instead ("top up by bank transfer").
- Reconciliation: extend existing recon to compare simulator statements against
  ledger settlement-account entries; breaks flow into the existing break-resolution
  tooling.

**Done when**: an injected inbound transfer lands in a wallet with full audit
trail; an outbound payment settles or bounces correctly; recon over a day of
simulated traffic reports zero breaks; duplicate/out-of-order webhook delivery
causes no double-credit (integration test with chaos knobs on).

## Phase 2 — Card processor simulator (`services/simulators/cardproc`)

Inverts the current flow: today our API calls authorize/capture directly; with the
simulator, an "external processor" originates those events — which is how Marqeta/
Visa actually behave.

**Simulator responsibilities**
- `POST /v1/cards` — card service calls this during issuance; simulator returns
  processor card ID, PAN token, expiry.
- Admin API to simulate merchant activity: `POST /sim/transactions` (card token,
  amount, MCC, merchant name, type: auth / auth+capture / refund). Simulator then
  drives the lifecycle via webhooks to the card service:
  `card.authorization.requested` (expects sync approve/decline response — the
  real-time auth flow), then `card.captured`, `card.auth.reversed`,
  `card.auth.expired` (if not captured within TTL), `card.chargeback.opened`.
- Magic values: amount `.13` → force our decline path exercised, MCC `7995`
  (gambling) → tests card controls, merchant `CHARGEBACK` → chargeback 1 simulated
  day after capture.
- Partial captures and multi-capture support (airline/hotel patterns).

**Neobank-side changes**
- Card service: issuance saga calls the simulator; new webhook endpoint
  (`POST /webhooks/cardproc`). Existing `authorize_transaction.go` /
  `capture_authorization.go` usecases become the handlers behind the webhook
  consumer instead of being exposed on our public API. The real-time auth webhook
  must respond within a deadline (e.g. 2s) with approve/decline — fraud rules and
  card controls run inside that window; timeout = decline (stand-in processing is
  a later phase).
- New flows to implement: auth expiry (release hold), chargeback (provisional
  credit ledger flow + dispute record).
- Ledger: card settlement account per currency (`cardproc:USD`); captures move
  money wallet → card settlement.

**Done when**: `POST /sim/transactions` end-to-end produces correct holds,
settlements, and history entries; auth expiry releases holds; a chargeback
produces provisional credit postings; declines happen for frozen cards, controls
violations, and insufficient funds.

## Phase 3 — KYC vendor simulator (`services/simulators/kyc`)

Mimics an Onfido/Sumsub-style identity vendor with async verdicts.

**Simulator responsibilities**
- `POST /v1/applicants` (PII payload) → applicant ID, status `pending`.
- `POST /v1/applicants/{id}/documents` — fake document upload (metadata only).
- Async verdict webhook `kyc.check.completed` after a configurable delay:
  `approved` / `rejected` / `manual_review` — driven by magic values (surname
  `REJECT` / `REVIEW`; DOB making applicant under 18 → rejected; anything else →
  approved). Manual-review cases resolvable via admin API
  (`POST /sim/reviews/{id}/resolve`) to mimic a human agent.
- `GET /v1/applicants/{id}` for polling fallback.

**Neobank-side changes**
- User service: `submit_kyc.go` stops auto-approving; it submits to the vendor via
  a `port.KYCVendor` interface and parks the user in `kyc_pending`. Webhook
  consumer advances the state machine: `approved` → run the existing wallet
  provisioning saga; `rejected` → terminal state + notification; `manual_review` →
  waiting state surfaced in (future) back-office.
- Screening tie-in: keep the existing sanctions/PEP stub (`pkg/screening`) as a
  separate step — vendor verdict and screening are independent gates, matching
  reality.
- Mobile: KYC screen already exists; it gains a "verification in progress" state
  (poll or push notification on verdict).

**Done when**: registration → KYC → provisioning works with a genuinely async
verdict; each magic-value branch has an integration test; a user stuck in
`manual_review` can be resolved via the admin API and completes provisioning.

## Phase 4 — FX / rates simulator (`services/simulators/fx`)

Enables multi-currency wallets and conversion. goledger already supports
per-account ISO 4217 currencies, so this is neobank-side work, not ledger work.

**Simulator responsibilities**
- `GET /v1/rates?base=EUR&quote=USD` → mid rate. Deterministic random-walk around
  seeded mids so tests are stable-ish but charts look alive.
- `POST /v1/quotes` (pair, amount, side) → quote ID, rate with spread applied,
  expiry (e.g. 30s). `POST /v1/quotes/{id}/execute` → executed or `quote_expired`.
- Historical rates endpoint for charting.

**Neobank-side changes**
- Wallet model: user can hold multiple wallets, one per currency (provisioning
  saga parameterized by currency; ledger accounts per currency already work).
- New payment-service usecase `convert`: get quote → show to user → execute within
  TTL → ledger postings: user EUR wallet → FX position account (EUR side); FX
  position account (USD side) → user USD wallet; spread margin → fee income
  account. Quote ID recorded on the transaction for auditability.
- Recon: FX position accounts should net to the executed quotes; add to recon job.
- Gateway/mobile: currency selector, conversion screen with countdown on quote TTL.

**Done when**: a user converts EUR→USD at a quoted rate, both wallet histories
show the conversion with the quote ID, expired quotes are rejected, and the FX
position + fee accounts reconcile against executed quotes.

## Testing strategy

- Each simulator ships with its own unit tests, but the real value is
  **scenario tests** in `tests/`: script the simulators' admin APIs to compose
  flows ("deposit arrives → card auth → capture → chargeback → recon clean").
- Run integration suites twice in CI: once with chaos knobs off (fast,
  deterministic) and a nightly run with duplicates/delays/reordering on.
- Simulator admin APIs replace most hand-rolled test fixtures for money movement.

## Sequencing

| Phase | Depends on | Rough size |
|-------|-----------|------------|
| 0 — `pkg/vendorsim` + compose | — | small |
| 1 — Rails | 0 | large (new sagas + recon) |
| 2 — Card processor | 0 | large (flow inversion + chargebacks) |
| 3 — KYC | 0 | medium (state machine + async) |
| 4 — FX | 0, benefits from 1 | medium (multi-currency plumbing) |

Phases 1–4 are independent enough to reorder, but rails-first removes the least
realistic part of the system (minted deposits) and forces the webhook-consumer
pattern everything else reuses.

## Kubernetes operators

What genuinely fits the operator pattern here (declarative desired state +
reconcile loop) vs. what should stay a service or job:

**Good operator candidates**

1. **Ledger bootstrap operator** — the strongest fit. A `LedgerAccount` CRD
   declares system accounts ("settlement `rails:sepa:EUR`", "fee income USD",
   "FX position EUR/USD"); the operator reconciles against goledger's API,
   creates missing accounts, and publishes the resulting account IDs into a
   ConfigMap/Secret that services consume. This kills the current hand-copied
   ULIDs (`SETTLEMENT_LEDGER_ACCOUNT_ID`, `DEPOSIT_SOURCE_LEDGER_ACCOUNT_ID`)
   and makes new environments self-provisioning. Idempotent, safe to retry,
   no money movement — ideal reconcile semantics.
2. **Environment operator** — a `NeobankEnv` CRD that stands up a full
   bank-in-a-box (services + simulators + migrations + ledger bootstrap) per
   namespace, for PR preview environments and load-test sandboxes. Reconciles
   the whole stack, tears it down on CR deletion.
3. **Scenario operator (optional, later)** — `SimScenario` CRDs describing
   simulator traffic ("50 deposits/min, 2% returns, chargebacks after 1 day")
   that the operator drives against simulator admin APIs; useful for soak
   environments. A CronJob gets 80% of this value for 20% of the effort — only
   build the CRD version if scenario definitions need to live in Git per env.

**Use existing operators, don't write these**: Postgres (CloudNativePG),
Kafka/topics (Strimzi — `KafkaTopic` CRDs replace hand-managed topics), Vault
(bank-vaults or Vault's own), certificates for gRPC mTLS (cert-manager — replaces
the static files in `deployments/grpc-mtls/`).

**Anti-patterns — keep these out of operators**: anything that moves money or
executes business processes (sagas, reconciliation, outbox dispatch, the saga
watchdog). Reconcile loops are level-triggered and may re-run at any time; money
movement is event-driven and exactly-once-ish. Those stay as services and
CronJobs (`deployments/docker-compose.jobs.yml` maps to k8s CronJobs as-is).
