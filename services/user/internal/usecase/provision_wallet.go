package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/saga"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/port"
	"github.com/jackc/pgx/v5"
)

type LedgerClient interface {
	CreateAccount(ctx context.Context, in ledgerclient.CreateAccountInput) (*goledgerv1.Account, error)
}

type ProvisionWalletInput struct {
	UserID         string
	Currency       string
	IdempotencyKey string
}

type ProvisionWalletOutput struct {
	WalletID        string
	LedgerAccountID string
	Created         bool
}

type ProvisionWalletUseCase struct {
	wallets   port.WalletRepository
	ledger    LedgerClient
	outbox    outbox.TxPublisher
	audit     audit.Recorder
	sagaStore saga.InstanceStore
	tx        *pgutil.TxRunner
}

func NewProvisionWalletUseCase(
	wallets port.WalletRepository,
	ledger LedgerClient,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	sagaStore saga.InstanceStore,
	tx *pgutil.TxRunner,
) *ProvisionWalletUseCase {
	return &ProvisionWalletUseCase{
		wallets:   wallets,
		ledger:    ledger,
		outbox:    outboxPublisher,
		audit:     auditRecorder,
		sagaStore: sagaStore,
		tx:        tx,
	}
}

func (uc *ProvisionWalletUseCase) Execute(ctx context.Context, in ProvisionWalletInput) (ProvisionWalletOutput, error) {
	if in.UserID == "" {
		return ProvisionWalletOutput{}, fmt.Errorf("user_id is required")
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}
	if in.IdempotencyKey == "" {
		in.IdempotencyKey = fmt.Sprintf("wallet:%s:%s", in.UserID, in.Currency)
	}

	existing, err := uc.wallets.GetByUserAndCurrency(ctx, in.UserID, in.Currency)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ProvisionWalletOutput{}, err
	}
	if existing != nil {
		return ProvisionWalletOutput{
			WalletID:        existing.ID,
			LedgerAccountID: existing.LedgerAccountID,
			Created:         false,
		}, nil
	}

	walletID := uuid.NewString()
	state := saga.NewState(map[string]string{
		"user_id":         in.UserID,
		"currency":        in.Currency,
		"wallet_id":       walletID,
		"idempotency_key": in.IdempotencyKey,
	})

	orchestrator := saga.New("wallet_provision", uc.steps(), uc.sagaStore)
	if err := orchestrator.Run(ctx, in.IdempotencyKey, state); err != nil {
		return ProvisionWalletOutput{}, err
	}

	wallet, err := uc.wallets.GetByUserAndCurrency(ctx, in.UserID, in.Currency)
	if err != nil {
		return ProvisionWalletOutput{}, err
	}

	return ProvisionWalletOutput{
		WalletID:        wallet.ID,
		LedgerAccountID: wallet.LedgerAccountID,
		Created:         true,
	}, nil
}

func (uc *ProvisionWalletUseCase) steps() []saga.Step {
	return []saga.Step{
		{
			Name: "create_ledger_account",
			Execute: func(ctx context.Context, state *saga.State) error {
				if uc.ledger == nil {
					return fmt.Errorf("ledger unavailable")
				}
				account, err := uc.ledger.CreateAccount(ctx, ledgerclient.CreateAccountInput{
					Name:                 fmt.Sprintf("CUSTOMER:%s:%s", state.Get("user_id"), state.Get("currency")),
					Currency:             state.Get("currency"),
					AllowNegativeBalance: false,
					AllowPositiveBalance: true,
				})
				if err != nil {
					return fmt.Errorf("create ledger account: %w", err)
				}
				state.Set("ledger_account_id", account.Id)
				return nil
			},
			// goledger has no account deletion API; orphan ledger accounts require manual reconciliation.
		},
		{
			Name: "persist_wallet",
			Execute: func(ctx context.Context, state *saga.State) error {
				existing, err := uc.wallets.GetByUserAndCurrency(ctx, state.Get("user_id"), state.Get("currency"))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					return err
				}
				if existing != nil {
					return nil
				}

				wallet := domain.Wallet{
					ID:              state.Get("wallet_id"),
					UserID:          state.Get("user_id"),
					Currency:        state.Get("currency"),
					LedgerAccountID: state.Get("ledger_account_id"),
					Status:          "active",
				}
				event := events.WalletProvisioned{
					UserID:          state.Get("user_id"),
					WalletID:        state.Get("wallet_id"),
					LedgerAccountID: state.Get("ledger_account_id"),
					Currency:        state.Get("currency"),
				}
				return uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
					if err := uc.wallets.WithTx(tx).Create(ctx, wallet); err != nil {
						return err
					}
					if err := uc.audit.WithTx(tx).Record(ctx, audit.Entry{
						EntityType: "wallet",
						EntityID:   state.Get("wallet_id"),
						Action:     "provisioned",
						ToStatus:   "active",
						Metadata:   map[string]any{"currency": state.Get("currency"), "ledger_account_id": state.Get("ledger_account_id")},
					}); err != nil {
						return err
					}
					return uc.outbox.WithTx(tx).Publish(ctx, event)
				})
			},
			Compensate: func(ctx context.Context, state *saga.State) error {
				if !state.Has("wallet_id") {
					return nil
				}
				return uc.wallets.DeleteByID(ctx, state.Get("wallet_id"))
			},
		},
	}
}
