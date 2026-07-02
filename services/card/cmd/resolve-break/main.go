// resolve-break updates the status of a card reconciliation break
// (open → investigated → closed) so operators can track investigation to closure.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/services/card/internal/config"
	"github.com/iho/neobank/services/card/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var (
		breakID = flag.String("id", "", "break UUID to update")
		status  = flag.String("status", "", "new status: investigated or closed")
		by      = flag.String("by", "", "operator identity (email or name)")
		notes   = flag.String("notes", "", "optional investigation notes")
		list    = flag.Bool("list", false, "list open/investigated breaks")
		limit   = flag.Int("limit", 50, "max rows when listing")
	)
	flag.Parse()

	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database connect failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	if *list {
		if err := listBreaks(ctx, queries, *limit); err != nil {
			fmt.Fprintf(os.Stderr, "list breaks: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *breakID == "" || *status == "" || *by == "" {
		fmt.Fprintln(os.Stderr, "usage: resolve-break -id <uuid> -status investigated|closed -by <operator> [-notes <text>]")
		fmt.Fprintln(os.Stderr, "       resolve-break -list [-limit N]")
		os.Exit(2)
	}

	newStatus := strings.ToLower(strings.TrimSpace(*status))
	if newStatus != "investigated" && newStatus != "closed" {
		fmt.Fprintf(os.Stderr, "invalid status %q: must be investigated or closed\n", *status)
		os.Exit(2)
	}

	id, err := uuid.Parse(*breakID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid break id: %v\n", err)
		os.Exit(2)
	}

	now := time.Now().UTC()
	rows, err := queries.ResolveReconciliationBreak(ctx, sqlc.ResolveReconciliationBreakParams{
		ID:         id,
		Status:     newStatus,
		ResolvedBy: pgtype.Text{String: strings.TrimSpace(*by), Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		Notes:      pgtype.Text{String: strings.TrimSpace(*notes), Valid: strings.TrimSpace(*notes) != ""},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve break: %v\n", err)
		os.Exit(1)
	}
	if rows == 0 {
		fmt.Fprintf(os.Stderr, "break %s not found or already closed\n", id)
		os.Exit(1)
	}

	fmt.Printf("break %s marked %s by %s\n", id, newStatus, *by)
}

func listBreaks(ctx context.Context, queries *sqlc.Queries, limit int) error {
	if limit <= 0 {
		limit = 50
	}
	rows, err := queries.ListOpenReconciliationBreaks(ctx, int32(limit))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Println("no open reconciliation breaks")
		return nil
	}
	for _, b := range rows {
		fmt.Printf("%s  %s/%s  reason=%s  status=%s  run=%s  created=%s\n",
			b.ID, b.EntityType, b.EntityID, b.Reason, b.Status, b.RunID, b.CreatedAt.Time.Format(time.RFC3339))
	}
	return nil
}