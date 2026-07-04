package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/iho/neobank/services/simulators/fx/internal/domain"
	"github.com/iho/neobank/services/simulators/fx/internal/usecase"
)

type Server struct {
	getQuoteUC     *usecase.GetQuoteUseCase
	executeQuoteUC *usecase.ExecuteQuoteUseCase
}

func NewServer(getQuote *usecase.GetQuoteUseCase, executeQuote *usecase.ExecuteQuoteUseCase) *Server {
	return &Server{getQuoteUC: getQuote, executeQuoteUC: executeQuote}
}

func (s *Server) Mount(r chi.Router) {
	r.Get("/health", s.health)
	r.Get("/v1/rates", s.getRate)
	r.Post("/v1/quotes", s.createQuote)
	r.Post("/v1/quotes/{id}/execute", s.executeQuote)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "fx-simulator"})
}

type rateResponse struct {
	AsOf time.Time `json:"as_of"`
	From string    `json:"from_currency"`
	To   string    `json:"to_currency"`
	Rate string    `json:"rate"`
}

func (s *Server) getRate(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from_currency")
	to := r.URL.Query().Get("to_currency")

	mid, err := usecase.MidRate(from, to, time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, rateResponse{
		From: from,
		To:   to,
		Rate: mid.String(),
		AsOf: time.Now().UTC(),
	})
}

type createQuoteRequest struct {
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       string `json:"amount"`
}

type quoteResponse struct {
	ExpiresAt       time.Time  `json:"expires_at"`
	ExecutedAt      *time.Time `json:"executed_at,omitempty"`
	ID              string     `json:"id"`
	FromCurrency    string     `json:"from_currency"`
	ToCurrency      string     `json:"to_currency"`
	Amount          string     `json:"amount"`
	ConvertedAmount string     `json:"converted_amount"`
	Rate            string     `json:"rate"`
	Status          string     `json:"status"`
	SpreadBps       int        `json:"spread_bps"`
}

func (s *Server) createQuote(w http.ResponseWriter, r *http.Request) {
	var req createQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	quote, err := s.getQuoteUC.Execute(r.Context(), usecase.GetQuoteInput{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Amount:       req.Amount,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toQuoteResponse(quote))
}

func (s *Server) executeQuote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	quote, err := s.executeQuoteUC.Execute(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toQuoteResponse(quote))
}

func toQuoteResponse(q domain.Quote) quoteResponse {
	return quoteResponse{
		ID:              q.ID,
		FromCurrency:    q.FromCurrency,
		ToCurrency:      q.ToCurrency,
		Amount:          q.Amount,
		ConvertedAmount: q.ConvertedAmount,
		Rate:            q.Rate,
		SpreadBps:       q.SpreadBps,
		Status:          q.Status,
		ExpiresAt:       q.ExpiresAt,
		ExecutedAt:      q.ExecutedAt,
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
