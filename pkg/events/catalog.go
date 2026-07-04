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

package events

// CatalogVersion identifies the contract document; bump when entries or envelope
// requirements change.
const CatalogVersion = "1.0.0"

// CatalogEntry documents one published domain event for replay and audit.
type CatalogEntry struct {
	EventType     string   `json:"event_type"`
	EventVersion  int      `json:"event_version"`
	AggregateType string   `json:"aggregate_type"`
	Description   string   `json:"description"`
	Topics        []string `json:"topics"`
	PayloadFields []string `json:"payload_fields"`
}

// CatalogDocument is the machine-readable event contract.
type CatalogDocument struct {
	CatalogVersion string         `json:"catalog_version"`
	Envelope       EnvelopeSpec   `json:"envelope"`
	Events         []CatalogEntry `json:"events"`
}

// EnvelopeSpec describes the Kafka/HTTP wrapper persisted in outbox tables.
type EnvelopeSpec struct {
	Required []string `json:"required"`
	Optional []string `json:"optional"`
}

// Catalog returns all registered domain events.
func Catalog() []CatalogEntry {
	return []CatalogEntry{
		{
			EventType:     TypeUserRegistered,
			EventVersion:  1,
			AggregateType: "user",
			Description:   "New user account created",
			Topics:        []string{"user.events"},
			PayloadFields: []string{"user_id", "email"},
		},
		{
			EventType:     TypeKYCApproved,
			EventVersion:  1,
			AggregateType: "user",
			Description:   "KYC case approved and wallet provisioning may proceed",
			Topics:        []string{"user.events"},
			PayloadFields: []string{"user_id", "kyc_case_id"},
		},
		{
			EventType:     TypeKYCRejected,
			EventVersion:  1,
			AggregateType: "user",
			Description:   "KYC case rejected after screening",
			Topics:        []string{"user.events"},
			PayloadFields: []string{"user_id", "kyc_case_id", "rejection_reason"},
		},
		{
			EventType:     TypeWalletProvisioned,
			EventVersion:  1,
			AggregateType: "wallet",
			Description:   "Ledger account linked to a user wallet",
			Topics:        []string{"user.events"},
			PayloadFields: []string{"user_id", "wallet_id", "ledger_account_id", "currency"},
		},
		{
			EventType:     TypeDepositCompleted,
			EventVersion:  1,
			AggregateType: "deposit",
			Description:   "Simulated deposit credited to a user wallet",
			Topics:        []string{"user.events"},
			PayloadFields: []string{"deposit_id", "user_id", "wallet_id", "ledger_transfer_id", "amount", "currency"},
		},
		{
			EventType:     TypeTransferCompleted,
			EventVersion:  1,
			AggregateType: "transfer",
			Description:   "P2P transfer settled on the ledger",
			Topics:        []string{"payment.events"},
			PayloadFields: []string{"transfer_id", "ledger_transfer_id", "sender_user_id", "recipient_user_id", "amount", "currency"},
		},
		{
			EventType:     TypeBankTransferReceived,
			EventVersion:  1,
			AggregateType: "bank_transfer",
			Description:   "Inbound bank transfer (rails simulator or real rail) credited to a user wallet",
			Topics:        []string{"payment.events"},
			PayloadFields: []string{"transfer_id", "user_id", "ledger_transfer_id", "amount", "currency", "sender_name", "reference"},
		},
		{
			EventType:     TypeFXConversionCompleted,
			EventVersion:  1,
			AggregateType: "fx_conversion",
			Description:   "Currency conversion executed via the fx rates simulator",
			Topics:        []string{"payment.events"},
			PayloadFields: []string{"conversion_id", "user_id", "quote_id", "from_currency", "to_currency", "amount", "converted_amount", "rate", "from_ledger_transfer_id", "to_ledger_transfer_id"},
		},
		{
			EventType:     TypeBankTransferSent,
			EventVersion:  1,
			AggregateType: "bank_transfer_order",
			Description:   "Outbound bank transfer debited from a user wallet, pending rail settlement",
			Topics:        []string{"payment.events"},
			PayloadFields: []string{"order_id", "user_id", "ledger_transfer_id", "amount", "currency", "counterparty_iban", "reference"},
		},
		{
			EventType:     TypeBankTransferReturned,
			EventVersion:  1,
			AggregateType: "bank_transfer_order",
			Description:   "Outbound bank transfer bounced or failed after appearing to settle; funds reversed to the wallet",
			Topics:        []string{"payment.events"},
			PayloadFields: []string{"order_id", "user_id", "amount", "currency", "return_transfer_id", "reason"},
		},
		{
			EventType:     TypeCardIssued,
			EventVersion:  1,
			AggregateType: "card",
			Description:   "Virtual card issued to a user",
			Topics:        []string{"card.events"},
			PayloadFields: []string{"card_id", "user_id", "wallet_id", "last_four", "expiry_month", "expiry_year"},
		},
		{
			EventType:     TypeCardFrozen,
			EventVersion:  1,
			AggregateType: "card",
			Description:   "Card frozen by user or operator",
			Topics:        []string{"card.events"},
			PayloadFields: []string{"card_id", "user_id"},
		},
		{
			EventType:     TypeCardUnfrozen,
			EventVersion:  1,
			AggregateType: "card",
			Description:   "Previously frozen card reactivated",
			Topics:        []string{"card.events"},
			PayloadFields: []string{"card_id", "user_id"},
		},
		{
			EventType:     TypeCardAuthApproved,
			EventVersion:  1,
			AggregateType: "authorization",
			Description:   "Card authorization hold placed on the ledger",
			Topics:        []string{"card.events"},
			PayloadFields: []string{"authorization_id", "card_id", "user_id", "amount", "currency", "merchant_name", "merchant_category_code"},
		},
		{
			EventType:     TypeCardAuthCaptured,
			EventVersion:  1,
			AggregateType: "authorization",
			Description:   "Authorization captured and settled to merchant",
			Topics:        []string{"card.events"},
			PayloadFields: []string{"authorization_id", "card_id", "user_id", "amount", "currency", "ledger_transfer_id"},
		},
		{
			EventType:     TypeCardAuthVoided,
			EventVersion:  1,
			AggregateType: "authorization",
			Description:   "Authorization hold released without capture (reversal or expiry)",
			Topics:        []string{"card.events"},
			PayloadFields: []string{"authorization_id", "card_id", "user_id", "amount", "currency"},
		},
		{
			EventType:     TypeCardChargebackOpened,
			EventVersion:  1,
			AggregateType: "dispute",
			Description:   "Captured transaction disputed; cardholder issued a provisional credit",
			Topics:        []string{"card.events"},
			PayloadFields: []string{"dispute_id", "authorization_id", "card_id", "user_id", "amount", "currency", "reason", "provisional_credit_transfer_id"},
		},
		{
			EventType:     TypeCardChargebackResolved,
			EventVersion:  1,
			AggregateType: "dispute",
			Description:   "Dispute resolved won (credit finalized) or lost (credit reversed)",
			Topics:        []string{"card.events"},
			PayloadFields: []string{"dispute_id", "authorization_id", "card_id", "user_id", "amount", "currency", "outcome", "reversal_transfer_id"},
		},
	}
}

// CatalogDocumentJSON returns the full contract for export and compliance tooling.
func CatalogDocumentJSON() CatalogDocument {
	return CatalogDocument{
		CatalogVersion: CatalogVersion,
		Envelope: EnvelopeSpec{
			Required: []string{
				"event_id", "event_type", "event_version", "occurred_at",
				"aggregate_type", "aggregate_id", "payload",
			},
			Optional: []string{"correlation_id", "causation_id"},
		},
		Events: Catalog(),
	}
}

// LookupCatalog finds metadata for an event type.
func LookupCatalog(eventType string) (CatalogEntry, bool) {
	for _, entry := range Catalog() {
		if entry.EventType == eventType {
			return entry, true
		}
	}
	return CatalogEntry{}, false
}

// RegisteredEvents returns sample Event values used to validate the catalog.
func RegisteredEvents() []Event {
	return []Event{
		UserRegistered{UserID: "user-1", Email: "a@example.com"},
		KYCApproved{UserID: "user-1", KYCCaseID: "kyc-1"},
		KYCRejected{UserID: "user-1", KYCCaseID: "kyc-1", RejectionReason: "stub_sanctions_match"},
		WalletProvisioned{UserID: "user-1", WalletID: "wallet-1", LedgerAccountID: "acct-1", Currency: "USD"},
		DepositCompleted{
			DepositID: "dep-1", UserID: "user-1", WalletID: "wallet-1",
			LedgerTransferID: "ltx-dep-1", Amount: "100.00", Currency: "USD",
		},
		TransferCompleted{
			TransferID: "tx-1", LedgerTransferID: "ltx-1",
			SenderUserID: "user-1", RecipientUserID: "user-2",
			Amount: "10.00", Currency: "USD",
		},
		BankTransferReceived{
			TransferID: "bank-tx-1", UserID: "user-1", LedgerTransferID: "ltx-bank-1",
			Amount: "250.00", Currency: "USD", SenderName: "Jane Doe", Reference: "rent",
		},
		FXConversionCompleted{
			ConversionID: "fxc-1", UserID: "user-1", QuoteID: "quote-1",
			FromCurrency: "EUR", ToCurrency: "USD", Amount: "100.00", ConvertedAmount: "107.46",
			Rate: "1.0746", FromLedgerTransferID: "ltx-fx-1", ToLedgerTransferID: "ltx-fx-2",
		},
		BankTransferSent{
			OrderID: "bto-1", UserID: "user-1", LedgerTransferID: "ltx-out-1",
			Amount: "75.00", Currency: "USD", CounterpartyIBAN: "DE00OTHER", Reference: "invoice",
		},
		BankTransferReturned{
			OrderID: "bto-1", UserID: "user-1", Amount: "75.00", Currency: "USD",
			ReturnTransferID: "ltx-out-2", Reason: "returned",
		},
		CardIssued{
			CardID: "card-1", UserID: "user-1", WalletID: "wallet-1",
			LastFour: "4242", ExpiryMonth: 12, ExpiryYear: 2028,
		},
		CardFrozen{CardID: "card-1", UserID: "user-1"},
		CardUnfrozen{CardID: "card-1", UserID: "user-1"},
		CardAuthApproved{
			AuthorizationID: "auth-1", CardID: "card-1", UserID: "user-1",
			Amount: "25.00", Currency: "USD", MerchantName: "Coffee Shop",
			MerchantCategoryCode: "5812",
		},
		CardAuthCaptured{
			AuthorizationID: "auth-1", CardID: "card-1", UserID: "user-1",
			Amount: "25.00", Currency: "USD", LedgerTransferID: "ltx-2",
		},
		CardAuthVoided{
			AuthorizationID: "auth-2", CardID: "card-1", UserID: "user-1",
			Amount: "15.00", Currency: "USD",
		},
		CardChargebackOpened{
			DisputeID: "dispute-1", AuthorizationID: "auth-1", CardID: "card-1", UserID: "user-1",
			Amount: "25.00", Currency: "USD", Reason: "fraud", ProvisionalCreditTransferID: "ltx-cb-1",
		},
		CardChargebackResolved{
			DisputeID: "dispute-1", AuthorizationID: "auth-1", CardID: "card-1", UserID: "user-1",
			Amount: "25.00", Currency: "USD", Outcome: "lost", ReversalTransferID: "ltx-cb-2",
		},
	}
}
