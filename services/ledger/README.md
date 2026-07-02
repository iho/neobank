# Ledger Service (goledger)

The ledger runs from the standalone [goledger](https://github.com/iho/goledger) repository.

## Local setup

**Option A — in-compose (recommended with `make up-all`)**

```bash
make up-all-ledger    # neobank infra + services + goledger on the compose network
```

Services use `LEDGER_GRPC_ADDR=goledger:50051`. Host binaries can still reach gRPC on
`localhost:50051`. Pin a release with `GOLEDGER_REF=v1.0.0 make up-all-ledger`.

**Option B — standalone goledger**

```bash
git clone https://github.com/iho/goledger.git /tmp/goledger
cd /tmp/goledger
docker compose -f docker-compose.full.yml up -d
./scripts/setup-and-test.sh
```

Default endpoints:

- gRPC: `localhost:50051`
- HTTP: `localhost:8080`

Use `make up-all` (without ledger overlay) and `LEDGER_GRPC_ADDR=host.docker.internal:50051`
inside compose, or `localhost:50051` for host-run binaries.

## Integration

- Shared client: `github.com/iho/neobank/pkg/ledgerclient`
- Protobuf: import `goledger/v1` protos via buf (see `proto/buf.yaml`)
- Only Payment and Card services should mutate balances through goledger