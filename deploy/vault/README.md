# HashiCorp Vault (production)

Vault is the source of truth for database URLs, JWT secrets, and integration credentials.

## Secret paths

| Path | Keys | Consumers |
|------|------|-----------|
| `secret/neobank/database-url` | `value` | neobank services (via ESO) |
| `secret/neobank/redis-url` | `value` | gateway, user, payment, card |
| `secret/neobank/jwt-secret` | `value` | gateway, user |
| `secret/neobank/kafka-brokers` | `value` | all services (`redpanda:9092`) |
| `secret/neobank/ledger-grpc-addr` | `value` | user, payment, card |
| `secret/neobank/vault-addr` | `value` | user service (Transit PII) |
| `secret/neobank/vault-token` | `value` | user service (dev only — prefer K8s auth in prod) |

## Deploy order

```bash
kubectl apply -f deploy/argocd/deps/vault.yaml
kubectl apply -f deploy/argocd/deps/external-secrets.yaml

# Initialize and unseal (once per cluster)
kubectl -n vault port-forward svc/vault 8200:8200 &
export VAULT_ADDR=http://127.0.0.1:8200
vault operator init -key-shares=5 -key-threshold=3
vault operator unseal  # repeat with 3 keys on each replica

export VAULT_TOKEN=<root-token>
deploy/vault/bootstrap/configure.sh

cp deploy/vault/bootstrap/seed-secrets.example.sh /tmp/seed-neobank.sh
# edit REPLACE_ME values, then:
bash /tmp/seed-neobank.sh

kubectl apply -f deploy/argocd/application-vault-config.yaml
```

Auto-unseal (AWS KMS / GCP CKM / HSM) is environment-specific — see
`deploy/vault/helm-values-auto-unseal-aws.example.yaml` for an AWS KMS `awskms` seal
snippet to merge into `deploy/argocd/deps/vault.yaml` Helm values before production cutover.