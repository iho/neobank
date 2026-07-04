package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/iho/neobank/pkg/vendorsim"
	"github.com/iho/neobank/services/simulators/rails/internal/domain"
	"github.com/iho/neobank/services/simulators/rails/internal/port"
	"github.com/iho/neobank/services/simulators/rails/internal/usecase"
)

type Server struct {
	issueAccount   *usecase.IssueAccountUseCase
	injectTransfer *usecase.InjectInboundTransferUseCase
	statements     *usecase.GetStatementUseCase
	accounts       port.AccountRepository
	deliveries     vendorsim.DeliveryStore
}

func NewServer(
	issueAccount *usecase.IssueAccountUseCase,
	injectTransfer *usecase.InjectInboundTransferUseCase,
	statements *usecase.GetStatementUseCase,
	accounts port.AccountRepository,
	deliveries vendorsim.DeliveryStore,
) *Server {
	return &Server{
		issueAccount:   issueAccount,
		injectTransfer: injectTransfer,
		statements:     statements,
		accounts:       accounts,
		deliveries:     deliveries,
	}
}

func (s *Server) Mount(r chi.Router) {
	r.Get("/health", s.health)
	r.Post("/v1/accounts", s.createAccount)
	r.Get("/v1/accounts/{id}", s.getAccount)
	r.Get("/v1/statements/{date}", s.getStatement)
	r.Post("/sim/inbound-transfers", s.injectInboundTransfer)
	r.Get("/sim/deliveries", s.listDeliveries)
	r.Get("/sim/deliveries/{id}", s.getDelivery)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "rails-simulator"})
}

type createAccountRequest struct {
	ExternalRef string `json:"external_ref"`
	Currency    string `json:"currency"`
}

type accountResponse struct {
	ID          string `json:"id"`
	ExternalRef string `json:"external_ref"`
	Currency    string `json:"currency"`
	IBAN        string `json:"iban"`
}

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	account, err := s.issueAccount.Execute(r.Context(), usecase.IssueAccountInput{
		ExternalRef: req.ExternalRef,
		Currency:    req.Currency,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toAccountResponse(account))
}

func (s *Server) getAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	account, err := s.accounts.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if account == nil {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}

	writeJSON(w, http.StatusOK, toAccountResponse(*account))
}

type injectInboundTransferRequest struct {
	AccountID  string `json:"account_id"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	SenderName string `json:"sender_name"`
	Reference  string `json:"reference"`
}

type inboundTransferResponse struct {
	ID         string `json:"id"`
	AccountID  string `json:"account_id"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	SenderName string `json:"sender_name"`
	Reference  string `json:"reference"`
	Status     string `json:"status"`
}

func (s *Server) injectInboundTransfer(w http.ResponseWriter, r *http.Request) {
	var req injectInboundTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	transfer, err := s.injectTransfer.Execute(r.Context(), usecase.InjectInboundTransferInput{
		AccountID:  req.AccountID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		SenderName: req.SenderName,
		Reference:  req.Reference,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toInboundTransferResponse(transfer))
}

func (s *Server) getStatement(w http.ResponseWriter, r *http.Request) {
	date := chi.URLParam(r, "date")

	transfers, err := s.statements.Execute(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	entries := make([]inboundTransferResponse, 0, len(transfers))
	for _, t := range transfers {
		entries = append(entries, toInboundTransferResponse(t))
	}

	writeJSON(w, http.StatusOK, map[string]any{"date": date, "inbound_transfers": entries})
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

func toAccountResponse(a domain.Account) accountResponse {
	return accountResponse{ID: a.ID, ExternalRef: a.ExternalRef, Currency: a.Currency, IBAN: a.IBAN}
}

func toInboundTransferResponse(t domain.InboundTransfer) inboundTransferResponse {
	return inboundTransferResponse{
		ID:         t.ID,
		AccountID:  t.AccountID,
		Amount:     t.Amount,
		Currency:   t.Currency,
		SenderName: t.SenderName,
		Reference:  t.Reference,
		Status:     t.Status,
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
