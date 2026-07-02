package events

const TypeTransferCompleted = "payment.transfer.completed"

type TransferCompleted struct {
	TransferID       string `json:"transfer_id"`
	LedgerTransferID string `json:"ledger_transfer_id"`
	SenderUserID     string `json:"sender_user_id"`
	RecipientUserID  string `json:"recipient_user_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
}

func (e TransferCompleted) EventType() string     { return TypeTransferCompleted }
func (e TransferCompleted) AggregateType() string { return "transfer" }
func (e TransferCompleted) AggregateID() string   { return e.TransferID }
func (e TransferCompleted) Version() int          { return 1 }