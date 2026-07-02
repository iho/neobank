// saga-watchdog scans user, payment, and card saga_instances for workflows
// stuck in running/compensating state and records alerts for operators.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/sagawatchdog"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var (
		databaseURL = flag.String("database-url", envOr("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"), "PostgreSQL URL")
		staleAfter  = flag.Duration("stale-after", 15*time.Minute, "consider sagas stuck after this duration without progress")
		schema      = flag.String("schema", "all", "schema to scan: user, payment, card, or all")
		list        = flag.Bool("list", false, "list open saga alerts")
		resolveID   = flag.String("resolve-id", "", "alert UUID to mark investigating/resolved")
		resolveStatus = flag.String("resolve-status", "", "investigating or resolved")
		resolvedBy  = flag.String("by", "", "operator identity for resolve")
		notes       = flag.String("notes", "", "optional notes when resolving")
		limit       = flag.Int("limit", 50, "max alerts to list")
	)
	flag.Parse()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, *databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database connect failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	scanner := sagawatchdog.New(pool)

	if *list {
		if err := listAlerts(ctx, scanner, *schema, *limit); err != nil {
			fmt.Fprintf(os.Stderr, "list alerts: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *resolveID != "" {
		if *resolveStatus == "" || *resolvedBy == "" {
			fmt.Fprintln(os.Stderr, "usage: saga-watchdog -schema <schema> -resolve-id <uuid> -resolve-status investigating|resolved -by <operator> [-notes <text>]")
			os.Exit(2)
		}
		if *schema == "" || *schema == "all" {
			fmt.Fprintln(os.Stderr, "-schema is required when resolving (user, payment, or card)")
			os.Exit(2)
		}
		id, err := uuid.Parse(*resolveID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid alert id: %v\n", err)
			os.Exit(2)
		}
		ok, err := scanner.ResolveAlert(ctx, *schema, id, *resolveStatus, *resolvedBy, *notes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve alert: %v\n", err)
			os.Exit(1)
		}
		if !ok {
			fmt.Fprintf(os.Stderr, "alert %s not found or already resolved\n", id)
			os.Exit(1)
		}
		fmt.Printf("alert %s marked %s by %s\n", id, *resolveStatus, *resolvedBy)
		return
	}

	schemas, err := targetSchemas(*schema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	totalOpen := 0
	for _, sch := range schemas {
		result, err := scanner.Scan(ctx, sch, *staleAfter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan %s: %v\n", sch, err)
			os.Exit(1)
		}
		fmt.Printf("schema=%s stuck_found=%d alerts_open=%d auto_resolved=%d\n",
			result.Schema, result.StuckFound, result.AlertsOpen, result.AutoResolved)
		totalOpen += result.AlertsOpen
	}

	if totalOpen > 0 {
		fmt.Fprintf(os.Stderr, "saga-watchdog: %d open alert(s) — see *.saga_alerts and run with -list\n", totalOpen)
		os.Exit(1)
	}
}

func listAlerts(ctx context.Context, scanner *sagawatchdog.Scanner, schema string, limit int) error {
	schemas, err := targetSchemas(schema)
	if err != nil {
		return err
	}
	any := false
	for _, sch := range schemas {
		alerts, err := scanner.ListOpenAlerts(ctx, sch, limit)
		if err != nil {
			return err
		}
		for _, a := range alerts {
			any = true
			fmt.Printf("%s  alert=%s  saga=%s  type=%s  key=%s  instance_status=%s  alert_status=%s  stuck_since=%s\n",
				sch, a.ID, a.SagaInstanceID, a.SagaType, a.IdempotencyKey, a.InstanceStatus, a.AlertStatus,
				a.StuckSince.Format(time.RFC3339))
		}
	}
	if !any {
		fmt.Println("no open saga alerts")
	}
	return nil
}

func targetSchemas(schema string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(schema)) {
	case "", "all":
		return sagawatchdog.AllSchemas, nil
	case "user", "payment", "card":
		return []string{schema}, nil
	default:
		return nil, fmt.Errorf("invalid schema %q: use user, payment, card, or all", schema)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}