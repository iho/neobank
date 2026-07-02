# Database restore drill

## Scope

Single Postgres cluster, schema-per-service (`user`, `payment`, `card`, `notification`).

## Backup (operator responsibility)

Use your managed Postgres PITR (CloudNativePG, RDS, etc.). This repo does not deploy the database.

## Restore drill (quarterly)

1. Restore a point-in-time clone to an isolated instance.
2. Run migrations: `helm upgrade` migrate hook or `make migrate` against the clone.
3. Smoke test: register/login via gateway against the clone.
4. Run reconciliation jobs — expect zero new breaks if ledger matches.
5. Record RTO/RPO achieved and gaps.

## Outbox archival

Monthly CronJob (`outbox-archiver`) exports published events older than the retention window to JSONL for WORM object storage. Upload archives to the compliance bucket before dropping old partitions.