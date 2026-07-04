package api

import (
	"encoding/json"
	"net/http"

	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/usecase"
)

// RailsHandlers are the bank-transfer endpoints layered on top of the
// oapi-codegen strict Server: a user-facing virtual-IBAN endpoint and the
// webhook consumer for the rails simulator. They are plain net/http (not
// generated from the OpenAPI spec) because the webhook route in particular
// must sit outside the global Idempotency-Key middleware — see
// cmd/server/main.go.
type RailsHandlers struct {
	getOrCreateBankAccount *usecase.GetOrCreateBankAccountUseCase
	processInboundTransfer *usecase.ProcessInboundTransferUseCase
}

func NewRailsHandlers(
	getOrCreateBankAccount *usecase.GetOrCreateBankAccountUseCase,
	processInboundTransfer *usecase.ProcessInboundTransferUseCase,
) *RailsHandlers {
	return &RailsHandlers{
		getOrCreateBankAccount: getOrCreateBankAccount,
		processInboundTransfer: processInboundTransfer,
	}
}

type bankAccountResponse struct {
	UserID   string `json:"user_id"`
	Currency string `json:"currency"`
	IBAN     string `json:"iban"`
}

// GetBankAccount handles GET /api/v1/payments/bank-accounts?currency=USD,
// returning the caller's virtual IBAN for topping up by bank transfer.
func (h *RailsHandlers) GetBankAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeAPIError(w, http.StatusUnauthorized, "missing X-User-Id")
		return
	}

	currency := r.URL.Query().Get("currency")
	if currency == "" {
		currency = "USD"
	}

	account, err := h.getOrCreateBankAccount.Execute(r.Context(), usecase.GetOrCreateBankAccountInput{
		UserID:   userID,
		Currency: currency,
	})
	if err != nil {
		writeAPIError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeAPIJSON(w, http.StatusOK, bankAccountResponse{
		UserID:   account.UserID,
		Currency: account.Currency,
		IBAN:     account.IBAN,
	})
}

// HandleRailsWebhook handles POST /webhooks/rails. It is mounted behind
// vendorsim.VerifyWebhook, so the request is already authenticated and
// de-duplicated by delivery ID by the time it reaches here; the use case
// applies its own idempotency on the rail's transfer ID as a second,
// durable layer (the webhook nonce store is in-memory only).
func (h *RailsHandlers) HandleRailsWebhook(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-Vendorsim-Event")

	switch eventType {
	case "rails.transfer.received":
		h.handleTransferReceived(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "unknown event type: "+eventType)
	}
}

func (h *RailsHandlers) handleTransferReceived(w http.ResponseWriter, r *http.Request) {
	var payload usecase.InboundTransferWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	transfer, err := h.processInboundTransfer.Execute(r.Context(), payload)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeAPIJSON(w, http.StatusOK, toBankTransferResponse(transfer))
}

type bankTransferResponse struct {
	ID               string `json:"id"`
	RailsTransferID  string `json:"rails_transfer_id"`
	UserID           string `json:"user_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	LedgerTransferID string `json:"ledger_transfer_id"`
	Status           string `json:"status"`
}

func toBankTransferResponse(t domain.BankTransfer) bankTransferResponse {
	return bankTransferResponse{
		ID:               t.ID,
		RailsTransferID:  t.RailsTransferID,
		UserID:           t.UserID,
		Amount:           t.Amount,
		Currency:         t.Currency,
		LedgerTransferID: t.LedgerTransferID,
		Status:           t.Status,
	}
}

func writeAPIJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeAPIError(w http.ResponseWriter, status int, msg string) {
	writeAPIJSON(w, status, map[string]string{"error": msg})
}
