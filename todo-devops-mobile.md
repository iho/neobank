# DevOps & Flutter Mobile Client — Plan / TODO

Companion to [todo.md](todo.md) (compliance backlog). Scope here: getting the backend
deployable beyond a laptop, and building the mobile client the BFF was designed for.

## Current state (analysis summary)

**What exists**
- 5 Go services + goledger, full compose (`make up-all-ledger`), GHCR publish (incl. `goledger` image).
- k8s: `deploy/helm/neobank` + `deploy/helm/platform` (CNPG, Redis, goledger), ArgoCD apps,
  kube-prometheus-stack monitoring, Vault HA bootstrap, restore drill script.
- Observability: compose stack + k8s PrometheusRules + ServiceMonitors.
- Local k8s: `deploy/kind/bootstrap.sh`.

**Remaining gaps (environment-specific)**
- CNPG S3 backup bucket + credentials (template: `deploy/platform/cnpg-backup-values.example.yaml`).
- Vault auto-unseal KMS config (template: `deploy/vault/helm-values-auto-unseal-aws.example.yaml`).
- Mobile client — see [mobile/TODO.md](mobile/TODO.md).

---

## Part 1 — DevOps

### Phase 1: Containerize & compose — [x] complete

### Phase 2: CI → images → CD — [x] complete

### Phase 3: Kubernetes — [x] complete

- Platform chart deploys CNPG, Redis, **goledger** (CNPG + setup Job + Deployment).
- ArgoCD deps: cert-manager, Vault, ESO, CNPG operator, Redpanda, monitoring stack.
- Staging VM deploy includes goledger overlay.

### Phase 4: Observability & operations — [x] complete

- k8s: `prometheus-grafana` + `monitoring-config` ArgoCD apps, ServiceMonitors on neobank chart.
- Restore drill: `deploy/scripts/restore-drill.sh`.

---

## Part 2 — Flutter mobile client

Moved to [mobile/TODO.md](mobile/TODO.md).

---

## Before production cutover

1. Fill CNPG backup values + run quarterly restore drill.
2. Configure Vault auto-unseal for your cloud.
3. Set Grafana admin password / SSO; wire Alertmanager receivers.