# Ledger Service (goledger)

The ledger runs from the standalone [goledger](https://github.com/iho/goledger) repository.

## Local setup

```bash
git clone https://github.com/iho/goledger.git /tmp/goledger
cd /tmp/goledger
docker compose -f docker-compose.full.yml up -d
./scripts/setup-and-test.sh
```

Default endpoints:

- gRPC: `localhost:50051`
- HTTP: `localhost:8080`

Neobank services connect via `LEDGER_GRPC_ADDR=localhost:50051`.

## Integration

- Shared client: `github.com/iho/neobank/pkg/ledgerclient`
- Protobuf: import `goledger/v1` protos via buf (see `proto/buf.yaml`)
- Only Payment and Card services should mutate balances through goledger