module github.com/iho/neobank/tools/saga-watchdog

go 1.25.0

toolchain go1.25.11

require (
	github.com/google/uuid v1.6.0
	github.com/iho/neobank/pkg v0.0.0
	github.com/jackc/pgx/v5 v5.10.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)

replace github.com/iho/neobank/pkg => ../../pkg
