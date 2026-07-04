package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/services/simulators/fx/internal/domain"
	"github.com/iho/neobank/services/simulators/fx/internal/port"
	"github.com/shopspring/decimal"
)

// DefaultSpreadBps is the retail markup applied to the mid rate: 50bps (0.5%).
const DefaultSpreadBps = 50

// DefaultQuoteTTL is how long a quote is valid before execution is refused.
const DefaultQuoteTTL = 30 * time.Second

type GetQuoteInput struct {
	FromCurrency string
	ToCurrency   string
	Amount       string
}

// GetQuoteUseCase prices a conversion: the caller shows this quote to the
// user, who must execute it (ExecuteQuoteUseCase) before DefaultQuoteTTL
// elapses.
type GetQuoteUseCase struct {
	quotes port.QuoteRepository
}

func NewGetQuoteUseCase(quotes port.QuoteRepository) *GetQuoteUseCase {
	return &GetQuoteUseCase{quotes: quotes}
}

func (uc *GetQuoteUseCase) Execute(ctx context.Context, in GetQuoteInput) (domain.Quote, error) {
	if in.FromCurrency == "" || in.ToCurrency == "" || in.Amount == "" {
		return domain.Quote{}, fmt.Errorf("from_currency, to_currency, and amount are required")
	}

	amount, err := decimal.NewFromString(in.Amount)
	if err != nil || amount.IsNegative() || amount.IsZero() {
		return domain.Quote{}, fmt.Errorf("invalid amount %q", in.Amount)
	}

	mid, err := MidRate(in.FromCurrency, in.ToCurrency, time.Now())
	if err != nil {
		return domain.Quote{}, err
	}

	rate := ApplyClientSpread(mid, DefaultSpreadBps)
	converted := amount.Mul(rate).Round(2)

	now := time.Now().UTC()

	return uc.quotes.Create(ctx, domain.Quote{
		FromCurrency:    in.FromCurrency,
		ToCurrency:      in.ToCurrency,
		Amount:          amount.Round(2).String(),
		ConvertedAmount: converted.String(),
		Rate:            rate.String(),
		SpreadBps:       DefaultSpreadBps,
		Status:          domain.QuoteStatusPending,
		ExpiresAt:       now.Add(DefaultQuoteTTL),
	})
}
