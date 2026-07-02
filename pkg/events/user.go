package events

import "encoding/json"

const (
	TypeUserRegistered    = "user.registered"
	TypeKYCApproved       = "user.kyc.approved"
	TypeWalletProvisioned = "user.wallet.provisioned"
)

type UserRegistered struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (e UserRegistered) EventType() string     { return TypeUserRegistered }
func (e UserRegistered) AggregateType() string { return "user" }
func (e UserRegistered) AggregateID() string { return e.UserID }
func (e UserRegistered) Version() int          { return 1 }

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

func MarshalPayload(evt Event) (json.RawMessage, error) {
	return json.Marshal(evt)
}