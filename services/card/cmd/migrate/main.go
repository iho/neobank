package main

import (
	"context"
	"fmt"
	"os"

	"github.com/iho/neobank/services/card/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	cfg := config.Load()
	migrations := []string{
		"migrations/001_init.sql",
		"migrations/002_authorizations.sql",
		"migrations/003_traceability.sql",
		"migrations/004_fraud_rule_version.sql",
		"migrations/005_reconciliation_breaks.sql",
		"migrations/006_saga_alerts.sql",
		"migrations/007_outbox_immutability.sql",
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	for _, path := range migrations {
		sql, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read migration %s: %v\n", path, err)
			os.Exit(1)
		}
		if _, err := conn.Exec(ctx, string(sql)); err != nil {
			fmt.Fprintf(os.Stderr, "migrate %s: %v\n", path, err)
			os.Exit(1)
		}
	}

	fmt.Println("card service migrations applied")
}
