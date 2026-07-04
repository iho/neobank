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

const TypeTransferCompleted = "payment.transfer.completed"

type TransferCompleted struct {
	TransferID           string `json:"transfer_id"`
	LedgerTransferID     string `json:"ledger_transfer_id"`
	SenderUserID         string `json:"sender_user_id"`
	RecipientUserID      string `json:"recipient_user_id"`
	Amount               string `json:"amount"`
	Currency             string `json:"currency"`
	Memo                 string `json:"memo,omitempty"`
	SenderDisplayName    string `json:"sender_display_name,omitempty"`
	RecipientDisplayName string `json:"recipient_display_name,omitempty"`
}

func (e TransferCompleted) EventType() string     { return TypeTransferCompleted }
func (e TransferCompleted) AggregateType() string { return "transfer" }
func (e TransferCompleted) AggregateID() string   { return e.TransferID }
func (e TransferCompleted) Version() int          { return 1 }

const TypeBankTransferReceived = "payment.bank_transfer.received"

// BankTransferReceived is published when the rails simulator (or, later, a
// real payment rail) delivers an inbound transfer that payment credited to
// a user's wallet.
type BankTransferReceived struct {
	TransferID       string `json:"transfer_id"`
	UserID           string `json:"user_id"`
	LedgerTransferID string `json:"ledger_transfer_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	SenderName       string `json:"sender_name,omitempty"`
	Reference        string `json:"reference,omitempty"`
}

func (e BankTransferReceived) EventType() string     { return TypeBankTransferReceived }
func (e BankTransferReceived) AggregateType() string { return "bank_transfer" }
func (e BankTransferReceived) AggregateID() string   { return e.TransferID }
func (e BankTransferReceived) Version() int          { return 1 }
