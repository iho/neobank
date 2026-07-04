package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iho/neobank/services/payment/internal/adapter/fxclient"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/usecase"
)

// FXHandlers are the currency-conversion endpoints layered on top of the
// oapi-codegen strict Server, plain net/http like RailsHandlers since there
// is no OpenAPI spec for them yet (see docs/vendor-simulators-plan.md Phase 4).
type FXHandlers struct {
	getQuote     *usecase.GetFXQuoteUseCase
	executeQuote *usecase.ExecuteFXConversionUseCase
}

func NewFXHandlers(getQuote *usecase.GetFXQuoteUseCase, executeQuote *usecase.ExecuteFXConversionUseCase) *FXHandlers {
	return &FXHandlers{getQuote: getQuote, executeQuote: executeQuote}
}

type createFXQuoteRequest struct {
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       string `json:"amount"`
}

// CreateQuote handles POST /api/v1/payments/fx/quotes, pricing a conversion
// for the caller to review before executing it within the quote's TTL.
func (h *FXHandlers) CreateQuote(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeAPIError(w, http.StatusUnauthorized, "missing X-User-Id")
		return
	}

	var req createFXQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	quote, err := h.getQuote.Execute(r.Context(), usecase.GetFXQuoteInput{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Amount:       req.Amount,
	})
	if err != nil {
		writeAPIError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeAPIJSON(w, http.StatusCreated, toFXQuoteResponse(quote))
}

// ExecuteQuote handles POST /api/v1/payments/fx/quotes/{id}/execute,
// locking in the quote and moving funds between the caller's wallets.
func (h *FXHandlers) ExecuteQuote(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeAPIError(w, http.StatusUnauthorized, "missing X-User-Id")
		return
	}

	quoteID := chi.URLParam(r, "id")

	conversion, err := h.executeQuote.Execute(r.Context(), usecase.ExecuteFXConversionInput{
		UserID:  userID,
		QuoteID: quoteID,
	})
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeAPIJSON(w, http.StatusOK, toFXConversionResponse(conversion))
}

type fxQuoteResponse struct {
	ID              string `json:"id"`
	FromCurrency    string `json:"from_currency"`
	ToCurrency      string `json:"to_currency"`
	Amount          string `json:"amount"`
	ConvertedAmount string `json:"converted_amount"`
	Rate            string `json:"rate"`
	Status          string `json:"status"`
}

func toFXQuoteResponse(q fxclient.Quote) fxQuoteResponse {
	return fxQuoteResponse{
		ID:              q.ID,
		FromCurrency:    q.FromCurrency,
		ToCurrency:      q.ToCurrency,
		Amount:          q.Amount,
		ConvertedAmount: q.ConvertedAmount,
		Rate:            q.Rate,
		Status:          q.Status,
	}
}

type fxConversionResponse struct {
	ID                   string `json:"id"`
	QuoteID              string `json:"quote_id"`
	FromCurrency         string `json:"from_currency"`
	ToCurrency           string `json:"to_currency"`
	Amount               string `json:"amount"`
	ConvertedAmount      string `json:"converted_amount"`
	Rate                 string `json:"rate"`
	FromLedgerTransferID string `json:"from_ledger_transfer_id"`
	ToLedgerTransferID   string `json:"to_ledger_transfer_id"`
	Status               string `json:"status"`
}

func toFXConversionResponse(c domain.FXConversion) fxConversionResponse {
	return fxConversionResponse{
		ID:                   c.ID,
		QuoteID:              c.QuoteID,
		FromCurrency:         c.FromCurrency,
		ToCurrency:           c.ToCurrency,
		Amount:               c.Amount,
		ConvertedAmount:      c.ConvertedAmount,
		Rate:                 c.Rate,
		FromLedgerTransferID: c.FromLedgerTransferID,
		ToLedgerTransferID:   c.ToLedgerTransferID,
		Status:               c.Status,
	}
}
