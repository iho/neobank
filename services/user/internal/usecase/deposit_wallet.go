package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/money"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/walletprojection"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/port"
	"github.com/jackc/pgx/v5"
)

type DepositWalletInput struct {
	UserID         string
	Amount         string
	Currency       string
	IdempotencyKey string
}

type DepositLedgerClient interface {
	CreateTransfer(ctx context.Context, in ledgerclient.CreateTransferInput) (*goledgerv1.Transfer, error)
}

type DepositWalletUseCase struct {
	wallets       port.WalletRepository
	deposits      port.DepositRepository
	walletTx      WalletTransactionRepository
	outbox        outbox.TxPublisher
	audit         audit.Recorder
	ledger        DepositLedgerClient
	sourceAccount string
	maxAmount     string
	tx            *pgutil.TxRunner
}

func NewDepositWalletUseCase(
	wallets port.WalletRepository,
	deposits port.DepositRepository,
	walletTx WalletTransactionRepository,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	ledger DepositLedgerClient,
	sourceAccount, maxAmount string,
	tx *pgutil.TxRunner,
) *DepositWalletUseCase {
	return &DepositWalletUseCase{
		wallets:       wallets,
		deposits:      deposits,
		walletTx:      walletTx,
		outbox:        outboxPublisher,
		audit:         auditRecorder,
		ledger:        ledger,
		sourceAccount: sourceAccount,
		maxAmount:     maxAmount,
		tx:            tx,
	}
}

type DepositWalletOutput struct {
	Deposit domain.Deposit
	Created bool
}

func (uc *DepositWalletUseCase) Execute(ctx context.Context, in DepositWalletInput) (DepositWalletOutput, error) {
	if in.UserID == "" || in.IdempotencyKey == "" {
		return DepositWalletOutput{}, fmt.Errorf("user_id and idempotency_key are required")
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}
	if _, err := money.Parse(in.Amount); err != nil {
		return DepositWalletOutput{}, err
	}
	if uc.maxAmount != "" {
		max, err := money.Parse(uc.maxAmount)
		if err != nil {
			return DepositWalletOutput{}, fmt.Errorf("invalid max deposit amount config: %w", err)
		}
		amt, _ := money.Parse(in.Amount)
		if amt.GreaterThan(max) {
			return DepositWalletOutput{}, fmt.Errorf("deposit amount exceeds maximum %s", uc.maxAmount)
		}
	}
	if uc.sourceAccount == "" {
		return DepositWalletOutput{}, fmt.Errorf("deposits are not configured")
	}
	if uc.ledger == nil {
		return DepositWalletOutput{}, fmt.Errorf("ledger unavailable")
	}

	existing, err := uc.deposits.GetByUserAndIdempotencyKey(ctx, in.UserID, in.IdempotencyKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return DepositWalletOutput{}, err
	}
	if existing != nil {
		return DepositWalletOutput{Deposit: *existing, Created: false}, nil
	}

	wallet, err := uc.wallets.GetByUserAndCurrency(ctx, in.UserID, in.Currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DepositWalletOutput{}, fmt.Errorf("wallet not found")
		}
		return DepositWalletOutput{}, err
	}

	depositID := uuid.NewString()
	ledgerTransfer, err := uc.ledger.CreateTransfer(ctx, ledgerclient.CreateTransferInput{
		FromAccountID:  uc.sourceAccount,
		ToAccountID:    wallet.LedgerAccountID,
		Amount:         in.Amount,
		IdempotencyKey: in.IdempotencyKey,
		Metadata: map[string]string{
			"deposit_id": depositID,
			"type":       "deposit",
		},
	})
	if err != nil {
		return DepositWalletOutput{}, err
	}

	now := time.Now().UTC()
	completed := now
	deposit := domain.Deposit{
		ID:               depositID,
		UserID:           in.UserID,
		WalletID:         wallet.ID,
		Amount:           in.Amount,
		Currency:         in.Currency,
		LedgerTransferID: ledgerTransfer.Id,
		Status:           domain.DepositStatusCompleted,
		IdempotencyKey:   in.IdempotencyKey,
		CreatedAt:        now,
		CompletedAt:      &completed,
	}

	event := events.DepositCompleted{
		DepositID:        depositID,
		UserID:           in.UserID,
		WalletID:         wallet.ID,
		LedgerTransferID: ledgerTransfer.Id,
		Amount:           in.Amount,
		Currency:         in.Currency,
	}

	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		depositRepo := uc.deposits.WithTx(tx)
		if err := depositRepo.Insert(ctx, deposit); err != nil {
			return err
		}
		if err := uc.walletTx.Insert(ctx, walletprojection.Row{
			ID:            depositID,
			UserID:        in.UserID,
			SourceEventID: depositID,
			Type:          "deposit",
			Amount:        in.Amount,
			Currency:      in.Currency,
			Direction:     "credit",
			Status:        "completed",
			Counterparty:  "Simulated deposit",
			Memo:          "Top-up",
			CreatedAt:     now,
		}); err != nil {
			return err
		}
		if err := uc.outbox.WithTx(tx).Publish(ctx, event); err != nil {
			return err
		}
		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "deposit",
			EntityID:   depositID,
			Action:     "completed",
			ToStatus:   string(domain.DepositStatusCompleted),
			Metadata: map[string]any{
				"amount":             in.Amount,
				"currency":           in.Currency,
				"ledger_transfer_id": ledgerTransfer.Id,
			},
		})
	}); err != nil {
		return DepositWalletOutput{}, err
	}

	return DepositWalletOutput{Deposit: deposit, Created: true}, nil
}