//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package walletprojection

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/iho/neobank/pkg/events"
)

// Row is one persisted wallet transaction entry for a user.
type Row struct {
	CreatedAt     time.Time
	UserID        string
	ID            string
	SourceEventID string
	Type          string
	Amount        string
	Currency      string
	Direction     string
	Status        string
	Counterparty  string
	Memo          string
}

// Apply derives read-model rows or updates from a domain event envelope.
// Returns nil when the event type is not part of the wallet history projection.
func Apply(envelope events.Envelope) ([]Row, *CaptureUpdate, error) {
	switch envelope.EventType {
	case events.TypeTransferCompleted:
		return applyTransferCompleted(envelope)
	case events.TypeBankTransferReceived:
		return applyBankTransferReceived(envelope)
	case events.TypeFXConversionCompleted:
		return applyFXConversionCompleted(envelope)
	case events.TypeBankTransferSent:
		return applyBankTransferSent(envelope)
	case events.TypeBankTransferReturned:
		update, err := applyBankTransferReturned(envelope)
		return nil, update, err
	case events.TypeCardAuthApproved:
		return applyCardAuthApproved(envelope)
	case events.TypeCardAuthCaptured:
		update, err := applyCardAuthCaptured(envelope)
		return nil, update, err
	case events.TypeCardAuthVoided:
		update, err := applyCardAuthVoided(envelope)
		return nil, update, err
	case events.TypeCardChargebackOpened:
		return applyCardChargebackOpened(envelope)
	case events.TypeCardChargebackResolved:
		update, err := applyCardChargebackResolved(envelope)
		return nil, update, err

	default:
		return nil, nil, nil
	}
}

// CaptureUpdate mutates an existing card authorization row to captured.
type CaptureUpdate struct {
	CreatedAt     time.Time
	UserID        string
	ID            string
	SourceEventID string
	Type          string
	Amount        string
	Currency      string
	Status        string
}

func applyTransferCompleted(envelope events.Envelope) ([]Row, *CaptureUpdate, error) {
	var payload events.TransferCompleted
	err := json.Unmarshal(envelope.Payload, &payload)
	if err != nil {
		return nil, nil, fmt.Errorf("parse transfer completed: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	senderCounterparty := payload.RecipientDisplayName
	if senderCounterparty == "" {
		senderCounterparty = payload.RecipientUserID
	}
	recipientCounterparty := payload.SenderDisplayName
	if recipientCounterparty == "" {
		recipientCounterparty = payload.SenderUserID
	}

	rows := []Row{
		{
			UserID:        payload.SenderUserID,
			ID:            payload.TransferID,
			SourceEventID: envelope.EventID,
			Type:          "p2p_out",
			Amount:        payload.Amount,
			Currency:      payload.Currency,
			Direction:     "debit",
			Status:        "completed",
			Counterparty:  senderCounterparty,
			Memo:          payload.Memo,
			CreatedAt:     createdAt,
		},
		{
			UserID:        payload.RecipientUserID,
			ID:            payload.TransferID,
			SourceEventID: envelope.EventID,
			Type:          "p2p_in",
			Amount:        payload.Amount,
			Currency:      payload.Currency,
			Direction:     "credit",
			Status:        "completed",
			Counterparty:  recipientCounterparty,
			Memo:          payload.Memo,
			CreatedAt:     createdAt,
		},
	}

	return rows, nil, nil
}

func applyBankTransferReceived(envelope events.Envelope) ([]Row, *CaptureUpdate, error) {
	var payload events.BankTransferReceived
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return nil, nil, fmt.Errorf("parse bank transfer received: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	counterparty := payload.SenderName
	if counterparty == "" {
		counterparty = "Incoming bank transfer"
	}

	return []Row{{
		UserID:        payload.UserID,
		ID:            payload.TransferID,
		SourceEventID: envelope.EventID,
		Type:          "bank_transfer_in",
		Amount:        payload.Amount,
		Currency:      payload.Currency,
		Direction:     "credit",
		Status:        "completed",
		Counterparty:  counterparty,
		Memo:          payload.Reference,
		CreatedAt:     createdAt,
	}}, nil, nil
}

func applyBankTransferSent(envelope events.Envelope) ([]Row, *CaptureUpdate, error) {
	var payload events.BankTransferSent
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return nil, nil, fmt.Errorf("parse bank transfer sent: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return []Row{{
		UserID:        payload.UserID,
		ID:            payload.OrderID,
		SourceEventID: envelope.EventID,
		Type:          "bank_transfer_out",
		Amount:        payload.Amount,
		Currency:      payload.Currency,
		Direction:     "debit",
		Status:        "processing",
		Counterparty:  payload.CounterpartyIBAN,
		Memo:          payload.Reference,
		CreatedAt:     createdAt,
	}}, nil, nil
}

// applyBankTransferReturned updates the row applyBankTransferSent created
// (same ID) rather than adding a new one — the ledger transfer already
// reversed the debit, so the wallet balance is correct; this just reflects
// that the transfer didn't go through.
func applyBankTransferReturned(envelope events.Envelope) (*CaptureUpdate, error) {
	var payload events.BankTransferReturned
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return nil, fmt.Errorf("parse bank transfer returned: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return &CaptureUpdate{
		UserID:        payload.UserID,
		ID:            payload.OrderID,
		SourceEventID: envelope.EventID,
		Type:          "bank_transfer_out",
		Amount:        payload.Amount,
		Currency:      payload.Currency,
		Status:        "returned",
		CreatedAt:     createdAt,
	}, nil
}

func applyFXConversionCompleted(envelope events.Envelope) ([]Row, *CaptureUpdate, error) {
	var payload events.FXConversionCompleted
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return nil, nil, fmt.Errorf("parse fx conversion completed: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	// Both rows belong to the same user, so they need distinct IDs — the
	// wallet_transactions table is unique on (user_id, id).
	rows := []Row{
		{
			UserID:        payload.UserID,
			ID:            payload.ConversionID + "-debit",
			SourceEventID: envelope.EventID,
			Type:          "fx_conversion_out",
			Amount:        payload.Amount,
			Currency:      payload.FromCurrency,
			Direction:     "debit",
			Status:        "completed",
			Counterparty:  "FX conversion",
			Memo:          fmt.Sprintf("Converted to %s at %s", payload.ToCurrency, payload.Rate),
			CreatedAt:     createdAt,
		},
		{
			UserID:        payload.UserID,
			ID:            payload.ConversionID + "-credit",
			SourceEventID: envelope.EventID,
			Type:          "fx_conversion_in",
			Amount:        payload.ConvertedAmount,
			Currency:      payload.ToCurrency,
			Direction:     "credit",
			Status:        "completed",
			Counterparty:  "FX conversion",
			Memo:          fmt.Sprintf("Converted from %s at %s", payload.FromCurrency, payload.Rate),
			CreatedAt:     createdAt,
		},
	}

	return rows, nil, nil
}

func applyCardAuthApproved(envelope events.Envelope) ([]Row, *CaptureUpdate, error) {
	var payload events.CardAuthApproved
	err := json.Unmarshal(envelope.Payload, &payload)
	if err != nil {
		return nil, nil, fmt.Errorf("parse card auth approved: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return []Row{{
		UserID:        payload.UserID,
		ID:            payload.AuthorizationID,
		SourceEventID: envelope.EventID,
		Type:          "card_hold",
		Amount:        payload.Amount,
		Currency:      payload.Currency,
		Direction:     "debit",
		Status:        "authorized",
		Counterparty:  payload.MerchantName,
		CreatedAt:     createdAt,
	}}, nil, nil
}

func applyCardAuthVoided(envelope events.Envelope) (*CaptureUpdate, error) {
	var payload events.CardAuthVoided
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return nil, fmt.Errorf("parse card auth voided: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return &CaptureUpdate{
		UserID:        payload.UserID,
		ID:            payload.AuthorizationID,
		SourceEventID: envelope.EventID,
		Type:          "card_hold_released",
		Amount:        payload.Amount,
		Currency:      payload.Currency,
		Status:        "voided",
		CreatedAt:     createdAt,
	}, nil
}

func applyCardAuthCaptured(envelope events.Envelope) (*CaptureUpdate, error) {
	var payload events.CardAuthCaptured
	err := json.Unmarshal(envelope.Payload, &payload)
	if err != nil {
		return nil, fmt.Errorf("parse card auth captured: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return &CaptureUpdate{
		UserID:        payload.UserID,
		ID:            payload.AuthorizationID,
		SourceEventID: envelope.EventID,
		Type:          "card_purchase",
		Amount:        payload.Amount,
		Currency:      payload.Currency,
		Status:        "captured",
		CreatedAt:     createdAt,
	}, nil
}

func applyCardChargebackOpened(envelope events.Envelope) ([]Row, *CaptureUpdate, error) {
	var payload events.CardChargebackOpened
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return nil, nil, fmt.Errorf("parse card chargeback opened: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return []Row{{
		UserID:        payload.UserID,
		ID:            payload.DisputeID,
		SourceEventID: envelope.EventID,
		Type:          "chargeback_credit",
		Amount:        payload.Amount,
		Currency:      payload.Currency,
		Direction:     "credit",
		Status:        "provisional",
		CreatedAt:     createdAt,
	}}, nil, nil
}

// applyCardChargebackResolved reuses the chargeback-opened row's ID: the
// ledger transfer (if any) already moved the money, this only needs to
// reflect the dispute's final state ("won" the credit stays, "lost" it was
// clawed back) — same pattern as applyBankTransferReturned.
func applyCardChargebackResolved(envelope events.Envelope) (*CaptureUpdate, error) {
	var payload events.CardChargebackResolved
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return nil, fmt.Errorf("parse card chargeback resolved: %w", err)
	}

	createdAt := envelope.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return &CaptureUpdate{
		UserID:        payload.UserID,
		ID:            payload.DisputeID,
		SourceEventID: envelope.EventID,
		Type:          "chargeback_credit",
		Amount:        payload.Amount,
		Currency:      payload.Currency,
		Status:        payload.Outcome,
		CreatedAt:     createdAt,
	}, nil
}
