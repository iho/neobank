package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/iho/neobank/pkg/vendorsim"
	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/port"
	"github.com/iho/neobank/services/simulators/cardproc/internal/usecase"
)

type Server struct {
	issueCard         *usecase.IssueCardUseCase
	simulateTx        *usecase.SimulateTransactionUseCase
	captureTx         *usecase.CaptureTransactionUseCase
	reverseTx         *usecase.ReverseTransactionUseCase
	openChargeback    *usecase.OpenChargebackUseCase
	resolveChargeback *usecase.ResolveChargebackUseCase
	cards             port.CardRepository
	chargebacks       port.ChargebackRepository
	deliveries        vendorsim.DeliveryStore
}

func NewServer(
	issueCard *usecase.IssueCardUseCase,
	simulateTx *usecase.SimulateTransactionUseCase,
	captureTx *usecase.CaptureTransactionUseCase,
	reverseTx *usecase.ReverseTransactionUseCase,
	openChargeback *usecase.OpenChargebackUseCase,
	resolveChargeback *usecase.ResolveChargebackUseCase,
	cards port.CardRepository,
	chargebacks port.ChargebackRepository,
	deliveries vendorsim.DeliveryStore,
) *Server {
	return &Server{
		issueCard:         issueCard,
		simulateTx:        simulateTx,
		captureTx:         captureTx,
		reverseTx:         reverseTx,
		openChargeback:    openChargeback,
		resolveChargeback: resolveChargeback,
		cards:             cards,
		chargebacks:       chargebacks,
		deliveries:        deliveries,
	}
}

func (s *Server) Mount(r chi.Router) {
	r.Get("/health", s.health)
	r.Post("/v1/cards", s.createCard)
	r.Post("/v1/cards/{ref}/cancel", s.cancelCard)
	r.Post("/sim/transactions", s.simulateTransaction)
	r.Post("/sim/transactions/{id}/capture", s.captureTransaction)
	r.Post("/sim/transactions/{id}/reverse", s.reverseTransaction)
	r.Post("/sim/transactions/{id}/chargeback", s.openChargebackHandler)
	r.Post("/sim/chargebacks/{id}/resolve", s.resolveChargebackHandler)
	r.Get("/sim/chargebacks/{id}", s.getChargeback)
	r.Get("/sim/deliveries", s.listDeliveries)
	r.Get("/sim/deliveries/{id}", s.getDelivery)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "cardproc-simulator"})
}

type createCardRequest struct {
	ExternalRef    string `json:"external_ref"`
	CardholderName string `json:"cardholder_name"`
}

type cardResponse struct {
	Ref         string `json:"ref"`
	PANToken    string `json:"pan_token"`
	LastFour    string `json:"last_four"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

func (s *Server) createCard(w http.ResponseWriter, r *http.Request) {
	var req createCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	card, err := s.issueCard.Execute(r.Context(), usecase.IssueCardInput{
		ExternalRef:    req.ExternalRef,
		CardholderName: req.CardholderName,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toCardResponse(card))
}

func (s *Server) cancelCard(w http.ResponseWriter, r *http.Request) {
	ref := chi.URLParam(r, "ref")

	if err := s.cards.Cancel(r.Context(), ref); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

type simulateTransactionRequest struct {
	CardRef      string `json:"card_ref"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	MerchantName string `json:"merchant_name"`
	MCC          string `json:"mcc"`
	Capture      bool   `json:"capture"`
}

type transactionResponse struct {
	ID              string `json:"id"`
	CardID          string `json:"card_id"`
	AuthorizationID string `json:"authorization_id,omitempty"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	MerchantName    string `json:"merchant_name"`
	MCC             string `json:"mcc"`
	Status          string `json:"status"`
	ReasonCode      string `json:"reason_code,omitempty"`
}

func (s *Server) simulateTransaction(w http.ResponseWriter, r *http.Request) {
	var req simulateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := s.simulateTx.Execute(r.Context(), usecase.SimulateTransactionInput{
		CardRef:      req.CardRef,
		Amount:       req.Amount,
		Currency:     req.Currency,
		MerchantName: req.MerchantName,
		MCC:          req.MCC,
		Capture:      req.Capture,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toTransactionResponse(tx))
}

func (s *Server) captureTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	tx, err := s.captureTx.Execute(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toTransactionResponse(tx))
}

type reverseTransactionRequest struct {
	Reason string `json:"reason"`
}

func (s *Server) reverseTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req reverseTransactionRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	tx, err := s.reverseTx.Execute(r.Context(), id, req.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toTransactionResponse(tx))
}

type openChargebackRequest struct {
	Reason string `json:"reason"`
}

type resolveChargebackRequest struct {
	Outcome string `json:"outcome"`
}

type chargebackResponse struct {
	ID              string `json:"id"`
	TransactionID   string `json:"transaction_id"`
	AuthorizationID string `json:"authorization_id"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	Reason          string `json:"reason"`
	Status          string `json:"status"`
}

func (s *Server) openChargebackHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req openChargebackRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	cb, err := s.openChargeback.Execute(r.Context(), id, req.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toChargebackResponse(cb))
}

func (s *Server) resolveChargebackHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req resolveChargebackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cb, err := s.resolveChargeback.Execute(r.Context(), id, req.Outcome)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toChargebackResponse(cb))
}

func (s *Server) getChargeback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cb, err := s.chargebacks.GetByID(r.Context(), id)
	if err != nil || cb == nil {
		writeError(w, http.StatusNotFound, "chargeback not found")
		return
	}

	writeJSON(w, http.StatusOK, toChargebackResponse(*cb))
}

func toChargebackResponse(cb domain.Chargeback) chargebackResponse {
	return chargebackResponse{
		ID:              cb.ID,
		TransactionID:   cb.TransactionID,
		AuthorizationID: cb.AuthorizationID,
		Amount:          cb.Amount,
		Currency:        cb.Currency,
		Reason:          cb.Reason,
		Status:          cb.Status,
	}
}

func (s *Server) listDeliveries(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	deliveries, err := s.deliveries.List(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"deliveries": deliveries})
}

func (s *Server) getDelivery(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	delivery, err := s.deliveries.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "delivery not found")
		return
	}

	writeJSON(w, http.StatusOK, delivery)
}

func toCardResponse(c domain.Card) cardResponse {
	return cardResponse{
		Ref:         c.ID,
		PANToken:    c.PANToken,
		LastFour:    c.LastFour,
		ExpiryMonth: c.ExpiryMonth,
		ExpiryYear:  c.ExpiryYear,
	}
}

func toTransactionResponse(t domain.Transaction) transactionResponse {
	return transactionResponse{
		ID:              t.ID,
		CardID:          t.CardID,
		AuthorizationID: t.AuthorizationID,
		Amount:          t.Amount,
		Currency:        t.Currency,
		MerchantName:    t.MerchantName,
		MCC:             t.MCC,
		Status:          t.Status,
		ReasonCode:      t.ReasonCode,
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
