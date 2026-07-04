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

Postgres-backed `DeliveryStore` and per-simulator service skeletons were
deferred out of Phase 0 itself and built as part of Phase 1 below, once there
was a real simulator to shape them against
(`services/simulators/rails/internal/adapter/deliverystore`). A dedicated
`docker-compose.simulators.yml` was skipped in favor of adding each simulator
directly to `deployments/docker-compose.services.yml` — one compose file per
environment turned out simpler than one per simulator-vs-service split.

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
the "mint money out of thin air" deposit. Split into 1a (inbound, **done**) and
1b (outbound + reconciliation, **not yet built**) — see below for why.

**Phase 1a — done.** Implemented and wired end-to-end:

- `services/simulators/rails`: a standalone HTTP service (plain `chi` handlers,
  not oapi-codegen — this simulator has no mobile-facing contract to keep in
  sync, so the extra generation step wasn't worth it) with:
  - `POST /v1/accounts` — get-or-create a virtual IBAN for `(external_ref,
    currency)`, idempotent.
  - `POST /sim/inbound-transfers` — admin endpoint that records a transfer and
    schedules a `rails.transfer.received` webhook via `pkg/vendorsim.Dispatcher`.
  - `GET /v1/statements/{date}` — everything that arrived that day, independent
    of webhook delivery outcome (the recon source of truth Phase 1b will use).
  - `GET /sim/deliveries[/{id}]` — admin visibility into webhook delivery state.
  - Its own Postgres schema (`rails.accounts`, `rails.inbound_transfers`,
    `rails.webhook_deliveries`) and a Postgres-backed `vendorsim.DeliveryStore`
    (`internal/adapter/deliverystore`), so delivery retries survive restarts.
- Payment service:
  - `GET /api/v1/payments/bank-accounts?currency=USD` — user-facing endpoint
    that calls the simulator to mint/fetch a virtual IBAN and caches the
    mapping in `payment.bank_accounts`.
  - `POST /webhooks/rails` — the webhook consumer, mounted on a **separate**
    chi router (`root.Mount("/webhooks", webhookRouter)` in `cmd/server/main.go`)
    so it sits behind `vendorsim.VerifyWebhook` instead of the global
    `Idempotency-Key` middleware, which webhook deliveries don't carry.
  - `ProcessInboundTransferUseCase`: idempotent on the rail's transfer ID
    (`payment.bank_transfers.rails_transfer_id UNIQUE`) — a redelivered or
    chaos-duplicated webhook is a no-op, not a double credit. Moves funds
    `RAILS_SETTLEMENT_LEDGER_ACCOUNT_ID → user wallet` via the existing
    `ledgerclient`, then publishes `events.BankTransferReceived` in the same
    DB transaction as the local record and audit entry.
  - New event `payment.bank_transfer.received`, registered in `pkg/events`
    catalog and projected into wallet tx history by
    `pkg/walletprojection.applyBankTransferReceived` (shows up as
    `bank_transfer_in` alongside `p2p_in`/`p2p_out`).
- `deployments/docker-compose.services.yml` runs `rails` alongside the other
  services; `deployments/docker-compose.rails-override.yml` is where
  `RAILS_SETTLEMENT_LEDGER_ACCOUNT_ID` gets filled in after creating that
  ledger account (same pattern as the existing deposit-override file).

**Deliberately not done in 1a** (kept the vertical slice small and honest
rather than half-build everything):
- User service's `deposit_wallet.go` demo endpoint is untouched — it's still
  useful for seeding dev/test data, and nothing required removing it.
- No fraud/AML screening on the inbound-credit path yet (P2P transfers screen
  the counterparty; bank transfers currently don't). Worth adding before this
  is anything but a demo.
- No integration test in `tests/integration` yet driving the simulator's admin
  API end-to-end (unit tests with fakes cover the usecases in both services).

**Phase 1b — not yet built.**
- Accept outbound payment orders: `POST /v1/payments` → `accepted`, then async
  `rails.payment.settled` or `rails.payment.returned` webhook (magic values:
  reference `RETURN` bounces after settlement, amount `.99` fails validation
  asynchronously). `pkg/vendorsim.ContainsToken` / `AmountEndsInCents` are
  ready for this; the simulator just doesn't have the endpoint yet.
- Payment service outbound saga: user request → fraud screen → debit wallet →
  `POST /v1/payments` to simulator → on `settled`: no further action (funds
  already moved); on `returned`: reverse the ledger transfer back to the
  wallet — the "money bounced after it looked done" case is the interesting
  saga this unlocks.
- Reconciliation: extend the existing recon job to compare the simulator's
  `GET /v1/statements/{date}` against `payment.bank_transfers` /
  `RAILS_SETTLEMENT_LEDGER_ACCOUNT_ID` ledger entries, feeding breaks into the
  existing break-resolution tooling.
- An integration test exercising chaos (duplicate/out-of-order webhook
  delivery via `RAILS_CHAOS_*` env vars) against a real Postgres.

**Done when** (1a, met): an injected inbound transfer lands in a wallet with a
ledger transfer, an audit entry, and a wallet-history row, and a redelivered
webhook does not double-credit.
**Done when** (1b, not yet met): an outbound payment settles or bounces
correctly; recon over a day of simulated traffic reports zero breaks;
duplicate/out-of-order webhook delivery under chaos causes no double-credit
in an automated integration test (today only unit-tested with fakes).

## Phase 2 — Card processor simulator (`services/simulators/cardproc`)

Inverts the previous flow: card service used to call its own authorize/capture
logic directly (via the public gateway/mobile API — see "kept as-is" below); with
the simulator, an "external processor" originates those events instead, which is
how Marqeta/Visa actually behave. Split into 2a (issuance + real-time auth +
capture + reversal, **done**) and 2b (auth expiry, chargebacks, partial/multi
capture, **not yet built**).

**Phase 2a — done.** Implemented and wired end-to-end:

- `services/simulators/cardproc`: a standalone HTTP service (plain `chi`
  handlers, same rationale as rails) with:
  - `POST /v1/cards` — issues a virtual card (PAN token, last four, expiry);
    called by the card service during issuance.
  - `POST /v1/cards/{ref}/cancel` — card service's issuance-saga compensation
    step calls this on failure.
  - `POST /sim/transactions` — admin endpoint that simulates a merchant charge:
    creates a transaction record, then **synchronously** calls the card
    service's real-time auth webhook and waits for approve/decline (this is
    the one place in the vendor-simulator design that isn't fire-and-forget —
    matching how a real network's stand-in authorization actually works). If
    approved with `capture: true`, schedules a `card.captured` webhook via
    `pkg/vendorsim.Dispatcher`.
  - `POST /sim/transactions/{id}/capture` and `.../reverse` — settle or void
    an auth-only transaction later, async webhook either way.
  - `GET /sim/deliveries[/{id}]` — admin visibility into async delivery state.
  - Its own Postgres schema (`cardproc.cards`, `cardproc.transactions`,
    `cardproc.webhook_deliveries`) and a Postgres-backed `vendorsim.DeliveryStore`,
    same pattern as rails.
- Card service:
  - `internal/adapter/processor/httpclient.go` implements the existing
    `Processor` interface via HTTP to the simulator, swapped in for
    `processor.NewMock()` in `cmd/server/main.go` — the interface didn't need
    to change, only the wiring. `MockProcessor` is left in place, unused by
    default, in case tests want it later.
  - `POST /webhooks/cardproc/authorize` — the synchronous auth endpoint. It
    verifies the request signature inline (not via `vendorsim.VerifyWebhook`,
    since that middleware's replay de-dup doesn't apply to a call-and-response),
    resolves the card by the simulator's `card_ref` (new
    `CardRepository.GetByProcessorRef`), and runs the *existing*
    `AuthorizeTransactionUseCase` unchanged — the fraud check, card-active/
    controls/daily-limit checks, and ledger hold are the same code path the
    old public endpoint used. Mounted outside the global `Idempotency-Key`
    middleware (a bare root router, since the handler already verifies its
    own signature and the use case is already idempotent on the simulator's
    transaction ID).
  - `POST /webhooks/cardproc/events` — the async consumer for
    `card.captured` (calls the existing `CaptureAuthorizationUseCase`) and
    `card.auth.reversed` (calls the new `ReverseAuthorizationUseCase`, which
    voids the ledger hold and marks the authorization `voided`). Mounted
    behind `vendorsim.VerifyWebhook`.
  - New usecase `ReverseAuthorizationUseCase` and event
    `card.auth.voided`, registered in `pkg/events` and projected into wallet
    tx history by reusing the existing `CaptureUpdate` mechanism in
    `pkg/walletprojection` (an upsert keyed on the same authorization ID as
    the original hold row — no new query needed).
  - Magic value: amount ending in `.13` forces a deterministic decline inside
    `HandleAuthorize`, per the `pkg/vendorsim` convention, for tests that
    don't want to engineer a real decline condition.
- `deployments/docker-compose.services.yml` runs `cardproc` alongside the
  other services; card depends on it being healthy.

**Deliberately not done in 2a**:
- The public gateway/mobile `authorizeTransaction` / `captureAuthorization`
  endpoints are untouched (same call as `deposit_wallet.go` in Phase 1a) —
  they're still a convenient way to simulate a purchase without the cardproc
  simulator running, and nothing required removing them.
- No integration test in `tests/integration` yet driving the simulator's
  admin API end-to-end (unit tests with fakes and a real `httptest` server
  cover the sync-auth path in both services).

**Phase 2b — not yet built.**
- Auth expiry: a background sweep in the simulator that finds un-captured,
  un-reversed transactions past a TTL and fires `card.auth.expired` (the card
  service side can reuse `ReverseAuthorizationUseCase` for this — the event
  name is the only new thing needed).
- Chargebacks: `POST /sim/transactions/{id}/chargeback`, landing on the card
  service as a new event requiring a provisional-credit ledger flow and a
  dispute record — genuinely new state, not a reuse of existing usecases.
- Partial/multi-capture (airline/hotel patterns).
- MCC-based magic values (e.g. `7995` gambling) for exercising card controls
  — deferred because `AuthorizeTransactionUseCase` doesn't currently branch
  on MCC at all; adding that decline path is orthogonal to the simulator work.

**Done when** (2a, met): `POST /sim/transactions` end-to-end produces a real
ledger hold via the existing saga, `capture: true` settles it, and
`.../capture` / `.../reverse` drive the same outcomes for auth-only
transactions — all through the synchronous-auth-plus-async-webhook path
rather than the old direct-call shortcut.
**Done when** (2b, not yet met): auth expiry releases holds automatically; a
chargeback produces provisional-credit postings and a dispute record;
declines happen for MCC-restricted categories, not just frozen
cards/limits/funds.

## Phase 3 — KYC vendor simulator (`services/simulators/kyc`) — done

Mimics an Onfido/Sumsub-style identity vendor with async verdicts.

**Implemented and wired end-to-end:**

- `services/simulators/kyc`: a standalone HTTP service (plain `chi` handlers,
  same rationale as rails/cardproc) with:
  - `POST /v1/applicants` — submits an applicant (`external_ref`, `full_name`,
    `date_of_birth`, `country_code`). The verdict is decided *here*,
    deterministically, by the shared `pkg/vendorsim` magic-value conventions:
    a name containing `REJECT` → rejected, containing `REVIEW` → manual
    review, an applicant under 18 → rejected (`underage`), otherwise
    approved. Approved/rejected schedule the `kyc.check.completed` webhook
    immediately via `pkg/vendorsim.Dispatcher`; `manual_review` schedules
    nothing — it waits for a human.
  - `GET /v1/applicants/{id}` — polling fallback / status check.
  - `POST /sim/reviews/{id}/resolve` — the admin endpoint mimicking a human
    reviewer clearing a `manual_review` case to `approved`/`rejected`, which
    then fires the same webhook.
  - `GET /sim/deliveries[/{id}]` — admin visibility into delivery state.
  - Its own Postgres schema (`kyc.applicants`, `kyc.webhook_deliveries`) and
    a Postgres-backed `vendorsim.DeliveryStore`, same pattern as rails/cardproc.
- User service:
  - `SubmitKYCUseCase` no longer auto-approves. Sanctions/PEP screening
    (`pkg/screening`) stays a separate, instant hard-stop — it can still
    reject synchronously without ever calling the vendor. If screening
    passes, the case submits to the vendor and returns `pending`; no wallet
    is provisioned in this call anymore.
  - New `ProcessKYCVerdictUseCase` (called from the webhook consumer):
    `approved` runs the *existing* `ProvisionWalletUseCase` and publishes the
    *existing* `events.KYCApproved`; `rejected` publishes the existing
    `events.KYCRejected`; `manual_review` just flips the case's status. No
    new event types were needed — the async path reuses exactly what the old
    synchronous path already published.
  - `POST /webhooks/kyc/events` — mounted on a bare root router outside the
    global `Idempotency-Key` middleware, behind `vendorsim.VerifyWebhook`,
    same pattern as the rails/cardproc async webhooks. Idempotent by
    construction: `ProcessKYCVerdictUseCase` no-ops if the case is already in
    a terminal state (handles a redelivered or duplicated webhook).
  - Migration adds `kyc_cases.vendor_applicant_id` (unique, nullable) so the
    webhook can resolve the vendor's applicant ID back to a case — reusing
    the existing `kyc_cases`/`kyc_submissions` tables rather than a parallel
    structure.
- `deployments/docker-compose.services.yml` runs `kyc` alongside the other
  services; user depends on it being healthy.

**Deliberately not done:**
- No mobile UI change for a "verification in progress" state — `GET
  /v1/kyc/status` already returns `pending`/`manual_review` correctly via the
  existing polling endpoint, so nothing broke, but there's no push
  notification on verdict yet (would reuse the existing notification
  service, which already consumes `user.events`).
- No integration test in `tests/integration` yet driving the simulator's
  admin API end-to-end (unit tests cover each magic-value branch and the
  manual-review resolution flow with fakes).
- Document upload (`POST /v1/applicants/{id}/documents`) was dropped from
  the original sketch — it would only ever be metadata with no consumer, so
  it stayed unbuilt rather than adding an endpoint nothing calls.

**Done when** (met): registration → KYC → provisioning works with a
genuinely async verdict; each magic-value branch (approve/reject/manual
review/underage) is unit-tested; a user stuck in `manual_review` can be
resolved via the admin API and completes provisioning; a redelivered verdict
webhook is a no-op, not a double wallet-provision.

## Phase 4 — FX / rates simulator (`services/simulators/fx`) — done

Enables multi-currency wallets and conversion. goledger already supports
per-account ISO 4217 currencies and multi-wallet-per-user already worked
before this phase (`UNIQUE (user_id, currency)`, `ProvisionWalletInput`
already parameterized by currency) — this phase was neobank-side work, not
ledger work, and needed no wallet-model changes at all.

One design constraint discovered while building this: **goledger rejects
cross-currency transfers outright** (`ErrCurrencyMismatch` in
`internal/usecase/transfer_usecase.go`). A conversion can never be a single
`CreateTransfer` call between a EUR account and a USD account — it has to be
two same-currency legs through per-currency FX position accounts. This
shaped the whole design below.

**Implemented and wired end-to-end:**

- `services/simulators/fx`: a standalone HTTP service with **no webhooks at
  all** — the only simulator in this plan that's purely synchronous
  request/response, since pricing and executing a quote don't need a
  vendor-originated callback the way a card auth or a bank transfer does.
  - `GET /v1/rates?from_currency=EUR&to_currency=USD` — the mid rate, a
    deterministic pseudo-random walk seeded per pair (`internal/usecase/rates.go`):
    the same (pair, 30-second bucket) always yields the same rate, but it
    drifts bucket to bucket so a rates chart looks alive. Each direction
    (EUR→USD vs USD→EUR) walks independently, so round trips are lossy like
    a real market, not perfectly invertible.
  - `POST /v1/quotes` — prices a conversion: mid rate widened by a 50bps
    retail spread, 30-second TTL, persisted in `fx.quotes` (every quote is
    kept whether or not it's ever executed — the audit trail a real vendor
    would show a regulator).
  - `POST /v1/quotes/{id}/execute` — locks in the quote. Idempotent:
    executing an already-executed quote returns the same result rather than
    erroring (the payment service may retry); executing an expired quote is
    refused.
  - Supports EUR/USD/GBP pairs (six seeded directions); adding a currency is
    a one-line map entry in `rates.go`.
- Payment service:
  - `POST /api/v1/payments/fx/quotes` and `.../quotes/{id}/execute` — plain
    net/http endpoints (same rationale as the rails/cardproc hand-rolled
    routes: no OpenAPI spec to keep in sync yet), mounted on the normal
    `Idempotency-Key`-protected router (no webhook involved, so no special
    router carve-out needed here unlike every other phase).
  - `ExecuteFXConversionUseCase`: calls the simulator's execute endpoint,
    resolves the caller's two currency wallets via the existing
    `userclient.GetWallet` (returns a clear error telling the user to open
    that currency wallet first if it doesn't exist yet — no
    auto-provisioning), then performs **two** same-currency ledger
    transfers: source wallet → source-currency FX position account, and
    destination-currency FX position account → destination wallet.
    Idempotent on the quote ID (`payment.fx_conversions.quote_id UNIQUE`) —
    re-executing the same quote is a no-op, not a second conversion.
  - New event `payment.fx_conversion.completed`, registered in `pkg/events`,
    projected into wallet tx history as **two** rows (debit in the source
    wallet, credit in the destination wallet) — both belong to the same
    user, so they need distinct IDs (`{conversion_id}-debit` /
    `{conversion_id}-credit`), unlike the existing `TransferCompleted`
    two-row case where the two rows belong to different users.
  - `FX_POSITION_ACCOUNT_EUR` / `_USD` / `_GBP` env vars map currency →
    ledger account, following the same manual-bootstrap pattern as
    `RAILS_SETTLEMENT_LEDGER_ACCOUNT_ID` (see
    `deployments/docker-compose.fx-override.yml`). A currency with no
    account configured simply can't be converted into or out of yet.
- `deployments/docker-compose.services.yml` runs `fx` alongside the other
  services; payment depends on it being healthy.

**Deliberately not done:**
- The spread isn't separated into its own fee-income ledger posting — it
  stays implicitly in the FX position accounts' growing balances (still
  economically captured, just not segregated into a dedicated line;
  segregating it would mean computing the spread's value in a common
  currency, which is real work, not a quick add).
- Reconciliation: the existing recon job doesn't yet check that FX position
  accounts net against `payment.fx_conversions` the way payment/card recon
  already checks transfers/authorizations against ledger state.
- No historical-rates endpoint for charting (`GET /v1/rates` only returns
  the current mid) and no gateway/mobile exposure (currency selector,
  conversion screen) — same category of gap as every other phase's
  gateway/mobile deferral.
- No integration test in `tests/integration` yet driving a real
  EUR→USD→ledger round trip end to end (13 unit tests cover the simulator's
  rate math, quote pricing, and execute idempotency/expiry with fakes).

**Done when** (met): a user converts EUR→USD at a quoted rate through two
real ledger transfers via FX position accounts; re-executing the same quote
ID is a no-op; executing an expired quote is refused; both wallet histories
show the conversion with the quote ID traceable via the ledger transfer IDs
recorded on `payment.fx_conversions`.
**Not yet met**: FX position accounts reconciling against executed quotes in
an automated recon job; fee income segregated from position-account balance.

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
