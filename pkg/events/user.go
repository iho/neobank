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

import "encoding/json"

const (
	TypeUserRegistered    = "user.registered"
	TypeKYCApproved       = "user.kyc.approved"
	TypeWalletProvisioned = "user.wallet.provisioned"
	TypeDepositCompleted  = "user.deposit.completed"
)

type UserRegistered struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (e UserRegistered) EventType() string     { return TypeUserRegistered }
func (e UserRegistered) AggregateType() string { return "user" }
func (e UserRegistered) AggregateID() string   { return e.UserID }
func (e UserRegistered) Version() int          { return 1 }

type KYCApproved struct {
	UserID    string `json:"user_id"`
	KYCCaseID string `json:"kyc_case_id"`
}

func (e KYCApproved) EventType() string     { return TypeKYCApproved }
func (e KYCApproved) AggregateType() string { return "user" }
func (e KYCApproved) AggregateID() string   { return e.UserID }
func (e KYCApproved) Version() int          { return 1 }

type WalletProvisioned struct {
	UserID          string `json:"user_id"`
	WalletID        string `json:"wallet_id"`
	LedgerAccountID string `json:"ledger_account_id"`
	Currency        string `json:"currency"`
}

func (e WalletProvisioned) EventType() string     { return TypeWalletProvisioned }
func (e WalletProvisioned) AggregateType() string { return "wallet" }
func (e WalletProvisioned) AggregateID() string   { return e.WalletID }
func (e WalletProvisioned) Version() int          { return 1 }

type DepositCompleted struct {
	DepositID        string `json:"deposit_id"`
	UserID           string `json:"user_id"`
	WalletID         string `json:"wallet_id"`
	LedgerTransferID string `json:"ledger_transfer_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
}

func (e DepositCompleted) EventType() string     { return TypeDepositCompleted }
func (e DepositCompleted) AggregateType() string { return "deposit" }
func (e DepositCompleted) AggregateID() string   { return e.DepositID }
func (e DepositCompleted) Version() int          { return 1 }

func MarshalPayload(evt Event) (json.RawMessage, error) {
	return json.Marshal(evt)
}
