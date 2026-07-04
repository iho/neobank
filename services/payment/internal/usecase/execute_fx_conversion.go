package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/userclient"
	"github.com/iho/neobank/services/payment/internal/adapter/fxclient"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/port"
	"github.com/jackc/pgx/v5"
)

type ExecuteFXConversionInput struct {
	UserID  string
	QuoteID string
}

// ExecuteFXConversionUseCase locks in an fx simulator quote and moves funds
// between a user's two currency wallets. goledger rejects cross-currency
// transfers outright (ErrCurrencyMismatch), so a conversion is always two
// same-currency legs through a per-currency FX position account: the
// source currency leaves the user's wallet into that currency's position
// account, and the destination currency leaves the other position account
// into the user's wallet. The spread is not separated into its own fee
// posting in this phase — it stays implicitly in the position accounts'
// balances (see docs/vendor-simulators-plan.md Phase 4).
type ExecuteFXConversionUseCase struct {
	conversions   port.FXConversionRepository
	users         *userclient.Client
	ledger        *ledgerclient.Client
	fx            *fxclient.Client
	outbox        outbox.TxPublisher
	audit         audit.Recorder
	positionAccts map[string]string
	tx            *pgutil.TxRunner
}

func NewExecuteFXConversionUseCase(
	conversions port.FXConversionRepository,
	users *userclient.Client,
	ledger *ledgerclient.Client,
	fx *fxclient.Client,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	positionAccts map[string]string,
	tx *pgutil.TxRunner,
) *ExecuteFXConversionUseCase {
	return &ExecuteFXConversionUseCase{
		conversions:   conversions,
		users:         users,
		ledger:        ledger,
		fx:            fx,
		outbox:        outboxPublisher,
		audit:         auditRecorder,
		positionAccts: positionAccts,
		tx:            tx,
	}
}

func (uc *ExecuteFXConversionUseCase) Execute(ctx context.Context, in ExecuteFXConversionInput) (domain.FXConversion, error) {
	if in.UserID == "" || in.QuoteID == "" {
		return domain.FXConversion{}, fmt.Errorf("user_id and quote_id are required")
	}

	if existing, err := uc.conversions.GetByQuoteID(ctx, in.QuoteID); err != nil {
		return domain.FXConversion{}, err
	} else if existing != nil {
		return *existing, nil
	}

	if uc.fx == nil || uc.ledger == nil {
		return domain.FXConversion{}, fmt.Errorf("fx simulator or ledger unavailable")
	}

	quote, err := uc.fx.ExecuteQuote(ctx, in.QuoteID)
	if err != nil {
		return domain.FXConversion{}, fmt.Errorf("execute quote: %w", err)
	}

	fromAcct, err := uc.positionAccount(quote.FromCurrency)
	if err != nil {
		return domain.FXConversion{}, err
	}

	toAcct, err := uc.positionAccount(quote.ToCurrency)
	if err != nil {
		return domain.FXConversion{}, err
	}

	fromWallet, err := uc.users.GetWallet(ctx, in.UserID, quote.FromCurrency)
	if err != nil {
		return domain.FXConversion{}, fmt.Errorf("source wallet (open a %s wallet first): %w", quote.FromCurrency, err)
	}

	toWallet, err := uc.users.GetWallet(ctx, in.UserID, quote.ToCurrency)
	if err != nil {
		return domain.FXConversion{}, fmt.Errorf("destination wallet (open a %s wallet first): %w", quote.ToCurrency, err)
	}

	debitTransfer, err := uc.ledger.CreateTransfer(ctx, ledgerclient.CreateTransferInput{
		FromAccountID:  fromWallet.LedgerAccountID,
		ToAccountID:    fromAcct,
		Amount:         quote.Amount,
		IdempotencyKey: in.QuoteID + ":debit",
		Metadata: map[string]string{
			"quote_id": in.QuoteID,
			"type":     "fx_conversion_debit",
		},
	})
	if err != nil {
		return domain.FXConversion{}, fmt.Errorf("debit source wallet: %w", err)
	}

	creditTransfer, err := uc.ledger.CreateTransfer(ctx, ledgerclient.CreateTransferInput{
		FromAccountID:  toAcct,
		ToAccountID:    toWallet.LedgerAccountID,
		Amount:         quote.ConvertedAmount,
		IdempotencyKey: in.QuoteID + ":credit",
		Metadata: map[string]string{
			"quote_id": in.QuoteID,
			"type":     "fx_conversion_credit",
		},
	})
	if err != nil {
		return domain.FXConversion{}, fmt.Errorf("credit destination wallet: %w", err)
	}

	conversion := domain.FXConversion{
		QuoteID:              in.QuoteID,
		UserID:               in.UserID,
		FromCurrency:         quote.FromCurrency,
		ToCurrency:           quote.ToCurrency,
		Amount:               quote.Amount,
		ConvertedAmount:      quote.ConvertedAmount,
		Rate:                 quote.Rate,
		FromLedgerTransferID: debitTransfer.Id,
		ToLedgerTransferID:   creditTransfer.Id,
		Status:               "completed",
	}

	event := events.FXConversionCompleted{
		UserID:               in.UserID,
		QuoteID:              in.QuoteID,
		FromCurrency:         quote.FromCurrency,
		ToCurrency:           quote.ToCurrency,
		Amount:               quote.Amount,
		ConvertedAmount:      quote.ConvertedAmount,
		Rate:                 quote.Rate,
		FromLedgerTransferID: debitTransfer.Id,
		ToLedgerTransferID:   creditTransfer.Id,
	}

	var created domain.FXConversion
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var createErr error
		created, createErr = uc.conversions.WithTx(tx).Create(ctx, conversion)
		if createErr != nil {
			return createErr
		}

		event.ConversionID = created.ID

		if err := uc.outbox.WithTx(tx).Publish(ctx, event); err != nil {
			return err
		}

		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "fx_conversion",
			EntityID:   created.ID,
			Action:     "completed",
			ToStatus:   "completed",
			Metadata: map[string]any{
				"quote_id":         in.QuoteID,
				"user_id":          in.UserID,
				"from_currency":    quote.FromCurrency,
				"to_currency":      quote.ToCurrency,
				"amount":           quote.Amount,
				"converted_amount": quote.ConvertedAmount,
				"rate":             quote.Rate,
			},
		})
	}); err != nil {
		return domain.FXConversion{}, err
	}

	return created, nil
}

func (uc *ExecuteFXConversionUseCase) positionAccount(currency string) (string, error) {
	acct, ok := uc.positionAccts[currency]
	if !ok || acct == "" {
		return "", fmt.Errorf("fx position account not configured for %s", currency)
	}

	return acct, nil
}
