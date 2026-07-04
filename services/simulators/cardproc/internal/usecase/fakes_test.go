package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
)

type fakeCardRepository struct {
	cards map[string]domain.Card
	seq   int
}

func newFakeCardRepository() *fakeCardRepository {
	return &fakeCardRepository{cards: map[string]domain.Card{}}
}

func (r *fakeCardRepository) Create(_ context.Context, externalRef, cardholderName, panToken, lastFour string, expiryMonth, expiryYear int) (domain.Card, error) {
	r.seq++
	card := domain.Card{
		ID:             "card-" + strconv.Itoa(r.seq),
		ExternalRef:    externalRef,
		CardholderName: cardholderName,
		PANToken:       panToken,
		LastFour:       lastFour,
		ExpiryMonth:    expiryMonth,
		ExpiryYear:     expiryYear,
		Status:         "active",
	}
	r.cards[card.ID] = card

	return card, nil
}

func (r *fakeCardRepository) GetByID(_ context.Context, id string) (*domain.Card, error) {
	card, ok := r.cards[id]
	if !ok {
		return nil, nil
	}

	return &card, nil
}

func (r *fakeCardRepository) Cancel(_ context.Context, id string) error {
	card, ok := r.cards[id]
	if !ok {
		return fmt.Errorf("card %q not found", id)
	}

	card.Status = "cancelled"
	r.cards[id] = card

	return nil
}

type fakeTransactionRepository struct {
	txs map[string]domain.Transaction
	seq int
}

func newFakeTransactionRepository() *fakeTransactionRepository {
	return &fakeTransactionRepository{txs: map[string]domain.Transaction{}}
}

func (r *fakeTransactionRepository) Create(_ context.Context, cardID, amount, currency, merchantName, mcc string) (domain.Transaction, error) {
	r.seq++
	tx := domain.Transaction{
		ID:           "tx-" + strconv.Itoa(r.seq),
		CardID:       cardID,
		Amount:       amount,
		Currency:     currency,
		MerchantName: merchantName,
		MCC:          mcc,
		Status:       "pending",
	}
	r.txs[tx.ID] = tx

	return tx, nil
}

func (r *fakeTransactionRepository) GetByID(_ context.Context, id string) (*domain.Transaction, error) {
	tx, ok := r.txs[id]
	if !ok {
		return nil, nil
	}

	return &tx, nil
}

func (r *fakeTransactionRepository) SetAuthResult(_ context.Context, id, status, authorizationID, reasonCode string) (domain.Transaction, error) {
	tx, ok := r.txs[id]
	if !ok {
		return domain.Transaction{}, fmt.Errorf("transaction %q not found", id)
	}

	tx.Status = status
	tx.AuthorizationID = authorizationID
	tx.ReasonCode = reasonCode
	r.txs[id] = tx

	return tx, nil
}

func (r *fakeTransactionRepository) MarkCaptured(_ context.Context, id string) error {
	tx, ok := r.txs[id]
	if !ok {
		return fmt.Errorf("transaction %q not found", id)
	}

	tx.Status = domain.TransactionStatusCaptured
	r.txs[id] = tx

	return nil
}

func (r *fakeTransactionRepository) MarkReversed(_ context.Context, id string) error {
	tx, ok := r.txs[id]
	if !ok {
		return fmt.Errorf("transaction %q not found", id)
	}

	tx.Status = domain.TransactionStatusReversed
	r.txs[id] = tx

	return nil
}

func (r *fakeTransactionRepository) MarkExpired(_ context.Context, id string) error {
	tx, ok := r.txs[id]
	if !ok {
		return fmt.Errorf("transaction %q not found", id)
	}

	tx.Status = domain.TransactionStatusExpired
	r.txs[id] = tx

	return nil
}

func (r *fakeTransactionRepository) ListExpiredApproved(_ context.Context, cutoff time.Time, limit int32) ([]domain.Transaction, error) {
	var out []domain.Transaction
	for _, tx := range r.txs {
		if tx.Status != domain.TransactionStatusApproved || !tx.CreatedAt.Before(cutoff) {
			continue
		}
		out = append(out, tx)
		if int32(len(out)) >= limit {
			break
		}
	}

	return out, nil
}

type fakeChargebackRepository struct {
	chargebacks map[string]domain.Chargeback
	seq         int
}

func newFakeChargebackRepository() *fakeChargebackRepository {
	return &fakeChargebackRepository{chargebacks: map[string]domain.Chargeback{}}
}

func (r *fakeChargebackRepository) Create(_ context.Context, transactionID, authorizationID, amount, currency, reason string) (domain.Chargeback, error) {
	r.seq++
	cb := domain.Chargeback{
		ID:              "chargeback-" + strconv.Itoa(r.seq),
		TransactionID:   transactionID,
		AuthorizationID: authorizationID,
		Amount:          amount,
		Currency:        currency,
		Reason:          reason,
		Status:          domain.ChargebackStatusOpened,
	}
	r.chargebacks[cb.ID] = cb

	return cb, nil
}

func (r *fakeChargebackRepository) GetByID(_ context.Context, id string) (*domain.Chargeback, error) {
	cb, ok := r.chargebacks[id]
	if !ok {
		return nil, nil
	}

	return &cb, nil
}

func (r *fakeChargebackRepository) SetStatus(_ context.Context, id, status string) (domain.Chargeback, error) {
	cb, ok := r.chargebacks[id]
	if !ok {
		return domain.Chargeback{}, fmt.Errorf("chargeback %q not found", id)
	}

	cb.Status = status
	r.chargebacks[id] = cb

	return cb, nil
}

type fakeDispatcher struct {
	calls []struct {
		url       string
		eventType string
		payload   any
	}
}

func (d *fakeDispatcher) Enqueue(_ context.Context, url, eventType string, payload any) (string, error) {
	d.calls = append(d.calls, struct {
		url       string
		eventType string
		payload   any
	}{url, eventType, payload})

	return "delivery-1", nil
}
