// reconcile compares card.authorizations against goledger holds/transfers
// and records any drift so operators and auditors have a durable answer to
// "did we check, and what did we find". Intended to run on a schedule
// (cron/k8s CronJob), not as a long-lived process.
//
// goledger only exposes ListHoldsByAccount (no GetHold-by-ID), so each
// authorization's ledger account is resolved via the user service and
// cached per user+currency for the run.
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
	"github.com/iho/neobank/pkg/userclient"
	"github.com/iho/neobank/services/card/internal/config"
	"github.com/iho/neobank/services/card/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type reconciliationBreak struct {
	AuthorizationID string `json:"authorization_id"`
	LedgerHoldID    string `json:"ledger_hold_id"`
	LocalStatus     string `json:"local_status"`
	Reason          string `json:"reason"`
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

	users := userclient.New(cfg.UserURL)
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

	auths, err := queries.ListAuthorizationsForReconciliation(ctx, batchSize)
	if err != nil {
		logger.Error("list authorizations failed", "error", err)
		os.Exit(1)
	}

	type walletKey struct{ userID, currency string }
	walletCache := map[walletKey]string{}

	breaks := []reconciliationBreak{}
	checked := 0
	for _, a := range auths {
		if !a.LedgerHoldID.Valid || a.LedgerHoldID.String == "" {
			continue
		}
		key := walletKey{userID: a.UserID.String(), currency: a.Currency}
		ledgerAccountID, ok := walletCache[key]
		if !ok {
			wallet, err := users.GetWallet(ctx, key.userID, key.currency)
			if err != nil {
				logger.Warn("wallet lookup failed, skipping", "authorization_id", a.ID, "error", err)
				continue
			}
			ledgerAccountID = wallet.LedgerAccountID
			walletCache[key] = ledgerAccountID
		}

		holds, err := ledger.ListHoldsByAccount(ctx, ledgerAccountID, 500)
		if err != nil {
			logger.Warn("ledger hold lookup failed, skipping", "authorization_id", a.ID, "error", err)
			continue
		}
		checked++

		found := false
		for _, h := range holds {
			if h.Id == a.LedgerHoldID.String {
				found = true
				break
			}
		}
		if !found {
			breaks = append(breaks, reconciliationBreak{
				AuthorizationID: a.ID.String(),
				LedgerHoldID:    a.LedgerHoldID.String,
				LocalStatus:     a.Status,
				Reason:          "hold_missing_in_ledger",
			})
			continue
		}

		if a.Status == "captured" {
			if !a.LedgerTransferID.Valid || a.LedgerTransferID.String == "" {
				breaks = append(breaks, reconciliationBreak{
					AuthorizationID: a.ID.String(),
					LedgerHoldID:    a.LedgerHoldID.String,
					LocalStatus:     a.Status,
					Reason:          "captured_locally_but_no_ledger_transfer_recorded",
				})
				continue
			}
			transfer, err := ledger.GetTransfer(ctx, a.LedgerTransferID.String)
			if err != nil {
				logger.Warn("ledger transfer lookup failed, skipping", "authorization_id", a.ID, "error", err)
				continue
			}
			if transfer == nil {
				breaks = append(breaks, reconciliationBreak{
					AuthorizationID: a.ID.String(),
					LedgerHoldID:    a.LedgerHoldID.String,
					LocalStatus:     a.Status,
					Reason:          "captured_locally_but_missing_capture_transfer_in_ledger",
				})
			}
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
		CheckedCount: int32(checked),
		BreakCount:   int32(len(breaks)),
		Breaks:       breaksJSON,
		Status:       status,
	}); err != nil {
		logger.Error("finish reconciliation run failed", "error", err)
		os.Exit(1)
	}

	logger.Info("reconciliation complete", "run_id", runID, "checked", checked, "breaks", len(breaks))
	if len(breaks) > 0 {
		for _, b := range breaks {
			logger.Warn("reconciliation break", "authorization_id", b.AuthorizationID, "reason", b.Reason)
		}
		fmt.Fprintf(os.Stderr, "reconciliation found %d break(s), see card.reconciliation_runs (id=%s)\n", len(breaks), runID)
		os.Exit(1)
	}
}
