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
	TypeCardIssued       = "card.issued"
	TypeCardFrozen       = "card.frozen"
	TypeCardUnfrozen     = "card.unfrozen"
	TypeCardAuthApproved = "card.auth.approved"
	TypeCardAuthCaptured = "card.auth.captured"
	TypeCardAuthVoided   = "card.auth.voided"
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
	AuthorizationID       string `json:"authorization_id"`
	CardID                string `json:"card_id"`
	UserID                string `json:"user_id"`
	Amount                string `json:"amount"`
	Currency              string `json:"currency"`
	MerchantName          string `json:"merchant_name"`
	MerchantCategoryCode  string `json:"merchant_category_code,omitempty"`
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
