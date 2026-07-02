# Database restore drill

## Scope

Single Postgres cluster (`neobank-postgres`), schema-per-service (`user`, `payment`, `card`, `notification`).

## Backup

Production uses CloudNativePG barman-cloud backups to object storage. Configure before cutover:

```bash
# Fill deploy/platform/cnpg-backup-values.example.yaml, then:
helm upgrade neobank-platform deploy/helm/platform \
  -f deploy/helm/platform/values-production.yaml \
  -f deploy/platform/cnpg-backup-values.example.yaml \
  -n neobank
```

## Restore drill (quarterly)

```bash
./deploy/scripts/restore-drill.sh neobank neobank-postgres
# Optional PITR target:
./deploy/scripts/restore-drill.sh neobank neobank-postgres "2026-07-01T12:00:00Z"
```

Manual steps after a real incident:

1. Restore a point-in-time clone to an isolated CNPG cluster (or use the drill script pattern).
2. Run migrations: `helm upgrade` migrate hook or `make migrate` against the clone.
3. Smoke test: register/login via gateway against the clone.
4. Run reconciliation jobs — expect zero new breaks if ledger matches.
5. Record RTO/RPO achieved and gaps.

## Outbox archival

Monthly CronJob (`outbox-archiver`) exports published events older than the retention window to JSONL for WORM object storage. Upload archives to the compliance bucket before dropping old partitions.