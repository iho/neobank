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
	EntityType       string `json:"entity_type"`
	EntityID         string `json:"entity_id"`
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

	bankTransfers, err := queries.ListBankTransfersForReconciliation(ctx, batchSize)
	if err != nil {
		logger.Error("list bank transfers failed", "error", err)
		os.Exit(1)
	}

	bankTransferOrders, err := queries.ListBankTransferOrdersForReconciliation(ctx, batchSize)
	if err != nil {
		logger.Error("list bank transfer orders failed", "error", err)
		os.Exit(1)
	}

	checkedCount := len(transfers) + len(bankTransfers) + len(bankTransferOrders)
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
				EntityType:       "transfer",
				EntityID:         t.ID.String(),
				LedgerTransferID: t.LedgerTransferID.String,
				LocalStatus:      t.Status,
				Reason:           "completed_locally_but_missing_in_ledger",
			})
		}
	}

	// Every rails inbound transfer we've credited must have a matching
	// ledger transfer — the statement the simulator exposes at
	// GET /v1/statements/{date} is the source of truth this doesn't yet
	// cross-check (see docs/vendor-simulators-plan.md Phase 1b).
	for _, bt := range bankTransfers {
		if bt.LedgerTransferID == "" {
			continue
		}
		ledgerTransfer, err := ledger.GetTransfer(ctx, bt.LedgerTransferID)
		if err != nil {
			logger.Warn("ledger lookup failed, skipping", "bank_transfer_id", bt.ID, "error", err)
			continue
		}
		if ledgerTransfer == nil {
			breaks = append(breaks, reconciliationBreak{
				EntityType:       "bank_transfer",
				EntityID:         bt.ID.String(),
				LedgerTransferID: bt.LedgerTransferID,
				LocalStatus:      bt.Status,
				Reason:           "completed_locally_but_missing_in_ledger",
			})
		}
	}

	// Every settled/returned/failed outbound order must have its debit
	// transfer in the ledger; returned/failed orders must also have their
	// return transfer.
	for _, o := range bankTransferOrders {
		if o.LedgerTransferID != "" {
			ledgerTransfer, err := ledger.GetTransfer(ctx, o.LedgerTransferID)
			if err != nil {
				logger.Warn("ledger lookup failed, skipping", "bank_transfer_order_id", o.ID, "error", err)
			} else if ledgerTransfer == nil {
				breaks = append(breaks, reconciliationBreak{
					EntityType:       "bank_transfer_order",
					EntityID:         o.ID.String(),
					LedgerTransferID: o.LedgerTransferID,
					LocalStatus:      o.Status,
					Reason:           "debit_completed_locally_but_missing_in_ledger",
				})
			}
		}

		if (o.Status == "returned" || o.Status == "failed") && o.ReturnTransferID != "" {
			returnTransfer, err := ledger.GetTransfer(ctx, o.ReturnTransferID)
			if err != nil {
				logger.Warn("ledger lookup failed, skipping", "bank_transfer_order_id", o.ID, "error", err)
			} else if returnTransfer == nil {
				breaks = append(breaks, reconciliationBreak{
					EntityType:       "bank_transfer_order",
					EntityID:         o.ID.String(),
					LedgerTransferID: o.ReturnTransferID,
					LocalStatus:      o.Status,
					Reason:           "return_completed_locally_but_missing_in_ledger",
				})
			}
		}
	}

	now := time.Now().UTC()
	for _, b := range breaks {
		if err := queries.UpsertReconciliationBreak(ctx, sqlc.UpsertReconciliationBreakParams{
			ID:          uuid.New(),
			RunID:       runID,
			EntityType:  b.EntityType,
			EntityID:    b.EntityID,
			Reason:      b.Reason,
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			LocalStatus: pgtype.Text{String: b.LocalStatus, Valid: b.LocalStatus != ""},
			LedgerRef:   pgtype.Text{String: b.LedgerTransferID, Valid: b.LedgerTransferID != ""},
		}); err != nil {
			logger.Error("persist reconciliation break failed", "entity_type", b.EntityType, "entity_id", b.EntityID, "error", err)
			os.Exit(1)
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
		CheckedCount: int32(checkedCount),
		BreakCount:   int32(len(breaks)),
		Breaks:       breaksJSON,
		Status:       status,
	}); err != nil {
		logger.Error("finish reconciliation run failed", "error", err)
		os.Exit(1)
	}

	logger.Info("reconciliation complete", "run_id", runID, "checked", checkedCount, "breaks", len(breaks))
	if len(breaks) > 0 {
		for _, b := range breaks {
			logger.Warn("reconciliation break", "entity_type", b.EntityType, "entity_id", b.EntityID, "reason", b.Reason)
		}
		fmt.Fprintf(os.Stderr, "reconciliation found %d break(s), see payment.reconciliation_runs (id=%s) and payment.reconciliation_breaks\n", len(breaks), runID)
		os.Exit(1)
	}
}
