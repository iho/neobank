#!/usr/bin/env bash
# Copy to /tmp/seed-neobank.sh and fill in real values. Never commit secrets.
set -euo pipefail

: "${VAULT_ADDR:?}"
: "${VAULT_TOKEN:?}"

vault kv put secret/neobank/database-url \
  value='postgres://neobank:REPLACE_ME@neobank-postgres-rw.neobank.svc:5432/neobank?sslmode=require'

vault kv put secret/neobank/redis-url \
  value='redis://platform-redis.neobank.svc:6379/0'

vault kv put secret/neobank/jwt-secret \
  value='REPLACE_ME_LONG_RANDOM'

vault kv put secret/neobank/kafka-brokers \
  value='redpanda:9092'

vault kv put secret/neobank/ledger-grpc-addr \
  value='goledger:50051'

vault kv put secret/neobank/otel-endpoint \
  value='http://otel-collector.monitoring.svc:4317'

vault kv put secret/neobank/vault-addr \
  value='http://vault.vault.svc:8200'

vault kv put secret/neobank/settlement-ledger-account-id \
  value='REPLACE_ME'

echo ">>> Neobank secrets seeded"