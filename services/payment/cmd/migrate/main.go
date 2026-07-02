package main

import (
	"context"
	"fmt"
	"os"

	"github.com/iho/neobank/services/payment/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	cfg := config.Load()
	sql, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read migration: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, string(sql)); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("payment service migrations applied")
}