# Platform dependency Applications

Apply the **AppProject** first (`deploy/argocd/project.yaml`), then sync in this order:

1. `cert-manager.yaml` — TLS for ingress / internal mTLS
2. `vault.yaml` — HashiCorp Vault HA (Raft)
3. `external-secrets.yaml` — Vault → Kubernetes Secrets
4. `cloudnative-pg.yaml` — CNPG operator
5. `redpanda.yaml` — Redpanda cluster (`KAFKA_BROKERS=redpanda:9092`)
6. `application-platform.yaml` — CNPG cluster + Redis (`deploy/helm/platform`)
7. `application-vault-config.yaml` — ClusterSecretStore + neobank ExternalSecret template
8. `application-neobank.yaml` — neobank app chart (`deploy/helm/neobank`)

Vault requires manual init/unseal — see [`deploy/vault/README.md`](../../vault/README.md).

```bash
kubectl apply -f deploy/argocd/project.yaml
kubectl apply -f deploy/argocd/deps/cert-manager.yaml
kubectl apply -f deploy/argocd/deps/vault.yaml
# ... initialize Vault, then continue
```