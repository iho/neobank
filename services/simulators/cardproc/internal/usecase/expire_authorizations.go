package usecase

import (
	"context"
	"time"

	"github.com/iho/neobank/services/simulators/cardproc/internal/port"
)

// EventAuthExpired is delivered when a hold ages out unused — a distinct
// event type from EventAuthReversed (a merchant explicitly voiding), even
// though the card service handles both the same way (releases the hold).
const EventAuthExpired = "card.auth.expired"

// ExpireAuthorizationsUseCase is the background sweep counterpart to
// ReverseTransactionUseCase: instead of an admin explicitly voiding a
// transaction, it finds holds that have sat in "approved" past a TTL with
// no capture or reversal, and expires them itself — matching how a real
// processor eventually releases an unused authorization.
type ExpireAuthorizationsUseCase struct {
	txs        port.TransactionRepository
	dispatcher WebhookDispatcher
	eventsURL  string
}

func NewExpireAuthorizationsUseCase(txs port.TransactionRepository, dispatcher WebhookDispatcher, eventsURL string) *ExpireAuthorizationsUseCase {
	return &ExpireAuthorizationsUseCase{txs: txs, dispatcher: dispatcher, eventsURL: eventsURL}
}

// Sweep expires every approved transaction older than ttl, returning how
// many it expired.
func (uc *ExpireAuthorizationsUseCase) Sweep(ctx context.Context, ttl time.Duration) (int, error) {
	const batchSize = 100

	cutoff := time.Now().UTC().Add(-ttl)

	txs, err := uc.txs.ListExpiredApproved(ctx, cutoff, batchSize)
	if err != nil {
		return 0, err
	}

	for _, tx := range txs {
		if _, err := uc.dispatcher.Enqueue(ctx, uc.eventsURL, EventAuthExpired, CardEventPayload{
			AuthorizationID: tx.AuthorizationID,
			Reason:          "expired",
		}); err != nil {
			return 0, err
		}

		if err := uc.txs.MarkExpired(ctx, tx.ID); err != nil {
			return 0, err
		}
	}

	return len(txs), nil
}
