package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/payment/internal/adapter/fxclient"
)

type GetFXQuoteInput struct {
	FromCurrency string
	ToCurrency   string
	Amount       string
}

// GetFXQuoteUseCase prices a conversion by proxying to the fx simulator; the
// quote must be executed (ExecuteFXConversionUseCase) before it expires.
type GetFXQuoteUseCase struct {
	fx *fxclient.Client
}

func NewGetFXQuoteUseCase(fx *fxclient.Client) *GetFXQuoteUseCase {
	return &GetFXQuoteUseCase{fx: fx}
}

func (uc *GetFXQuoteUseCase) Execute(ctx context.Context, in GetFXQuoteInput) (fxclient.Quote, error) {
	if in.FromCurrency == "" || in.ToCurrency == "" || in.Amount == "" {
		return fxclient.Quote{}, fmt.Errorf("from_currency, to_currency, and amount are required")
	}

	if uc.fx == nil {
		return fxclient.Quote{}, fmt.Errorf("fx simulator unavailable")
	}

	return uc.fx.CreateQuote(ctx, in.FromCurrency, in.ToCurrency, in.Amount)
}
