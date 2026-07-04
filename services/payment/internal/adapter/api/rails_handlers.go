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
	sendBankTransfer       *usecase.SendBankTransferUseCase
	processOutboundWebhook *usecase.ProcessOutboundPaymentWebhookUseCase
}

func NewRailsHandlers(
	getOrCreateBankAccount *usecase.GetOrCreateBankAccountUseCase,
	processInboundTransfer *usecase.ProcessInboundTransferUseCase,
	sendBankTransfer *usecase.SendBankTransferUseCase,
	processOutboundWebhook *usecase.ProcessOutboundPaymentWebhookUseCase,
) *RailsHandlers {
	return &RailsHandlers{
		getOrCreateBankAccount: getOrCreateBankAccount,
		processInboundTransfer: processInboundTransfer,
		sendBankTransfer:       sendBankTransfer,
		processOutboundWebhook: processOutboundWebhook,
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

type sendBankTransferRequest struct {
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	CounterpartyIBAN string `json:"counterparty_iban"`
	Reference        string `json:"reference"`
}

type bankTransferOrderResponse struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	CounterpartyIBAN string `json:"counterparty_iban"`
	Reference        string `json:"reference"`
	LedgerTransferID string `json:"ledger_transfer_id"`
	Status           string `json:"status"`
}

// SendBankTransfer handles POST /api/v1/payments/bank-transfers, sending
// money out over the rails simulator.
func (h *RailsHandlers) SendBankTransfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeAPIError(w, http.StatusUnauthorized, "missing X-User-Id")
		return
	}

	var req sendBankTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.sendBankTransfer.Execute(r.Context(), usecase.SendBankTransferInput{
		UserID:           userID,
		Amount:           req.Amount,
		Currency:         req.Currency,
		CounterpartyIBAN: req.CounterpartyIBAN,
		Reference:        req.Reference,
	})
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeAPIJSON(w, http.StatusCreated, toBankTransferOrderResponse(order))
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
	case "rails.payment.settled":
		h.handlePaymentSettled(w, r)
	case "rails.payment.returned":
		h.handlePaymentReturned(w, r)
	case "rails.payment.failed":
		h.handlePaymentFailed(w, r)
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

func (h *RailsHandlers) handlePaymentSettled(w http.ResponseWriter, r *http.Request) {
	var payload usecase.OutboundPaymentWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.processOutboundWebhook.Settled(r.Context(), payload)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeAPIJSON(w, http.StatusOK, toBankTransferOrderResponse(order))
}

func (h *RailsHandlers) handlePaymentReturned(w http.ResponseWriter, r *http.Request) {
	h.handlePaymentReversal(w, r, domain.BankTransferOrderStatusReturned, "returned")
}

func (h *RailsHandlers) handlePaymentFailed(w http.ResponseWriter, r *http.Request) {
	h.handlePaymentReversal(w, r, domain.BankTransferOrderStatusFailed, "failed")
}

func (h *RailsHandlers) handlePaymentReversal(w http.ResponseWriter, r *http.Request, status, reason string) {
	var payload usecase.OutboundPaymentWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.processOutboundWebhook.ReturnedOrFailed(r.Context(), payload, status, reason)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeAPIJSON(w, http.StatusOK, toBankTransferOrderResponse(order))
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

func toBankTransferOrderResponse(o domain.BankTransferOrder) bankTransferOrderResponse {
	return bankTransferOrderResponse{
		ID:               o.ID,
		UserID:           o.UserID,
		Amount:           o.Amount,
		Currency:         o.Currency,
		CounterpartyIBAN: o.CounterpartyIBAN,
		Reference:        o.Reference,
		LedgerTransferID: o.LedgerTransferID,
		Status:           o.Status,
	}
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
