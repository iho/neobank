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

const TypeFXConversionCompleted = "payment.fx_conversion.completed"

// FXConversionCompleted is published when a user converts between two
// currency wallets via the fx rates simulator.
type FXConversionCompleted struct {
	ConversionID         string `json:"conversion_id"`
	UserID               string `json:"user_id"`
	QuoteID              string `json:"quote_id"`
	FromCurrency         string `json:"from_currency"`
	ToCurrency           string `json:"to_currency"`
	Amount               string `json:"amount"`
	ConvertedAmount      string `json:"converted_amount"`
	Rate                 string `json:"rate"`
	FromLedgerTransferID string `json:"from_ledger_transfer_id"`
	ToLedgerTransferID   string `json:"to_ledger_transfer_id"`
}

func (e FXConversionCompleted) EventType() string     { return TypeFXConversionCompleted }
func (e FXConversionCompleted) AggregateType() string { return "fx_conversion" }
func (e FXConversionCompleted) AggregateID() string   { return e.ConversionID }
func (e FXConversionCompleted) Version() int          { return 1 }

const TypeBankTransferSent = "payment.bank_transfer.sent"

// BankTransferSent is published when a user sends money out over the rails
// simulator: the wallet is debited immediately, before the rail confirms
// settlement.
type BankTransferSent struct {
	OrderID          string `json:"order_id"`
	UserID           string `json:"user_id"`
	LedgerTransferID string `json:"ledger_transfer_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	CounterpartyIBAN string `json:"counterparty_iban"`
	Reference        string `json:"reference,omitempty"`
}

func (e BankTransferSent) EventType() string     { return TypeBankTransferSent }
func (e BankTransferSent) AggregateType() string { return "bank_transfer_order" }
func (e BankTransferSent) AggregateID() string   { return e.OrderID }
func (e BankTransferSent) Version() int          { return 1 }

const TypeBankTransferReturned = "payment.bank_transfer.returned"

// BankTransferReturned is published when an outbound payment that already
// looked settled bounces back (or fails validation at the rail); the
// ledger transfer that originally debited the wallet has been reversed.
type BankTransferReturned struct {
	OrderID          string `json:"order_id"`
	UserID           string `json:"user_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	ReturnTransferID string `json:"return_transfer_id"`
	Reason           string `json:"reason,omitempty"`
}

func (e BankTransferReturned) EventType() string     { return TypeBankTransferReturned }
func (e BankTransferReturned) AggregateType() string { return "bank_transfer_order" }
func (e BankTransferReturned) AggregateID() string   { return e.OrderID }
func (e BankTransferReturned) Version() int          { return 1 }
