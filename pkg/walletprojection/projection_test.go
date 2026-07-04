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
	"testing"
	"time"

	"github.com/iho/neobank/pkg/events"
)

func TestApplyTransferCompleted(t *testing.T) {
	payload, _ := json.Marshal(events.TransferCompleted{
		TransferID: "t1", SenderUserID: "u1", RecipientUserID: "u2",
		Amount: "10.00", Currency: "USD", Memo: "thanks",
		SenderDisplayName: "sender@example.com", RecipientDisplayName: "recipient@example.com",
	})
	occurred := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)

	rows, update, err := Apply(events.Envelope{
		EventID: "e1", EventType: events.TypeTransferCompleted, OccurredAt: occurred, Payload: payload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if update != nil {
		t.Fatal("expected no capture update")
	}

	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}

	if rows[0].Type != "p2p_out" || rows[0].UserID != "u1" || rows[0].Counterparty != "recipient@example.com" || rows[0].Memo != "thanks" {
		t.Fatalf("sender row = %+v", rows[0])
	}

	if rows[1].Type != "p2p_in" || rows[1].UserID != "u2" || rows[1].Counterparty != "sender@example.com" {
		t.Fatalf("recipient row = %+v", rows[1])
	}
}

func TestApplyBankTransferReceived(t *testing.T) {
	payload, _ := json.Marshal(events.BankTransferReceived{
		TransferID: "bt1", UserID: "u1", LedgerTransferID: "ltx1",
		Amount: "250.00", Currency: "USD", SenderName: "Jane Doe", Reference: "rent",
	})
	occurred := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)

	rows, update, err := Apply(events.Envelope{
		EventID: "e4", EventType: events.TypeBankTransferReceived, OccurredAt: occurred, Payload: payload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if update != nil {
		t.Fatal("expected no capture update")
	}

	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}

	row := rows[0]
	if row.Type != "bank_transfer_in" || row.UserID != "u1" || row.Direction != "credit" ||
		row.Counterparty != "Jane Doe" || row.Memo != "rent" || row.Amount != "250.00" {
		t.Fatalf("bank transfer row = %+v", row)
	}
}

func TestApplyFXConversionCompleted(t *testing.T) {
	payload, _ := json.Marshal(events.FXConversionCompleted{
		ConversionID: "fxc1", UserID: "u1", QuoteID: "q1",
		FromCurrency: "EUR", ToCurrency: "USD", Amount: "100.00", ConvertedAmount: "107.46",
		Rate: "1.0746", FromLedgerTransferID: "ltx1", ToLedgerTransferID: "ltx2",
	})
	occurred := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)

	rows, update, err := Apply(events.Envelope{
		EventID: "e5", EventType: events.TypeFXConversionCompleted, OccurredAt: occurred, Payload: payload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if update != nil {
		t.Fatal("expected no capture update")
	}

	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}

	if rows[0].ID == rows[1].ID {
		t.Fatalf("expected distinct row IDs for same-user rows, got %q twice", rows[0].ID)
	}

	if rows[0].Type != "fx_conversion_out" || rows[0].UserID != "u1" || rows[0].Direction != "debit" ||
		rows[0].Currency != "EUR" || rows[0].Amount != "100.00" {
		t.Fatalf("debit row = %+v", rows[0])
	}

	if rows[1].Type != "fx_conversion_in" || rows[1].UserID != "u1" || rows[1].Direction != "credit" ||
		rows[1].Currency != "USD" || rows[1].Amount != "107.46" {
		t.Fatalf("credit row = %+v", rows[1])
	}
}

func TestApplyBankTransferSentAndReturned(t *testing.T) {
	sentPayload, _ := json.Marshal(events.BankTransferSent{
		OrderID: "bto1", UserID: "u1", LedgerTransferID: "ltx1",
		Amount: "75.00", Currency: "USD", CounterpartyIBAN: "DE00OTHER", Reference: "invoice",
	})

	rows, update, err := Apply(events.Envelope{
		EventID: "e6", EventType: events.TypeBankTransferSent, Payload: sentPayload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if update != nil {
		t.Fatal("expected no capture update for sent")
	}

	if len(rows) != 1 || rows[0].Type != "bank_transfer_out" || rows[0].Direction != "debit" ||
		rows[0].Status != "processing" || rows[0].Amount != "75.00" {
		t.Fatalf("sent row = %+v", rows)
	}

	returnedPayload, _ := json.Marshal(events.BankTransferReturned{
		OrderID: "bto1", UserID: "u1", Amount: "75.00", Currency: "USD",
		ReturnTransferID: "ltx2", Reason: "returned",
	})

	_, returnUpdate, err := Apply(events.Envelope{
		EventID: "e7", EventType: events.TypeBankTransferReturned, Payload: returnedPayload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if returnUpdate == nil || returnUpdate.ID != "bto1" || returnUpdate.Status != "returned" ||
		returnUpdate.Amount != "75.00" || returnUpdate.Currency != "USD" {
		t.Fatalf("return update = %+v", returnUpdate)
	}
}

func TestApplyCardAuthLifecycle(t *testing.T) {
	approvedPayload, _ := json.Marshal(events.CardAuthApproved{
		AuthorizationID: "a1", UserID: "u1", Amount: "5.00", Currency: "USD", MerchantName: "Coffee",
	})

	rows, _, err := Apply(events.Envelope{
		EventID: "e2", EventType: events.TypeCardAuthApproved, Payload: approvedPayload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 1 || rows[0].Type != "card_hold" || rows[0].Status != "authorized" {
		t.Fatalf("approved row = %+v", rows)
	}

	capturedPayload, _ := json.Marshal(events.CardAuthCaptured{
		AuthorizationID: "a1", UserID: "u1", Amount: "5.00", Currency: "USD",
	})

	_, update, err := Apply(events.Envelope{
		EventID: "e3", EventType: events.TypeCardAuthCaptured, Payload: capturedPayload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if update == nil || update.Type != "card_purchase" || update.Status != "captured" {
		t.Fatalf("capture update = %+v", update)
	}
}

func TestApplyCardAuthVoided(t *testing.T) {
	payload, _ := json.Marshal(events.CardAuthVoided{
		AuthorizationID: "a1", CardID: "c1", UserID: "u1", Amount: "5.00", Currency: "USD",
	})

	rows, update, err := Apply(events.Envelope{
		EventID: "e5", EventType: events.TypeCardAuthVoided, Payload: payload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if rows != nil {
		t.Fatal("expected no new rows for a void")
	}

	if update == nil || update.Type != "card_hold_released" || update.Status != "voided" {
		t.Fatalf("void update = %+v", update)
	}
}
