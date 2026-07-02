# DevOps & Flutter Mobile Client — Plan / TODO

Companion to [todo.md](todo.md) (compliance backlog). Scope here: getting the backend
deployable beyond a laptop, and building the mobile client the BFF was designed for.

## Current state (analysis summary)

**What exists**
- 5 Go services (gateway :8080, user :8081, payment :8082, notification :8083, card :8084)
  + external goledger (:50051), clean architecture, per-service migrations (golang-migrate).
- Local infra via `deployments/docker-compose.yml`: Postgres 16, Redis 7, Redpanda v26 (Kafka API),
  Vault dev, OTel collector (**debug exporter only** — traces go nowhere).
- Cron jobs containerized (`Dockerfile.jobs` + `docker-compose.jobs.yml`): reconciliation,
  saga watchdog.
- CI (`.github/workflows/ci.yml`): unit tests, buf lint + codegen + build, golangci-lint,
  testcontainers integration suite.

**Gaps**
- Services themselves are **not containerized** — no Dockerfiles, run as `./bin/*` binaries.
- No image registry publishing, no CD, no environments (staging/prod), no k8s manifests
  (explicitly deferred in README roadmap).
- No metrics/alerting (no Prometheus/Grafana), no log shipping, no tracing backend.
- Vault runs in dev mode (root token, no persistence); prod Vault (HA, AppRole, auto-unseal)
  is a known todo.
- No Postgres backup/restore or DR story; Redpanda is single-node dev-grade.
- No TLS/ingress/rate limiting in front of the gateway.
- **No mobile client** — but `services/gateway/api/openapi.yaml` is a complete contract
  (JWT auth + refresh, `Idempotency-Key` on mutations, `X-Correlation-Id`) to build against.

---

## Part 1 — DevOps

### Phase 1: Containerize & compose the full stack (local parity)

- [ ] Multi-stage Dockerfile per service (`services/*/Dockerfile` or one parameterized
      `deployments/Dockerfile.service` with `--build-arg SERVICE=`): distroless/alpine,
      non-root, static build, `/health` HEALTHCHECK.
- [ ] Extend `deployments/docker-compose.yml` (or add `docker-compose.services.yml`) so
      `make up-all` runs the five services + infra with correct `depends_on`/healthchecks
      and service-DNS URLs (`USER_SERVICE_URL=http://user:8081`, …).
- [ ] Fold goledger into the compose story (git submodule, published image, or documented
      `docker-compose.full.yml` bridge network) — today it's a manual out-of-band step.
- [ ] Run migrations as a one-shot compose service / init container (`pkg/migrate` CLI)
      instead of host-side `make migrate`.
- [ ] `.dockerignore`, reproducible builds (`-trimpath`, `CGO_ENABLED=0`), image labels
      (git SHA, build date).

### Phase 2: CI → images → CD

- [x] CI job: build & push images to GHCR on `main` (tags: `sha-<short>`, `latest`),
      with Go build cache + buildx layer cache.
- [x] Vulnerability scanning (Trivy/Grype) + SBOM (syft) on images; `govulncheck` job on code.
- [x] Version/release flow: tag → GitHub Release → immutable image tags.
- [x] Staging environment deploy on merge to `main` (compose on a VM is fine as step one;
      k8s in Phase 3). Smoke test job hits `/health` + register/login after deploy.
- [x] Gate deploys on migrations succeeding (run migrator, then roll services).

### Phase 3: Kubernetes (README roadmap item)

- [ ] Helm chart or Kustomize base + overlays (staging/prod) for the 5 services:
      Deployment, Service, HPA, PDB, resource requests/limits, liveness=`/health`,
      readiness probe, securityContext (non-root, read-only FS).
- [ ] Ingress (nginx/traefik) + cert-manager TLS in front of gateway only; internal
      services ClusterIP-only. Rate limiting at ingress (gateway has none).
- [ ] Migration Jobs (pre-upgrade hook) per service.
- [ ] CronJobs replacing `deployments/crontab`: reconcile-payment, reconcile-card,
      saga-watchdog (hourly UTC), aml-export.
- [ ] Managed/prod-grade stateful deps: Postgres (CloudNativePG or RDS-equivalent) with
      PITR backups + tested restore runbook; Redis (HA or managed); Redpanda (operator or
      managed) — flip services from HTTP fan-out to `KAFKA_BROKERS` in prod.
- [ ] Production Vault: HA raft, auto-unseal, AppRole/k8s auth for the user service,
      Transit key rotation policy (closes todo.md #7 ops half).
- [ ] Secrets: External Secrets Operator or sealed-secrets; kill `JWT_SECRET` default
      (`dev-secret-change-me`) — fail startup in prod if unset. `APP_ENV=production`
      enforced (disables dev-auth per todo.md #7b).
- [ ] NetworkPolicies: only gateway reachable from ingress; internal `/api/v1/internal/*`
      endpoints (GDPR, events ingest, user-by-phone) unreachable from outside the mesh.

### Phase 4: Observability & operations

- [ ] Tracing backend: point OTel collector at Tempo/Jaeger instead of `debug` exporter;
      trace across gateway → services → goledger (propagation already wired via `pkg/reqctx`/otel).
- [ ] Metrics: Prometheus scrape (add `/metrics` via OTel metrics or promhttp to services),
      Grafana dashboards — RED per service, saga latency/failures, outbox lag
      (unpublished `outbox_events` age), reconciliation break count, Kafka consumer lag.
- [ ] Alerting (Alertmanager/PagerDuty): saga_alerts rows, reconciliation exit 1, outbox
      lag threshold, 5xx rate, cert expiry, DB disk.
- [ ] Log shipping (Loki/ELK) with retention — `pkg/sloghttp` already emits structured
      JSON with `correlation_id` (closes todo.md #9 "retained log shipping" note).
- [ ] Outbox archival infra (todo.md #5): monthly partitions + export to object storage
      with WORM/object-lock; 7-year retention (`outbox.DefaultRetentionYears`).
- [ ] Runbooks in `docs/`: deploy/rollback, break resolution, stuck saga, Vault unseal,
      DB restore drill.

---

## Part 2 — Flutter mobile client

New top-level `mobile/` directory (own toolchain; keep out of Go workspace).

### Phase 1: Foundation

- [ ] Scaffold Flutter app (`mobile/`): flavors dev/staging/prod (`--dart-define` for
      `API_BASE_URL`), lint rules (`very_good_analysis` or `flutter_lints`).
- [ ] Architecture: feature-first folders + Riverpod (state) + go_router (navigation) +
      freezed/json_serializable models. Decision recorded in `mobile/README.md`.
- [ ] API client generated from `services/gateway/api/openapi.yaml`
      (openapi-generator dart-dio or swagger_parser) — wire into `make generate` so the
      contract stays the single source of truth.
- [ ] Dio interceptors:
      - JWT bearer injection + transparent refresh on 401 (single-flight refresh,
        logout on refresh failure) — tokens in `flutter_secure_storage`.
      - `Idempotency-Key: uuid` on every POST (gateway 400s without it).
      - `X-Correlation-Id` generation for supportability.
- [ ] Error model: map gateway error envelope → typed failures → user-facing messages;
      global retry/backoff policy for idempotent GETs.

### Phase 2: Core flows (mirrors backend MVP table)

- [ ] Auth: register, login, session restore, logout, token refresh edge cases.
- [ ] Onboarding/KYC: submit KYC form (`POST /v1/kyc`), poll `GET /v1/kyc/status`,
      gate wallet features on `approved`.
- [ ] Wallet home: balance (`GET /v1/wallet`), unified transaction history
      (`GET /v1/wallet/transactions`) with pagination + pull-to-refresh.
- [ ] P2P transfer: recipient by phone, amount entry (minor-units/decimal handling
      consistent with `pkg/money`), confirm screen, result states incl. fraud-declined;
      retry with the **same** Idempotency-Key on timeout (matches saga retry semantics).
- [ ] Cards: list, issue virtual card, card detail (masked PAN), freeze/unfreeze;
      authorizations list + detail.
- [ ] Notifications inbox: `GET /v1/notifications`, unread badge, polling first
      (push is Phase 4 — requires backend FCM work).

### Phase 3: Quality & mobile CI

- [ ] Unit tests (usecases/notifiers), widget tests for auth + transfer flows,
      one integration test (patrol/integration_test) against local compose stack.
- [ ] Contract safety: CI check that regenerating the Dart client from openapi.yaml is
      clean (same pattern as backend codegen check).
- [ ] `mobile.yml` GitHub workflow: analyze, test, build APK + iOS (no codesign) on PRs
      touching `mobile/`; path-filter so Go CI doesn't run for mobile-only changes.
- [ ] Security pass: certificate pinning (staging/prod), no secrets in code, screenshot
      obscuring on app switcher for balance screens, jailbreak/root detection decision.

### Phase 4: Release & beyond

- [ ] Fastlane lanes + CI: TestFlight / Play internal track from tags; build number from CI.
- [ ] Crash/analytics: Sentry (or Firebase Crashlytics) with correlation-id breadcrumbs.
- [ ] Push notifications: **requires backend work** — FCM/APNs sender in notification
      service (device token registry + send on event ingest); then deep links from push
      to transaction/card screens.
- [ ] Biometric unlock (local_auth) gating app open + transfer confirmation.
- [ ] Localization scaffold (intl), dark mode, accessibility audit.

---

## Suggested order of work

1. DevOps Phase 1 (containerize) — unblocks everything incl. the mobile integration test rig.
2. Mobile Phase 1 + 2 in parallel with DevOps Phase 2 — client only needs a running gateway.
3. DevOps Phase 3 (k8s + prod Vault/secrets) before any real users.
4. Observability (DevOps 4) alongside mobile Phase 3; push notifications last since they
   need new backend surface.
