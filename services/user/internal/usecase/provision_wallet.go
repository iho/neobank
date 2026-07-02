package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/saga"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet domain.Wallet) error
	DeleteByID(ctx context.Context, walletID string) error
	GetByUserAndCurrency(ctx context.Context, userID, currency string) (*domain.Wallet, error)
}

type LedgerClient interface {
	CreateAccount(ctx context.Context, in ledgerclient.CreateAccountInput) (*goledgerv1.Account, error)
}

type OutboxPublisher interface {
	Publish(ctx context.Context, evt events.Event) error
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
	wallets   WalletRepository
	ledger    LedgerClient
	outbox    OutboxPublisher
	sagaStore saga.InstanceStore
}

func NewProvisionWalletUseCase(
	wallets WalletRepository,
	ledger LedgerClient,
	outbox OutboxPublisher,
	sagaStore saga.InstanceStore,
) *ProvisionWalletUseCase {
	return &ProvisionWalletUseCase{
		wallets:   wallets,
		ledger:    ledger,
		outbox:    outbox,
		sagaStore: sagaStore,
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

	_ = uc.outbox.Publish(ctx, events.WalletProvisioned{
		UserID:          in.UserID,
		WalletID:        wallet.ID,
		LedgerAccountID: wallet.LedgerAccountID,
		Currency:        in.Currency,
	})

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
				wallet := domain.Wallet{
					ID:              state.Get("wallet_id"),
					UserID:          state.Get("user_id"),
					Currency:        state.Get("currency"),
					LedgerAccountID: state.Get("ledger_account_id"),
					Status:          "active",
				}
				return uc.wallets.Create(ctx, wallet)
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