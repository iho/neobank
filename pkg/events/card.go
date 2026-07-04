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

const (
	TypeCardIssued             = "card.issued"
	TypeCardFrozen             = "card.frozen"
	TypeCardUnfrozen           = "card.unfrozen"
	TypeCardAuthApproved       = "card.auth.approved"
	TypeCardAuthCaptured       = "card.auth.captured"
	TypeCardAuthVoided         = "card.auth.voided"
	TypeCardChargebackOpened   = "card.chargeback.opened"
	TypeCardChargebackResolved = "card.chargeback.resolved"
)

type CardIssued struct {
	CardID      string `json:"card_id"`
	UserID      string `json:"user_id"`
	WalletID    string `json:"wallet_id"`
	LastFour    string `json:"last_four"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

func (e CardIssued) EventType() string     { return TypeCardIssued }
func (e CardIssued) AggregateType() string { return "card" }
func (e CardIssued) AggregateID() string   { return e.CardID }
func (e CardIssued) Version() int          { return 1 }

type CardFrozen struct {
	CardID string `json:"card_id"`
	UserID string `json:"user_id"`
}

func (e CardFrozen) EventType() string     { return TypeCardFrozen }
func (e CardFrozen) AggregateType() string { return "card" }
func (e CardFrozen) AggregateID() string   { return e.CardID }
func (e CardFrozen) Version() int          { return 1 }

type CardUnfrozen struct {
	CardID string `json:"card_id"`
	UserID string `json:"user_id"`
}

func (e CardUnfrozen) EventType() string     { return TypeCardUnfrozen }
func (e CardUnfrozen) AggregateType() string { return "card" }
func (e CardUnfrozen) AggregateID() string   { return e.CardID }
func (e CardUnfrozen) Version() int          { return 1 }

type CardAuthApproved struct {
	AuthorizationID      string `json:"authorization_id"`
	CardID               string `json:"card_id"`
	UserID               string `json:"user_id"`
	Amount               string `json:"amount"`
	Currency             string `json:"currency"`
	MerchantName         string `json:"merchant_name"`
	MerchantCategoryCode string `json:"merchant_category_code,omitempty"`
}

func (e CardAuthApproved) EventType() string     { return TypeCardAuthApproved }
func (e CardAuthApproved) AggregateType() string { return "authorization" }
func (e CardAuthApproved) AggregateID() string   { return e.AuthorizationID }
func (e CardAuthApproved) Version() int          { return 1 }

type CardAuthCaptured struct {
	AuthorizationID  string `json:"authorization_id"`
	CardID           string `json:"card_id"`
	UserID           string `json:"user_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	LedgerTransferID string `json:"ledger_transfer_id"`
}

func (e CardAuthCaptured) EventType() string     { return TypeCardAuthCaptured }
func (e CardAuthCaptured) AggregateType() string { return "authorization" }
func (e CardAuthCaptured) AggregateID() string   { return e.AuthorizationID }
func (e CardAuthCaptured) Version() int          { return 1 }

// CardAuthVoided is published when a hold is released without ever being
// captured (the cardproc simulator's "reverse" flow, or an expired hold).
type CardAuthVoided struct {
	AuthorizationID string `json:"authorization_id"`
	CardID          string `json:"card_id"`
	UserID          string `json:"user_id"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
}

func (e CardAuthVoided) EventType() string     { return TypeCardAuthVoided }
func (e CardAuthVoided) AggregateType() string { return "authorization" }
func (e CardAuthVoided) AggregateID() string   { return e.AuthorizationID }
func (e CardAuthVoided) Version() int          { return 1 }

// CardChargebackOpened is published when a captured transaction is disputed:
// the cardholder gets a provisional credit immediately, before the dispute
// is resolved either way (see CardChargebackResolved).
type CardChargebackOpened struct {
	DisputeID                   string `json:"dispute_id"`
	AuthorizationID             string `json:"authorization_id"`
	CardID                      string `json:"card_id"`
	UserID                      string `json:"user_id"`
	Amount                      string `json:"amount"`
	Currency                    string `json:"currency"`
	Reason                      string `json:"reason"`
	ProvisionalCreditTransferID string `json:"provisional_credit_transfer_id"`
}

func (e CardChargebackOpened) EventType() string     { return TypeCardChargebackOpened }
func (e CardChargebackOpened) AggregateType() string { return "dispute" }
func (e CardChargebackOpened) AggregateID() string   { return e.DisputeID }
func (e CardChargebackOpened) Version() int          { return 1 }

// CardChargebackResolved closes out a dispute: "won" (cardholder keeps the
// provisional credit, it's now final) or "lost" (the credit is reversed
// back out of the wallet).
type CardChargebackResolved struct {
	DisputeID          string `json:"dispute_id"`
	AuthorizationID    string `json:"authorization_id"`
	CardID             string `json:"card_id"`
	UserID             string `json:"user_id"`
	Amount             string `json:"amount"`
	Currency           string `json:"currency"`
	Outcome            string `json:"outcome"`
	ReversalTransferID string `json:"reversal_transfer_id,omitempty"`
}

func (e CardChargebackResolved) EventType() string     { return TypeCardChargebackResolved }
func (e CardChargebackResolved) AggregateType() string { return "dispute" }
func (e CardChargebackResolved) AggregateID() string   { return e.DisputeID }
func (e CardChargebackResolved) Version() int          { return 1 }
