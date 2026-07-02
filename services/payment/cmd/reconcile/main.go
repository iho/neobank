// reconcile compares payment.transfers against goledger and records any
// drift (a transfer marked completed locally with no matching ledger
// transfer) so operators and auditors have a durable answer to "did we
// check, and what did we find". Intended to run on a schedule (cron/k8s
// CronJob), not as a long-lived process.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/services/payment/internal/config"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type reconciliationBreak struct {
	TransferID       string `json:"transfer_id"`
	LedgerTransferID string `json:"ledger_transfer_id"`
	LocalStatus      string `json:"local_status"`
	Reason           string `json:"reason"`
}

const batchSize = 500

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	ledger, err := ledgerclient.New(ctx, ledgerclient.Config{Addr: cfg.LedgerAddr})
	if err != nil {
		logger.Error("ledger connect failed", "error", err)
		os.Exit(1)
	}
	defer ledger.Close()

	queries := sqlc.New(pool)

	runID := uuid.New()
	startedAt := time.Now().UTC()
	if _, err := queries.StartReconciliationRun(ctx, sqlc.StartReconciliationRunParams{
		ID:        runID,
		StartedAt: pgtype.Timestamptz{Time: startedAt, Valid: true},
	}); err != nil {
		logger.Error("start reconciliation run failed", "error", err)
		os.Exit(1)
	}

	transfers, err := queries.ListTransfersForReconciliation(ctx, batchSize)
	if err != nil {
		logger.Error("list transfers failed", "error", err)
		os.Exit(1)
	}

	breaks := []reconciliationBreak{}
	for _, t := range transfers {
		if !t.LedgerTransferID.Valid || t.LedgerTransferID.String == "" {
			continue
		}
		ledgerTransfer, err := ledger.GetTransfer(ctx, t.LedgerTransferID.String)
		if err != nil {
			logger.Warn("ledger lookup failed, skipping", "transfer_id", t.ID, "error", err)
			continue
		}
		if t.Status == "completed" && ledgerTransfer == nil {
			breaks = append(breaks, reconciliationBreak{
				TransferID:       t.ID.String(),
				LedgerTransferID: t.LedgerTransferID.String,
				LocalStatus:      t.Status,
				Reason:           "completed_locally_but_missing_in_ledger",
			})
		}
	}

	breaksJSON, err := json.Marshal(breaks)
	if err != nil {
		logger.Error("marshal breaks failed", "error", err)
		os.Exit(1)
	}

	status := "clean"
	if len(breaks) > 0 {
		status = "breaks_found"
	}
	if err := queries.FinishReconciliationRun(ctx, sqlc.FinishReconciliationRunParams{
		ID:           runID,
		FinishedAt:   pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		CheckedCount: int32(len(transfers)),
		BreakCount:   int32(len(breaks)),
		Breaks:       breaksJSON,
		Status:       status,
	}); err != nil {
		logger.Error("finish reconciliation run failed", "error", err)
		os.Exit(1)
	}

	logger.Info("reconciliation complete", "run_id", runID, "checked", len(transfers), "breaks", len(breaks))
	if len(breaks) > 0 {
		for _, b := range breaks {
			logger.Warn("reconciliation break", "transfer_id", b.TransferID, "reason", b.Reason)
		}
		fmt.Fprintf(os.Stderr, "reconciliation found %d break(s), see payment.reconciliation_runs (id=%s)\n", len(breaks), runID)
		os.Exit(1)
	}
}
