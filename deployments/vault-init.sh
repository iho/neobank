#!/usr/bin/env bash
# Bootstrap Vault Transit keys for local PII field encryption.
set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://localhost:8200}"
export VAULT_TOKEN="${VAULT_TOKEN:-dev-root-token}"

vault secrets enable transit 2>/dev/null || true
vault write -f transit/keys/pii type=aes256-gcm96
vault write -f transit/keys/pii-phone type=hmac key_size=256

echo "Vault transit keys ready: pii (encrypt), pii-phone (hmac)"