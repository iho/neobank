package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/services/simulators/fx/internal/domain"
	"github.com/iho/neobank/services/simulators/fx/internal/port"
)

// ExecuteQuoteUseCase locks in a previously priced quote. Executing an
// already-executed quote is a no-op returning the same result (the caller,
// e.g. a payment service retrying a webhook-free HTTP call, may call this
// more than once); executing an expired quote is refused.
type ExecuteQuoteUseCase struct {
	quotes port.QuoteRepository
}

func NewExecuteQuoteUseCase(quotes port.QuoteRepository) *ExecuteQuoteUseCase {
	return &ExecuteQuoteUseCase{quotes: quotes}
}

func (uc *ExecuteQuoteUseCase) Execute(ctx context.Context, quoteID string) (domain.Quote, error) {
	if quoteID == "" {
		return domain.Quote{}, fmt.Errorf("quote_id is required")
	}

	quote, err := uc.quotes.GetByID(ctx, quoteID)
	if err != nil {
		return domain.Quote{}, err
	}

	if quote == nil {
		return domain.Quote{}, fmt.Errorf("quote %q not found", quoteID)
	}

	if quote.Status == domain.QuoteStatusExecuted {
		return *quote, nil
	}

	if time.Now().UTC().After(quote.ExpiresAt) {
		return domain.Quote{}, fmt.Errorf("quote_expired")
	}

	return uc.quotes.MarkExecuted(ctx, quote.ID)
}
