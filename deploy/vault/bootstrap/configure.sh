#!/usr/bin/env bash
# One-time Vault bootstrap: KV v2, Kubernetes auth, policies, roles.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"

if [[ -z "${VAULT_TOKEN:-}" ]]; then
  echo "Set VAULT_TOKEN to a root or bootstrap token" >&2
  exit 1
fi

echo ">>> Enabling secrets engines and auth methods..."
vault secrets enable -path=secret kv-v2 2>/dev/null || true
vault secrets enable transit 2>/dev/null || true
vault auth enable kubernetes 2>/dev/null || true

K8S_HOST=""
K8S_CA_FILE=""
REVIEWER_JWT=""

IN_CLUSTER_TOKEN="/var/run/secrets/kubernetes.io/serviceaccount/token"
IN_CLUSTER_CA="/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

if [[ -f "$IN_CLUSTER_TOKEN" && -f "$IN_CLUSTER_CA" ]]; then
  K8S_HOST="https://kubernetes.default.svc"
  K8S_CA_FILE="$IN_CLUSTER_CA"
  REVIEWER_JWT="$(cat "$IN_CLUSTER_TOKEN")"
else
  if ! command -v kubectl >/dev/null 2>&1; then
    echo "ERROR: kubectl required outside the cluster" >&2
    exit 1
  fi
  K8S_HOST="$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')"
  K8S_CA_FILE="$(mktemp)"
  kubectl config view --raw --minify -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' \
    | base64 -d > "$K8S_CA_FILE"
  REVIEWER_JWT="$(kubectl create token vault -n vault --duration=10m)"
fi

vault write auth/kubernetes/config \
  kubernetes_host="$K8S_HOST" \
  kubernetes_ca_cert=@"$K8S_CA_FILE" \
  token_reviewer_jwt="$REVIEWER_JWT"

vault policy write neobank-app "$ROOT_DIR/policies/neobank-app.hcl"
vault policy write external-secrets "$ROOT_DIR/policies/external-secrets.hcl"

vault write auth/kubernetes/role/neobank-app \
  bound_service_account_names=neobank \
  bound_service_account_namespaces=neobank \
  policies=neobank-app \
  ttl=1h

vault write auth/kubernetes/role/external-secrets \
  bound_service_account_names=external-secrets \
  bound_service_account_namespaces=external-secrets \
  policies=external-secrets \
  ttl=1h

vault write -f transit/keys/pii type=aes256-gcm96
vault write -f transit/keys/pii-phone type=hmac key_size=256

echo ">>> Vault bootstrap complete"