package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/iho/neobank/services/simulators/fx/internal/domain"
)

type fakeQuoteRepository struct {
	quotes map[string]domain.Quote
	seq    int
}

func newFakeQuoteRepository() *fakeQuoteRepository {
	return &fakeQuoteRepository{quotes: map[string]domain.Quote{}}
}

func (r *fakeQuoteRepository) Create(_ context.Context, q domain.Quote) (domain.Quote, error) {
	r.seq++
	q.ID = "quote-" + strconv.Itoa(r.seq)
	r.quotes[q.ID] = q

	return q, nil
}

func (r *fakeQuoteRepository) GetByID(_ context.Context, id string) (*domain.Quote, error) {
	q, ok := r.quotes[id]
	if !ok {
		return nil, nil
	}

	return &q, nil
}

func (r *fakeQuoteRepository) MarkExecuted(_ context.Context, id string) (domain.Quote, error) {
	q, ok := r.quotes[id]
	if !ok {
		return domain.Quote{}, fmt.Errorf("quote %q not found", id)
	}

	q.Status = domain.QuoteStatusExecuted
	r.quotes[id] = q

	return q, nil
}
