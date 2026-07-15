# neobank

Production-oriented neobank backend monorepo in Go, built around the existing [goledger](https://github.com/iho/goledger) double-entry ledger. Mobile clients talk to a single **API Gateway (BFF)**; domain services own their data and coordinate money movement exclusively through goledger.

Designed for **auditability**: correlation IDs end-to-end, append-only evidence tables, PII read logging, GDPR tooling, optional Vault Transit encryption, and reconciliation against the ledger.

## Contents

- [MVP status](#mvp-status)
- [Architecture](#architecture)
- [Orchestration vs choreography](#orchestration-vs-choreography-in-this-repo)
- [Repository layout](#repository-layout)
- [Mobile clients](#mobile-clients)
- [Quick start](#quick-start)
- [Services & ports](#services--ports)
- [Gateway API](#gateway-api)
- [Operational runbooks](#operational-runbooks)
- [Domain events](#domain-events)
- [Environment variables](#environment-variables)
- [Make targets](#make-targets)
- [Shared packages](#shared-packages-pkg)
- [Testing & CI](#testing--ci)
- [Troubleshooting](#troubleshooting)
- [Documentation](#documentation)
- [Roadmap](#roadmap-deferred)

## MVP status

| Capability | Status |
|------------|--------|
| User registration, login, JWT refresh | Done |
| KYC + wallet provisioning saga | Done — async vendor verdict via `services/simulators/kyc` (approve/reject/manual-review), not auto-approve; see [docs/vendor-simulators-plan.md](docs/vendor-simulators-plan.md) |
| Wallet balance (ledger `GetAccount`) | Done |
| P2P transfers (saga: fraud → ledger → outbox) | Done |
| Bank transfer top-up + outbound send (rails simulator: virtual IBAN, inbound/outbound webhooks, ledger credit/debit, recon) | Done (`services/simulators/rails`; see [docs/vendor-simulators-plan.md](docs/vendor-simulators-plan.md)) |
| Virtual cards (issue, freeze, unfreeze) | Done |
| Card authorization + capture (hold → settle), auth expiry sweep, chargebacks (provisional credit → won/lost) | Done, driven by cardproc simulator webhooks (`services/simulators/cardproc`); see [docs/vendor-simulators-plan.md](docs/vendor-simulators-plan.md) |
| Multi-currency wallets + FX conversion (fx simulator: quote → execute → two-leg ledger posting) | Done (`services/simulators/fx`; fee-income separation and recon not yet wired — see [docs/vendor-simulators-plan.md](docs/vendor-simulators-plan.md)) |
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

See [docs/architecture.md](docs/architecture.md) for design rationale and schema detail (some diagrams describe target architecture; **this README reflects what is implemented today**). See [todo.md](todo.md) for the compliance backlog (WORM archival, production Vault HA, real KYC vendor).

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

### P2P transfer flow (orchestration)

```mermaid
sequenceDiagram
    participant App as Mobile
    participant GW as Gateway
    participant Pay as Payment
    participant Fraud as pkg/fraud
    participant Led as goledger
    participant OB as Outbox
    participant User as User projection
    participant Notif as Notification

    App->>GW: POST /v1/transfers
    GW->>Pay: create transfer (saga)
    Pay->>Fraud: evaluate sender/recipient
    Pay->>Led: ledger transfer (idempotent)
    Pay->>OB: payment.transfer.completed
    OB-->>User: HTTP/Kafka ingest
    OB-->>Notif: HTTP/Kafka ingest
    Pay-->>GW: transfer completed
    GW-->>App: 200 + transfer id
```

Saga state (`saga_instances.completed_steps`) survives process restarts. If a step fails, earlier steps are compensated. Stuck sagas surface as `saga_alerts` via the watchdog — they are **not** auto-resumed; retry the client request with the same `Idempotency-Key`.

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

PostgreSQL uses **schema-per-service** (`user`, `payment`, `card`, `notification`) — see [deployments/init-db.sql](deployments/init-db.sql).

## Mobile clients

Three mobile clients talk to the same gateway BFF, as a comparison point for what each
toolkit looks like on identical backend contracts:

- [`mobile/`](mobile) — Flutter (iOS + Android)
- [`ios-native/`](ios-native) — native SwiftUI
- [`android-native/`](android-native) — native Jetpack Compose

<table>
  <tr>
    <td align="center"><img src="android-native/Screenshots/login.png" width="180"><br><sub>Login</sub></td>
    <td align="center"><img src="android-native/Screenshots/kyc.png" width="180"><br><sub>KYC gate</sub></td>
    <td align="center"><img src="android-native/Screenshots/wallet.png" width="180"><br><sub>Wallet</sub></td>
    <td align="center"><img src="android-native/Screenshots/cards.png" width="180"><br><sub>Cards</sub></td>
    <td align="center"><img src="android-native/Screenshots/alerts.png" width="180"><br><sub>Alerts</sub></td>
  </tr>
</table>

(Android client shown; see each client's own README for its full screenshot set —
[`android-native/README.md`](android-native/README.md#screenshots),
[`ios-native/README.md`](ios-native/README.md#screenshots).)

## Prerequisites

- Go 1.24+ (see [go.work](go.work))
- Docker & Docker Compose
- [goledger](https://github.com/iho/goledger) on `:50051` (gRPC) — required for wallet provisioning and money movement
- Optional: `oapi-codegen`, `sqlc`, `buf` (or `make generate`)

## Quick start

Run these in order on a fresh machine:

| Step | Command | Notes |
|------|---------|-------|
| 1 | `make up` | Postgres, Redis, Redpanda, Vault dev, OTel collector |
| 2 | Start goledger | See [services/ledger/README.md](services/ledger/README.md) |
| 3 | `make tools && make generate && make migrate && make build` | First-time codegen + migrations |
| 4 | Run five binaries | See [Run services](#run-services) |
| 5 | Smoke test | See [Smoke test](#smoke-test) |

### 1. Infrastructure

```bash
make up
```

Starts PostgreSQL (`:5432`), Redis (`:6379`), Redpanda (`:19092` Kafka API from host, `redpanda:9092` in compose), Vault dev server (`:8200`), and an OpenTelemetry collector (`:4317` gRPC / `:4318` HTTP).

Optional PII encryption at rest:

```bash
make vault-init   # enables Transit keys: pii, pii-phone
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-root-token
```

Without `VAULT_ADDR`, the user service stores phone/DOB/document numbers in plaintext (fine for local dev and integration tests).

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

**Internal gRPC mTLS** (optional; off by default for local dev):

```bash
make grpc-mtls-certs   # then export:
export GRPC_MTLS_ENABLED=true
export GRPC_TLS_CA_FILE=$PWD/deployments/grpc-mtls/ca.crt
export GRPC_TLS_SERVER_CERT_FILE=$PWD/deployments/grpc-mtls/server.crt
export GRPC_TLS_SERVER_KEY_FILE=$PWD/deployments/grpc-mtls/server.key
export GRPC_TLS_CLIENT_CERT_FILE=$PWD/deployments/grpc-mtls/client.crt
export GRPC_TLS_CLIENT_KEY_FILE=$PWD/deployments/grpc-mtls/client.key
export GRPC_TLS_SERVER_NAME=localhost
```

Servers require client certificates; clients verify the server against the CA. goledger stays on plaintext via `grpcutil.DialInsecure` until it supports mTLS.

### 3. Generate, migrate, build

```bash
make tools      # install oapi-codegen (first time)
make generate   # proto + sqlc + oapi
make migrate    # all services (golang-migrate)
make build
```

Migrations live in `services/*/migrations/` as `00000N_name.up.sql` / `.down.sql`. Each service tracks versions in its own schema (`user.schema_migrations`, etc.) via [golang-migrate](https://github.com/golang-migrate/migrate) and `pkg/migrate`.

### 4. Run services

Start each binary in its own terminal (or use a process manager):

```bash
./bin/user
./bin/payment
./bin/card
./bin/notification
./bin/gateway
```

Optional Redpanda event bus (Kafka API; instead of HTTP outbox fan-out):

```bash
# Host binaries: Redpanda external listener
export KAFKA_BROKERS=localhost:19092
# In compose (make up-all): KAFKA_BROKERS=redpanda:9092 is set automatically
```

Card capture requires a goledger settlement account:

```bash
export SETTLEMENT_LEDGER_ACCOUNT_ID=<ledger-account-uuid>
```

Verify the stack:

```bash
curl -s http://localhost:8080/health
```

### 5. Smoke test

```bash
# Register
curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"email":"alice@example.com","phone":"+15551234567","password":"secret123"}'

# Login — capture access token (requires jq)
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"email":"alice@example.com","password":"secret123"}' | jq -r .access_token)

# Submit KYC (auto-approved in MVP) — provisions wallet via saga
curl -s -X POST http://localhost:8080/v1/kyc \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"full_name":"Alice Smith","date_of_birth":"1990-01-15","country_code":"US"}'

# Wallet balance
curl -s http://localhost:8080/v1/wallet \
  -H "Authorization: Bearer $TOKEN" | jq .

# Unified transaction history (empty until transfers/cards)
curl -s "http://localhost:8080/v1/wallet/transactions" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

P2P transfers need a funded sender ledger account and a second registered user; see [tests/integration/p2p_transfer_test.go](tests/integration/p2p_transfer_test.go) for the full pattern.

## Services & ports

| Service | Port | Responsibility |
|---------|------|----------------|
| Gateway (BFF) | 8080 | Public REST API, JWT validation |
| User | 8081 | Auth, KYC, wallets, wallet-tx projection, GDPR (internal) |
| Payment | 8082 | P2P transfers, AML export |
| Notification | 8083 | Event ingest → notification inbox |
| Card | 8084 | Virtual cards, authorizations, capture |
| Rails simulator | 8090 | Simulated payment rail: virtual IBANs, inbound transfer webhooks (see [docs/vendor-simulators-plan.md](docs/vendor-simulators-plan.md)) |
| Cardproc simulator | 8091 | Simulated card network: issues virtual cards, drives real-time auth + capture/reversal webhooks into Card |
| KYC simulator | 8092 | Simulated identity vendor: async approve/reject/manual-review verdicts into User |
| FX simulator | 8093 | Simulated rates vendor: quote → execute, no webhooks (synchronous only) |
| Vault (local dev) | 8200 | Optional Transit encryption for PII |
| goledger (external) | 50051 | Double-entry ledger |

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
make event-catalog   # prints JSON contract from pkg/events/catalog.go
```

## Domain events

Published via the transactional outbox. Envelope fields include `event_id`, `event_type`, `event_version`, `occurred_at`, `correlation_id`, and `causation_id`.

| Event type | Topic | Producer | Consumers |
|------------|-------|----------|-----------|
| `user.registered` | `user.events` | User | Notification |
| `user.kyc.approved` | `user.events` | User | Notification |
| `user.wallet.provisioned` | `user.events` | User | Notification |
| `payment.transfer.completed` | `payment.events` | Payment | Notification, User projection |
| `card.issued` | `card.events` | Card | Notification |
| `card.frozen` / `card.unfrozen` | `card.events` | Card | Notification |
| `card.auth.approved` | `card.events` | Card | Notification, User projection |
| `card.auth.captured` | `card.events` | Card | Notification, User projection |

Contract source: [pkg/events/catalog.go](pkg/events/catalog.go). Export with `make event-catalog`.

## Environment variables

| Variable | Default | Used by |
|----------|---------|---------|
| `DATABASE_URL` | `postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable` | all services |
| `REDIS_URL` | `redis://localhost:6379/0` | gateway, user, payment, card |
| `JWT_SECRET` | `dev-secret-change-me` | gateway, user |
| `LEDGER_GRPC_ADDR` | `localhost:50051` | user, payment, card |
| `USER_GRPC_ADDR` | `localhost:50052` | gateway, payment, card, notification |
| `PAYMENT_GRPC_ADDR` | `localhost:50053` | gateway |
| `CARD_GRPC_ADDR` | `localhost:50054` | gateway |
| `NOTIFICATION_GRPC_ADDR` | `localhost:50055` | gateway |
| `USER_SERVICE_URL` | `http://localhost:8081` | payment/card wallet projection ingest |
| `GRPC_PORT` | `50052`–`50055` | user/payment/card/notification gRPC listeners |
| `GRPC_MTLS_ENABLED` | `false` | all gRPC servers/clients (set `true` for mTLS) |
| `GRPC_TLS_CA_FILE` | _(empty)_ | CA bundle to verify peer certs |
| `GRPC_TLS_SERVER_CERT_FILE` | _(empty)_ | server listener certificate (PEM) |
| `GRPC_TLS_SERVER_KEY_FILE` | _(empty)_ | server listener private key (PEM) |
| `GRPC_TLS_CLIENT_CERT_FILE` | _(empty)_ | client dial certificate (PEM) |
| `GRPC_TLS_CLIENT_KEY_FILE` | _(empty)_ | client dial private key (PEM) |
| `GRPC_TLS_SERVER_NAME` | _(empty)_ | SNI override for client cert verification (e.g. `localhost`) |
| `PAYMENT_SERVICE_URL` | `http://localhost:8082` | gateway |
| `CARD_SERVICE_URL` | `http://localhost:8084` | gateway |
| `NOTIFICATION_SERVICE_URL` | `http://localhost:8083` | gateway, outbox |
| `KAFKA_BROKERS` | _(empty)_ | user, payment, card, notification; `localhost:19092` on host, `redpanda:9092` in compose |
| `SETTLEMENT_LEDGER_ACCOUNT_ID` | _(empty)_ | card (capture) |
| `RAILS_SERVICE_URL` | `http://localhost:8090` | payment (issues virtual IBANs) |
| `RAILS_WEBHOOK_SECRET` | `dev-rails-webhook-secret` | payment, rails simulator (shared HMAC secret) |
| `RAILS_SETTLEMENT_LEDGER_ACCOUNT_ID` | _(empty)_ | payment (inbound bank transfer credit + outbound bank transfer debit/return); see [docker-compose.rails-override.yml](deployments/docker-compose.rails-override.yml) |
| `PAYMENT_WEBHOOK_URL` | `http://localhost:8082/webhooks/rails` | rails simulator (also used for `rails.payment.settled`/`.returned`/`.failed`) |
| `CARDPROC_SERVICE_URL` | `http://localhost:8091` | card (issues virtual cards) |
| `CARDPROC_WEBHOOK_SECRET` | `dev-cardproc-webhook-secret` | card, cardproc simulator (shared HMAC secret) |
| `CARD_SERVICE_AUTHORIZE_URL` | `http://localhost:8084/webhooks/cardproc/authorize` | cardproc simulator (synchronous auth decision) |
| `CARD_SERVICE_EVENTS_URL` | `http://localhost:8084/webhooks/cardproc/events` | cardproc simulator (async capture/reversal/expiry/chargeback) |
| `CARDPROC_AUTH_TTL` | `5m` | cardproc simulator (background sweep expires unauthorized-but-uncaptured holds older than this) |
| `CARDPROC_AUTH_SWEEP_INTERVAL` | `30s` | cardproc simulator (how often the auth expiry sweep runs) |
| `KYC_VENDOR_SERVICE_URL` | `http://localhost:8092` | user (submits applicants) |
| `KYC_WEBHOOK_SECRET` | `dev-kyc-webhook-secret` | user, kyc simulator (shared HMAC secret) |
| `USER_SERVICE_EVENTS_URL` | `http://localhost:8081/webhooks/kyc/events` | kyc simulator (async verdict) |
| `FX_SERVICE_URL` | `http://localhost:8093` | payment (quotes and executes conversions) |
| `FX_POSITION_ACCOUNT_EUR` / `_USD` / `_GBP` | _(empty)_ | payment (FX desk inventory per currency); see [docker-compose.fx-override.yml](deployments/docker-compose.fx-override.yml) |
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

make up / make down       # docker-compose infra (Postgres, Redis, Redpanda, Vault, OTel)
make up-all / down-all    # infra + migrate + all five services (local build)
make up-ghcr / down-ghcr  # same stack from GHCR images (IMAGE_TAG=sha-<short>|latest)
make up-obs-all           # full stack + Tempo, Prometheus, Grafana, Loki, ops-metrics
make migrate              # golang-migrate all services
make migrate-user      # … payment, notification, card
make vault-init        # local Vault Transit keys
make grpc-mtls-certs   # dev PKI for internal gRPC mTLS

make reconcile-payment / reconcile-card
make list-payment-breaks / list-card-breaks
make saga-watchdog / list-saga-alerts
make aml-export
make event-catalog

make up-jobs / down-jobs   # cron: reconcile + saga-watchdog

make helm-lint / helm-template   # Kubernetes chart (deploy/helm/neobank)
helm upgrade --install neobank deploy/helm/neobank -f deploy/helm/neobank/values-production.yaml -n neobank --create-namespace
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
| `auth` / `userclient` / `money` | JWT, internal user gRPC client, decimals |

## Patterns

- **Saga (orchestration)** — P2P, card auth, wallet provision: multi-step, compensating, state in `saga_instances`.
- **Outbox + choreography** — After domain commit, events notify and project without a central workflow engine.
- **Idempotency** — HTTP `Idempotency-Key` + domain unique constraints + ledger idempotency keys on transfers.
- **Append-only evidence** — `audit_log`, `fraud_decisions`, outbox payloads: DB triggers reject UPDATE/DELETE.

## Testing & CI

```bash
make test
make test-integration   # ~1–2 min, requires Docker
```

Integration tests cover P2P, card auth/capture, wallet projection dedup, notification dedup, saga watchdog, outbox/audit immutability, GDPR, PII read audit, KYC screening, and AML.

**CI** ([.github/workflows/ci.yml](.github/workflows/ci.yml)): unit tests, `make build` (with buf lint + codegen), `make lint`, and integration tests on every push/PR to `main`.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `connection refused` on wallet/KYC | goledger not running | Start goledger on `:50051` |
| `400` on register/login/transfer | Missing `Idempotency-Key` | Add header on all mutating calls |
| `401` on gateway | Expired JWT or wrong `APP_ENV` | Re-login; or use dev token only in `development` |
| Migrations fail | Postgres not up or wrong `DATABASE_URL` | `make up`, then `make migrate` |
| PII fields unreadable | Vault enabled but keys missing | `make vault-init` or unset `VAULT_ADDR` |
| Outbox events not projecting | User/Notification not running | Start all five binaries; check service URLs |
| Card capture fails | No settlement account | Set `SETTLEMENT_LEDGER_ACCOUNT_ID` |
| Integration tests hang | Docker not running | Start Docker Desktop / daemon |

## Documentation

- [docs/architecture.md](docs/architecture.md) — system design, schemas, events (extended notes)
- [todo.md](todo.md) — regulatory backlog and status
- [services/ledger/README.md](services/ledger/README.md) — goledger integration

## Roadmap (deferred)

- Standalone Fraud service (today: `pkg/fraud` in Payment/Card)
- Real KYC/AML vendors (today: stubs with persisted evidence)
- Production Vault HA / AppRole (dev Vault via `make vault-init`)
- Kubernetes manifests and production Vault (HA, AppRole, auto-unseal)
- Outbox partition + WORM archival to object storage
- Background saga recovery worker (today: client retry + watchdog alerts)

## License

See [LICENSE](LICENSE).