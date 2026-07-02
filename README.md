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
| Unified wallet tx history (CQRS projection) | Done |
| Notifications (HTTP ingest + optional Kafka) | Done |
| API Gateway BFF with JWT auth | Done |
| Fraud rules + persisted decisions | Done (`pkg/fraud`) |
| Sanctions/PEP screening stub + AML monitoring | Done (`pkg/screening`, `pkg/amlmonitor`) |
| Correlation ID + audit trail + PII read audit | Done |
| GDPR export / PII masking (internal) | Done |
| PII field encryption (Vault Transit, optional) | Done (`pkg/vault`, `pkg/piicrypto`) |
| Reconciliation + break resolution | Done |
| Saga watchdog + alerts | Done |
| golang-migrate per service | Done (`pkg/migrate`) |
| Kafka event bus | Optional (HTTP fan-out fallback) |

See [docs/architecture.md](docs/architecture.md) for the full system design and [todo.md](todo.md) for compliance backlog (WORM archival, production Vault HA, real KYC vendor).

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

    Payment --> OutboxP[Outbox]
    Card --> OutboxC[Outbox]
    User --> OutboxU[Outbox]

    OutboxP -->|events| Notif
    OutboxC -->|events| Notif
    OutboxP -->|events| Proj[User projection]
    OutboxC -->|events| Proj

    User -.->|optional encrypt| Vault[Vault :8200]

    User --> PG[(PostgreSQL)]
    Payment --> PG
    Card --> PG
    Notif --> PG
    GW --> Redis[(Redis)]
```

**Key principles**

- Only **goledger** mutates balances. Payment and Card reference ledger account IDs; User provisions wallets by creating ledger accounts on KYC approval.
- **Orchestration** (`pkg/saga`) drives money-moving flows in the request path (P2P, card auth, wallet provision).
- **Choreography** (outbox → Kafka/HTTP) drives side effects: notifications and the wallet-transaction read model. No central coordinator for those consumers.

## Orchestration vs choreography (in this repo)

| Pattern | Used for | Examples |
|---------|----------|----------|
| **Orchestration** | Multi-step flows that must succeed or compensate | `p2p_transfer`, `card_authorization`, `wallet_provision` — state in `saga_instances` |
| **Choreography** | Reacting to facts after the fact | Notification inbox, `user.wallet_transactions` projection from `payment.events` / `card.events` |

## Repository layout

```
neobank/
├── pkg/                    # Shared libraries (saga, outbox, audit, vault, fraud, …)
├── proto/                  # Protobuf → pkg/gen/
├── services/
│   ├── gateway/            # BFF — public REST API (:8080)
│   ├── user/               # Auth, KYC, wallets, projections (:8081)
│   ├── payment/            # P2P transfers, AML (:8082)
│   ├── notification/       # Event ingest + inbox (:8083)
│   └── card/               # Virtual cards + authorizations (:8084)
├── tools/
│   ├── saga-watchdog/      # Stuck-saga scanner
│   └── event-catalog/      # Published event contract export
├── tests/integration/      # Testcontainers suite
├── deployments/            # docker-compose, Vault init, cron jobs
├── docs/architecture.md
├── Makefile
└── go.work
```

Each service follows **clean architecture**: OpenAPI → oapi-codegen (strict Chi) → use cases → sqlc repositories.

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- [goledger](https://github.com/iho/goledger) on `:50051` (gRPC)
- Optional: `oapi-codegen`, `sqlc`, `buf` (or `make generate`)

## Quick start

### 1. Infrastructure

```bash
make up
```

Starts PostgreSQL (`:5432`), Redis (`:6379`), Kafka (`:9092`), Vault dev server (`:8200`), and an OpenTelemetry collector (`:4317` gRPC).

Optional PII encryption at rest:

```bash
make vault-init   # enables Transit keys: pii, pii-phone
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-root-token
```

Without `VAULT_ADDR`, the user service stores phone/DOB/document numbers in plaintext (fine for tests).

Tracing:

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

Neobank uses `LEDGER_GRPC_ADDR=localhost:50051`. See [services/ledger/README.md](services/ledger/README.md).

### 3. Generate, migrate, build

```bash
make tools      # install oapi-codegen (first time)
make generate   # proto + sqlc + oapi
make migrate    # all services (golang-migrate)
make build
```

Migrations live in `services/*/migrations/` as `00000N_name.up.sql` / `.down.sql`. Each service tracks versions in its own schema (`user.schema_migrations`, etc.) via [golang-migrate](https://github.com/golang-migrate/migrate) and `pkg/migrate`.

### 4. Run services

```bash
./bin/user
./bin/payment
./bin/card
./bin/notification
./bin/gateway
```

Optional Kafka (instead of HTTP outbox fan-out):

```bash
export KAFKA_BROKERS=localhost:9092
```

Card capture requires a goledger settlement account:

```bash
export SETTLEMENT_LEDGER_ACCOUNT_ID=<ledger-account-uuid>
```

### 5. Smoke test

```bash
# Register
curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"email":"alice@example.com","phone":"+15551234567","password":"secret123"}'

# Login
curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'

# Submit KYC (auto-approved in MVP) — provisions wallet
curl -s -X POST http://localhost:8080/v1/kyc \
  -H "Authorization: Bearer <access_token>" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"full_name":"Alice Smith","date_of_birth":"1990-01-15","country_code":"US"}'

# Wallet balance
curl -s http://localhost:8080/v1/wallet \
  -H "Authorization: Bearer <access_token>"
```

## Services & ports

| Service | Port | Responsibility |
|---------|------|----------------|
| Gateway (BFF) | 8080 | Public REST API, JWT validation |
| User | 8081 | Auth, KYC, wallets, wallet-tx projection, GDPR (internal) |
| Payment | 8082 | P2P transfers, AML export |
| Notification | 8083 | Event ingest → notification inbox |
| Card | 8084 | Virtual cards, authorizations, capture |
| Vault (local dev) | 8200 | Optional Transit encryption for PII |
| goledger (external) | 50051 | Double-entry ledger |

PostgreSQL uses **schema-per-service**: `user`, `payment`, `card`, `notification` ([deployments/init-db.sql](deployments/init-db.sql)).

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
| `GET` | `/v1/wallet/transactions` | JWT | Unified transaction history |
| `POST` | `/v1/wallets` | JWT | Provision wallet |
| `GET` | `/v1/transfers` | JWT | List transfers |
| `GET` | `/v1/transfers/{id}` | JWT | Transfer details |
| `POST` | `/v1/transfers` | JWT | P2P transfer |
| `GET` | `/v1/cards` | JWT | List cards |
| `POST` | `/v1/cards` | JWT | Issue virtual card |
| `GET` | `/v1/cards/{id}` | JWT | Card details |
| `POST` | `/v1/cards/{id}/freeze` | JWT | Freeze card |
| `POST` | `/v1/cards/{id}/unfreeze` | JWT | Unfreeze card |
| `POST` | `/v1/cards/{id}/authorize` | JWT | Authorize (ledger hold) |
| `GET` | `/v1/authorizations` | JWT | List authorizations |
| `GET` | `/v1/authorizations/{id}` | JWT | Authorization details |
| `POST` | `/v1/authorizations/{id}/capture` | JWT | Capture hold → settlement |
| `GET` | `/v1/notifications` | JWT | Notification inbox |
| `GET` | `/health` | — | Health check |

OpenAPI: [services/gateway/api/openapi.yaml](services/gateway/api/openapi.yaml)

### Internal APIs (service-to-service)

Not exposed on the gateway; used by Payment/Card outbox fan-out or ops:

| Service | Path | Purpose |
|---------|------|---------|
| User | `POST /api/v1/internal/events` | Wallet projection ingest |
| User | `POST /api/v1/internal/users/{id}/gdpr/export` | GDPR data bundle |
| User | `POST /api/v1/internal/users/{id}/gdpr/mask` | PII masking (retain financial rows) |
| User | `GET /api/v1/internal/users/by-phone/{phone}` | P2P recipient lookup |
| Notification | `POST /api/v1/internal/events` | Notification ingest |

### Authentication (local dev)

- **JWT** — `Authorization: Bearer <access_token>` (15 min access, 7 day refresh).
- **Legacy dev token** — `Bearer access.<user-id>.*` — only when `APP_ENV` is `development`/`local`/`dev`.
- **`X-User-Id` header** — same `APP_ENV` gate as above.

**Production:** set `APP_ENV=production` (or `staging`) on the gateway to disable dev-auth bypasses.

Mutating endpoints require `Idempotency-Key` (Redis-backed; in-memory fallback if Redis is down).

### Traceability & compliance

- **Correlation ID** — `X-Correlation-Id` from gateway through HTTP, gRPC (goledger), and outbox events (`pkg/reqctx`).
- **Write audit** — `audit_log` per schema, appended in the same transaction as status mutations; DB triggers block UPDATE/DELETE on audit/evidence tables.
- **Fraud / screening / AML** — every evaluation persisted (`fraud_decisions`, `screening_checks`, `aml_evaluations` / `aml_cases`).
- **PII read audit** — `user.pii_access_log` records successful reads of profile, KYC, wallet, and internal lookups.
- **GDPR** — export bundles customer data; mask overwrites PII in place (financial records retained); both logged in `gdpr_requests`.
- **PII encryption** — when Vault is configured, phone, DOB, and document numbers are Transit-encrypted; phone lookup uses an HMAC blind index (`phone_lookup`).
- **Reconciliation** — `reconciliation_runs` + `reconciliation_breaks` vs goledger; resolve via `cmd/resolve-break`.
- **Saga watchdog** — stuck `saga_instances` → `saga_alerts` for operator follow-up.

## Operational runbooks

### Reconciliation

```bash
make reconcile-payment   # exits 1 if breaks found
make reconcile-card
make list-payment-breaks
make list-card-breaks
```

Resolve a break (`open` → `investigated` → `closed`):

```bash
cd services/payment && go run ./cmd/resolve-break \
  -id <uuid> -status investigated -by ops@example.com -notes "checking ledger"
```

Scheduled (hourly cron UTC): `make up-jobs`

### Saga watchdog

```bash
make saga-watchdog
make list-saga-alerts
```

Resolves alerts when the saga later reaches `completed` or `failed`. Does not auto-resume stuck sagas — operators investigate and may retry the client request (same `Idempotency-Key`) so the orchestrator skips completed steps.

### Wallet transaction history (CQRS)

`GET /v1/wallet/transactions` reads `user.wallet_transactions`, projected from payment/card outbox events (`payment.transfer.completed`, `card.auth.approved`, `card.auth.captured`). Payment/Card outbox workers fan out to Notification and User ingest (HTTP in dev, Kafka in production).

### AML export

```bash
make aml-export
```

### Event catalog

```bash
make event-catalog
```

## Environment variables

| Variable | Default | Used by |
|----------|---------|---------|
| `DATABASE_URL` | `postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable` | all services |
| `REDIS_URL` | `redis://localhost:6379/0` | gateway, user, payment, card |
| `JWT_SECRET` | `dev-secret-change-me` | gateway, user |
| `LEDGER_GRPC_ADDR` | `localhost:50051` | user, payment, card |
| `USER_SERVICE_URL` | `http://localhost:8081` | gateway, payment, card |
| `PAYMENT_SERVICE_URL` | `http://localhost:8082` | gateway |
| `CARD_SERVICE_URL` | `http://localhost:8084` | gateway |
| `NOTIFICATION_SERVICE_URL` | `http://localhost:8083` | gateway, outbox |
| `KAFKA_BROKERS` | _(empty)_ | user, payment, card, notification |
| `SETTLEMENT_LEDGER_ACCOUNT_ID` | _(empty)_ | card (capture) |
| `APP_ENV` | `development` | gateway (dev-auth gate) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | _(empty)_ | all services (tracing off if unset) |
| `VAULT_ADDR` | _(empty)_ | user (noop encryption if unset) |
| `VAULT_TOKEN` | `dev-root-token` in dev | user |
| `VAULT_TRANSIT_KEY` | `pii` | user |
| `VAULT_HMAC_KEY` | `pii-phone` | user |

## Make targets

```bash
make deps              # go mod tidy all modules
make generate          # proto + sqlc + oapi-codegen
make build             # compile services and ops binaries to bin/
make test              # unit tests (pkg/)
make test-integration  # testcontainers (requires Docker)
make lint              # golangci-lint

make up / make down    # docker-compose infra
make migrate           # golang-migrate all services
make migrate-user      # … payment, notification, card
make vault-init        # local Vault Transit keys

make reconcile-payment / reconcile-card
make list-payment-breaks / list-card-breaks
make saga-watchdog / list-saga-alerts
make aml-export
make event-catalog

make up-jobs / down-jobs   # cron: reconcile + saga-watchdog
```

## Shared packages (`pkg/`)

| Package | Purpose |
|---------|---------|
| `saga` | In-process orchestrator; persisted `completed_steps` |
| `outbox` | Transactional outbox → Kafka or HTTP fan-out |
| `events` / `walletprojection` | Event envelopes and CQRS projection rules |
| `audit` | Write audit + PII read audit entries |
| `migrate` | golang-migrate wrapper (per-schema version table) |
| `vault` / `piicrypto` | HashiCorp Vault Transit encrypt + phone HMAC index |
| `gdpr` | Masked email helpers |
| `fraud` / `screening` / `amlmonitor` | Risk, sanctions stub, txn monitoring |
| `ledgerclient` | gRPC client for goledger |
| `idempotency` | Redis-backed `Idempotency-Key` middleware |
| `reqctx` / `otel` / `sloghttp` | Correlation IDs, tracing, access logs |
| `sagawatchdog` | Stuck-saga scanner and `saga_alerts` |
| `auth` / `userclient` / `money` | JWT, internal user HTTP client, decimals |

## Patterns

- **Saga (orchestration)** — P2P, card auth, wallet provision: multi-step, compensating, state in `saga_instances`.
- **Outbox + choreography** — After domain commit, events notify and project without a central workflow engine.
- **Idempotency** — HTTP `Idempotency-Key` + domain unique constraints + ledger idempotency keys on transfers.
- **Append-only evidence** — `audit_log`, `fraud_decisions`, outbox payloads: DB triggers reject UPDATE/DELETE.

## Testing

```bash
make test
make test-integration   # ~1–2 min, requires Docker
```

Integration tests cover P2P, card auth/capture, wallet projection dedup, notification dedup, saga watchdog, outbox/audit immutability, GDPR, PII read audit, and AML.

## Documentation

- [docs/architecture.md](docs/architecture.md) — system design, schemas, events
- [todo.md](todo.md) — regulatory backlog and status
- [services/ledger/README.md](services/ledger/README.md) — goledger integration

## Roadmap (deferred)

- Standalone Fraud service (today: `pkg/fraud` in Payment/Card)
- Real KYC/AML vendors (today: stubs with persisted evidence)
- gRPC between Gateway and services (today: HTTP)
- Kubernetes manifests and production Vault (HA, AppRole, auto-unseal)
- Outbox partition + WORM archival to object storage
- Background saga recovery worker (today: client retry + watchdog alerts)

## License

See [LICENSE](LICENSE).