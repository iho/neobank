# neobank

Production-oriented neobank backend monorepo in Go, built around the existing [goledger](https://github.com/iho/goledger) double-entry ledger. Mobile clients talk to a single **API Gateway (BFF)**; domain services own their data and coordinate money movement exclusively through goledger.

## MVP status

| Capability | Status |
|------------|--------|
| User registration, login, JWT refresh | Done |
| KYC-lite (auto-approve) + wallet provisioning saga | Done |
| Wallet balance (ledger `GetAccount`) | Done |
| P2P transfers (saga: fraud → ledger → outbox) | Done |
| Virtual cards (issue, freeze, unfreeze) | Done |
| Card authorization + capture (hold → settle) | Done |
| Notifications (HTTP ingest + optional Kafka) | Done (incl. user.events) |
| API Gateway BFF with JWT auth | Done |
| Fraud rules (velocity, amount caps) | Done (`pkg/fraud`) |
| Kafka event bus | Optional (HTTP fallback) |

See [docs/architecture.md](docs/architecture.md) for the full system design and phased roadmap.

## Architecture

```mermaid
flowchart LR
    Mobile[Mobile App] -->|REST| GW[Gateway :8080]
    GW --> User[User :8081]
    GW --> Payment[Payment :8082]
    GW --> Card[Card :8084]
    GW --> Notif[Notification :8083]

    Payment -->|gRPC| Ledger[goledger :50051]
    Card -->|gRPC| Ledger
    User -->|gRPC| Ledger

    Payment --> Outbox1[Outbox]
    Card --> Outbox2[Outbox]
    Outbox1 -->|Kafka or HTTP| Notif
    Outbox2 -->|Kafka or HTTP| Notif

    User --> PG[(PostgreSQL)]
    Payment --> PG
    Card --> PG
    Notif --> PG
    GW --> Redis[(Redis)]
```

**Key principle:** only goledger mutates balances. Payment and Card services reference ledger account IDs; User provisions wallets by creating ledger accounts on KYC approval.

## Repository layout

```
neobank/
├── pkg/                    # Shared libraries (auth, fraud, idempotency, ledgerclient, outbox, saga, events)
├── proto/                  # Protobuf (goledger + neobank contracts) → pkg/gen/
├── services/
│   ├── gateway/            # BFF — public REST API (:8080)
│   ├── user/               # Auth, KYC, wallets (:8081)
│   ├── payment/            # P2P transfers (:8082)
│   ├── notification/       # Event ingest + inbox (:8083)
│   ├── card/               # Virtual cards + authorizations (:8084)
│   └── ledger/             # Pointer to external goledger (not vendored)
├── deployments/            # docker-compose (Postgres, Redis, Kafka)
├── docs/architecture.md    # Full design document
├── Makefile
└── go.work
```

Each service follows **clean architecture**: OpenAPI spec → oapi-codegen (strict Chi handlers) → use cases → sqlc repositories.

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- [goledger](https://github.com/iho/goledger) running locally (gRPC on `:50051`)
- Optional: `oapi-codegen`, `sqlc`, `buf` (or use `make generate`)

## Quick start

### 1. Infrastructure

```bash
make up
```

Starts PostgreSQL (`:5432`), Redis (`:6379`), Kafka (`:9092`), and an optional
OpenTelemetry collector (`:4317` gRPC, `:4318` HTTP).

To emit traces from services:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

### 2. Ledger (external)

```bash
git clone https://github.com/iho/goledger.git /tmp/goledger
cd /tmp/goledger
docker compose -f docker-compose.full.yml up -d
./scripts/setup-and-test.sh
```

Neobank connects via `LEDGER_GRPC_ADDR=localhost:50051`. See [services/ledger/README.md](services/ledger/README.md).

### 3. Generate, migrate, build

```bash
make tools      # install oapi-codegen (first time)
make generate   # proto + sqlc + oapi
make migrate-user migrate-payment migrate-notification migrate-card
make build
```

### 4. Run services

In separate terminals (or use a process manager):

```bash
./bin/user
./bin/payment
./bin/card
./bin/notification
./bin/gateway
```

Optional: set `KAFKA_BROKERS=localhost:9092` on User, Payment, Card, and Notification for Kafka-based event delivery instead of direct HTTP.

For card capture, create a settlement account in goledger and set:

```bash
export SETTLEMENT_LEDGER_ACCOUNT_ID=<ledger-account-uuid>
```

### 5. Smoke test

```bash
# Register
curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"phone":"+15551234567","email":"alice@example.com","password":"secret123","first_name":"Alice","last_name":"Smith"}'

# Login (save access_token from response)
curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"+15551234567","password":"secret123"}'

# Submit KYC (auto-approved in MVP) — provisions wallet
curl -s -X POST http://localhost:8080/v1/kyc \
  -H "Authorization: Bearer <access_token>" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"document_type":"passport","document_number":"AB123456"}'

# Wallet balance
curl -s http://localhost:8080/v1/wallet \
  -H "Authorization: Bearer <access_token>"
```

## Services & ports

| Service | Port | Responsibility |
|---------|------|----------------|
| Gateway (BFF) | 8080 | Public REST API, JWT validation, service aggregation |
| User | 8081 | Registration, login, KYC, wallet provisioning |
| Payment | 8082 | P2P transfers, transfer history |
| Notification | 8083 | Event ingest, notification inbox |
| Card | 8084 | Virtual cards, authorizations, capture |
| goledger (external) | 50051 | Double-entry ledger (accounts, transfers, holds) |

PostgreSQL uses **schema-per-service**: `user`, `payment`, `card`, `notification` (see [deployments/init-db.sql](deployments/init-db.sql)).

## Gateway API

Base URL: `http://localhost:8080`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/v1/auth/register` | — | Create account |
| `POST` | `/v1/auth/login` | — | Issue access + refresh tokens |
| `POST` | `/v1/auth/refresh` | — | Rotate tokens |
| `GET` | `/v1/me` | JWT | User profile + KYC status |
| `POST` | `/v1/kyc` | JWT | Submit KYC (auto-approve MVP) |
| `GET` | `/v1/kyc/status` | JWT | KYC status |
| `GET` | `/v1/wallet` | JWT | Wallet balance |
| `GET` | `/v1/wallet/transactions` | JWT | Unified transaction history (P2P + card) |
| `POST` | `/v1/wallets` | JWT | Provision wallet |
| `GET` | `/v1/transfers` | JWT | List transfers |
| `GET` | `/v1/transfers/{id}` | JWT | Transfer details |
| `POST` | `/v1/transfers` | JWT | P2P transfer |
| `GET` | `/v1/cards` | JWT | List cards |
| `POST` | `/v1/cards` | JWT | Issue virtual card |
| `GET` | `/v1/cards/{id}` | JWT | Card details |
| `POST` | `/v1/cards/{id}/freeze` | JWT | Freeze card |
| `POST` | `/v1/cards/{id}/unfreeze` | JWT | Unfreeze card |
| `POST` | `/v1/cards/{id}/authorize` | JWT | Authorize transaction (ledger hold) |
| `GET` | `/v1/authorizations` | JWT | List authorizations |
| `GET` | `/v1/authorizations/{id}` | JWT | Authorization details |
| `POST` | `/v1/authorizations/{id}/capture` | JWT | Capture hold → settlement |
| `GET` | `/v1/notifications` | JWT | Notification inbox |
| `GET` | `/health` | — | Health check |

OpenAPI spec: [services/gateway/api/openapi.yaml](services/gateway/api/openapi.yaml)

### Authentication (local dev)

- **JWT** — `Authorization: Bearer <access_token>` from login/register (15 min access, 7 day refresh).
- **Legacy dev token** — `Bearer access.<user-id>.<anything>` for quick testing without login. Rejected unless `APP_ENV` is `development`/`local`/`dev` (see `AllowDevAuth` in `services/gateway/internal/config`).
- **`X-User-Id` header** — bypasses JWT parsing when set. Same `APP_ENV` gate as above.

Mutating endpoints accept `Idempotency-Key` (Redis-backed; in-memory fallback if Redis is unavailable).

### Traceability

Every request gets a correlation ID (`X-Correlation-Id`, generated by the gateway if the caller
doesn't send one) that propagates through HTTP calls, gRPC calls to goledger, and outbox events,
so a support ticket or regulator request can be traced end-to-end. Status-changing mutations
(transfers, cards, authorizations, KYC, wallets) write an append-only row to that service's
`audit_log` table in the same transaction as the change, and every fraud evaluation — allow or
deny — is recorded in `fraud_decisions`. Reconciliation jobs compare local state against goledger
and persist run summaries to `reconciliation_runs` plus trackable rows in `reconciliation_breaks`.

### Reconciliation runbook

Reconciliation verifies that Payment transfers and Card authorizations still match goledger
(transfers and holds). Runs are intended on a schedule, not as long-lived services.

**Manual run (local dev)**

```bash
make reconcile-payment   # exits 1 if breaks found
make reconcile-card
```

**Scheduled runs (docker)**

Requires Postgres (`make up`), goledger on `:50051`, and the user service on `:8081` (card job only):

```bash
make up-jobs    # hourly cron: payment :00, card :15 (UTC)
make down-jobs
```

**When a run reports breaks**

1. Inspect the latest run: `SELECT * FROM payment.reconciliation_runs ORDER BY started_at DESC LIMIT 1;`
2. List open breaks: `make list-payment-breaks` or `make list-card-breaks`
3. Investigate the entity (transfer/authorization), ledger record, and saga state
4. Mark progress:
   ```bash
   cd services/payment && go run ./cmd/resolve-break -id <uuid> -status investigated -by ops@example.com -notes "checking ledger"
   cd services/payment && go run ./cmd/resolve-break -id <uuid> -status closed -by ops@example.com -notes "manual ledger adjustment"
   ```
   (Same flags for `services/card/cmd/resolve-break`.)

Break statuses: `open` → `investigated` → `closed`. A break that reappears after closure
creates a new row; active duplicates on the same entity+reason update the latest run reference.

### Saga watchdog runbook

The saga watchdog scans `saga_instances` in the `user`, `payment`, and `card` schemas for
workflows stuck in `running` or `compensating` without progress (default: 15 minutes). It
upserts rows in `saga_alerts` so operators can investigate without polling raw saga tables.

**Manual run (local dev)**

```bash
make saga-watchdog        # exits 1 if open alerts remain
make list-saga-alerts     # list open/investigating alerts across all schemas
```

**Scheduled runs (docker)**

The jobs container runs the watchdog every 15 minutes alongside reconciliation:

```bash
make up-jobs    # cron: reconcile hourly, saga-watchdog every 15m (UTC)
make down-jobs
```

**When alerts are open**

1. List alerts: `make list-saga-alerts` or query `SELECT * FROM payment.saga_alerts WHERE alert_status IN ('open','investigating');`
2. Inspect the saga instance (`saga_instances`), related domain rows (transfer, authorization, wallet), and outbox events
3. Mark progress:
   ```bash
   cd tools/saga-watchdog && go run . -schema payment -resolve-id <uuid> -resolve-status investigating -by ops@example.com -notes "checking ledger"
   cd tools/saga-watchdog && go run . -schema payment -resolve-id <uuid> -resolve-status resolved -by ops@example.com -notes "manual compensation completed"
   ```
4. Alerts auto-resolve when the saga reaches `completed` or `failed` on a later scan

Alert statuses: `open` → `investigating` → `resolved`. Stuck sagas that remain non-terminal
keep refreshing `last_seen_at` on each scan.

### Wallet transaction history (CQRS read model)

`GET /v1/wallet/transactions` reads from `user.wallet_transactions`, a projection maintained
by the User service from payment/card outbox events (`payment.transfer.completed`,
`card.auth.approved`, `card.auth.captured`). Payment and Card outbox workers fan out to both
Notification ingest and User ingest (HTTP in dev, Kafka consumer in production). The gateway
no longer merges Payment + Card responses in memory.

## Environment variables

Shared defaults work for local development. Override as needed:

| Variable | Default | Used by |
|----------|---------|---------|
| `DATABASE_URL` | `postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable` | user, payment, card, notification |
| `REDIS_URL` | `redis://localhost:6379/0` | gateway, user, payment, card |
| `JWT_SECRET` | `dev-secret-change-me` | gateway, user |
| `LEDGER_GRPC_ADDR` | `localhost:50051` | gateway, user, payment, card |
| `USER_SERVICE_URL` | `http://localhost:8081` | gateway, payment, card |
| `PAYMENT_SERVICE_URL` | `http://localhost:8082` | gateway |
| `CARD_SERVICE_URL` | `http://localhost:8084` | gateway |
| `NOTIFICATION_SERVICE_URL` | `http://localhost:8083` | gateway |
| `KAFKA_BROKERS` | _(empty)_ | user, payment, card, notification |
| `SETTLEMENT_LEDGER_ACCOUNT_ID` | _(empty)_ | card (required for capture) |
| `APP_ENV` | `development` | gateway (gates dev auth bypass; set to `production`/`staging` to disable) |

## Make targets

```bash
make deps          # go mod tidy all modules
make generate      # proto + sqlc + oapi-codegen
make build         # compile all services to bin/
make test          # run unit tests
make lint          # golangci-lint
make up / make down   # docker-compose infra
make migrate-*      # run DB migrations per service
make reconcile-payment   # compare payment.transfers against goledger
make reconcile-card      # compare card.authorizations against goledger holds
make up-jobs / down-jobs # scheduled reconciliation (docker cron)
make list-payment-breaks / list-card-breaks
make saga-watchdog / list-saga-alerts   # stuck saga detection
```

## Shared packages (`pkg/`)

| Package | Purpose |
|---------|---------|
| `auth` | HS256 JWT issue/validate |
| `fraud` | Pre-auth risk checks (velocity, limits) |
| `idempotency` | Redis-backed idempotency middleware |
| `ledgerclient` | gRPC client for goledger |
| `outbox` | Transactional outbox worker (Kafka → HTTP → log) |
| `saga` | Multi-step orchestration with persisted state |
| `events` | Domain event envelopes (payment, card, user) |
| `userclient` | HTTP client for User service internals |
| `money` | Decimal helpers |
| `otel` | OpenTelemetry bootstrap |
| `reqctx` | Correlation/causation ID propagation (HTTP + gRPC) |
| `audit` | Shared `audit_log` entry shape, written in-tx alongside domain mutations |
| `sagawatchdog` | Stuck-saga scanner; upserts `saga_alerts` and auto-resolves on terminal state |
| `walletprojection` | Maps payment/card outbox events into unified `wallet_transactions` read model |
| `screening` | Sanctions/PEP screening stub with auditable `screening_checks` persistence |
| `sloghttp` | Structured HTTP access logs (`correlation_id`, `user_id`, `idempotency_key`) |

## Patterns

- **Saga** — P2P transfer and card issuance/auth run multi-step flows with compensating logic; state stored in PostgreSQL.
- **Outbox** — Events written in the same DB transaction as domain changes, then published asynchronously, tagged with the originating correlation ID.
- **Idempotency** — All mutating gateway and service endpoints honor `Idempotency-Key`.
- **Fraud** — Synchronous checks before ledger mutations (`pkg/fraud`); every decision (allow or deny) is persisted to `fraud_decisions`.
- **Audit trail** — Status-changing mutations append to `audit_log` in the same transaction as the change, so history survives destructive `UPDATE`s.

## Testing

```bash
make test
```

Unit tests cover JWT, idempotency middleware, fraud rules, and gateway auth resolution. Integration tests
(testcontainers Postgres + Redis, in-memory mock ledger) live under `tests/integration/`:

```bash
make test-integration   # requires Docker
```

## Documentation

- [docs/architecture.md](docs/architecture.md) — full system design, schemas, event catalog, roadmap
- [plan.md](plan.md) — original product/architecture brief
- [services/ledger/README.md](services/ledger/README.md) — goledger integration

## Roadmap (deferred)

- Standalone Fraud service (currently embedded in `pkg/fraud`)
- Real KYC provider (AML/sanctions screening), card processor, SMS/email/push providers
- gRPC between Gateway and services (currently HTTP)
- ~~CQRS read model for transaction history~~ — done: `user.wallet_transactions` projected from payment/card outbox events; gateway reads via User service
- Kubernetes manifests, OpenTelemetry collector wiring (trace propagation is stubbed via `pkg/reqctx`; not yet wired to a collector)
- Contract and integration test suites
- Outbox retention/archival policy (WORM/object-lock) and PII field-level encryption — see `todo.md`

## License

See [LICENSE](LICENSE).