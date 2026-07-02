# Vault operations (PII encryption)

## Local dev

```bash
make vault-init
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-root-token
```

## Production HA

Production Vault runs outside this chart (HA raft + auto-unseal). If seals engage:

1. Confirm auto-unseal KMS/Transit health.
2. Use break-glass unseal keys per your Vault runbook.
3. Verify Transit keys `pii` and `pii-phone` exist: `vault read transit/keys/pii`

User service reads `VAULT_ADDR` and `VAULT_TOKEN` (or k8s auth in prod). Without Vault, PII storage falls back to noop encryption — unacceptable in production.