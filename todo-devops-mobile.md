# DevOps & Flutter Mobile Client — Plan / TODO

Companion to [todo.md](todo.md) (compliance backlog). Scope here: getting the backend
deployable beyond a laptop, and building the mobile client the BFF was designed for.

## Current state (analysis summary)

**What exists**
- 5 Go services (gateway :8080, user :8081, payment :8082, notification :8083, card :8084)
  + goledger (:50051), clean architecture, per-service migrations (golang-migrate).
- Local infra via `deployments/docker-compose.yml`: Postgres 16, Redis 7, Redpanda v26 (Kafka API),
  Vault dev, OTel collector.
- Full stack compose: `make up-all` (services + migrate job + infra). Optional goledger:
  `make up-all-ledger` (in-compose goledger, `LEDGER_GRPC_ADDR=goledger:50051`).
- GHCR image publish on `main`, staging deploy workflow, Helm chart (`deploy/helm/neobank`),
  platform chart (`deploy/helm/platform`), ArgoCD apps (`deploy/argocd/`).
- Observability: Tempo, Prometheus, Grafana, Loki, ops-metrics, alerts, runbooks.
- Production guards: `APP_ENV=production` + non-default `JWT_SECRET` enforced at startup.

**Remaining gaps**
- goledger in k8s (still external; compose overlay covers local/staging VM).
- Vault auto-unseal (KMS) — bootstrap scripts exist, operator config is environment-specific.
- CNPG backup destination (S3/GCS) must be filled in `values-production.yaml` before prod.
- Mobile client — see [mobile/TODO.md](mobile/TODO.md).

---

## Part 1 — DevOps

### Phase 1: Containerize & compose the full stack (local parity)

- [x] Multi-stage Dockerfile per service (`deployments/Dockerfile.service` with `--build-arg SERVICE=`):
      distroless, non-root, static build, `/health` HEALTHCHECK.
- [x] `deployments/docker-compose.services.yml` — `make up-all` runs five services + infra with
      service-DNS URLs and migrate one-shot job.
- [x] goledger compose overlay (`deployments/docker-compose.goledger.yml`) — `make up-all-ledger`.
- [x] Migrations as compose one-shot (`Dockerfile.migrate` + `migrate` service).
- [x] `.dockerignore`, reproducible builds (`-trimpath`, `CGO_ENABLED=0`), image labels (git SHA).

### Phase 2: CI → images → CD

- [x] CI job: build & push images to GHCR on `main` (tags: `sha-<short>`, `latest`).
- [x] Vulnerability scanning (Trivy) + SBOM (syft); `govulncheck` on code.
- [x] Version/release flow: tag → GitHub Release → immutable image tags.
- [x] Staging deploy on merge to `main` with smoke test (`/health` + register/login).
- [x] Gate deploys on migrations succeeding.

### Phase 3: Kubernetes (README roadmap item)

- [x] Helm chart + overlays (staging/prod): Deployment, Service, HPA, PDB, probes, securityContext.
- [x] Ingress + cert-manager TLS on gateway; rate limiting at ingress.
- [x] Migration Jobs (pre-upgrade hook).
- [x] CronJobs: reconcile-payment, reconcile-card, saga-watchdog, aml-export.
- [x] Platform chart: CloudNativePG cluster + scheduled backups, Redis, NetworkPolicies
      (`deploy/helm/platform`). Redpanda via ArgoCD (`deploy/argocd/deps/redpanda.yaml`).
- [x] Production Vault: HA raft Helm values + bootstrap scripts (`deploy/vault/`).
      Auto-unseal is environment-specific (documented in README).
- [x] External Secrets Operator + `vault-neobank` ClusterSecretStore; prod JWT guard.
- [x] NetworkPolicies on app chart; prod `KAFKA_BROKERS=redpanda:9092` in values-production.

### Phase 4: Observability & operations

- [x] Tracing: OTel collector → Tempo.
- [x] Metrics: `/metrics` on services, ops-metrics, Grafana RED dashboard.
- [x] Alerting: saga, reconciliation, outbox lag, 5xx, Redpanda consumer lag + disk,
      Postgres size/disk, cert expiry (cert-manager).
- [x] Log shipping: Loki + promtail with `correlation_id`.
- [x] Outbox archival CronJob + compose tooling.
- [x] Runbooks: deploy, breaks, stuck saga, Vault, DB restore.

---

## Part 2 — Flutter mobile client

Moved to [mobile/TODO.md](mobile/TODO.md) now that the app is scaffolded under `mobile/`.

---

## Suggested order of work

1. ~~DevOps Phase 1~~ — done.
2. Mobile Phase 1 + 2 in parallel with staging deploy — client only needs a running gateway.
3. Fill CNPG S3 backup path + Vault auto-unseal before real users.
4. Mobile Phase 3+; push notifications last.