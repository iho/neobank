package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/iho/neobank/pkg/money"
	"github.com/iho/neobank/pkg/vendorsim"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/port"
	"github.com/iho/neobank/services/card/internal/usecase"
	"github.com/shopspring/decimal"
)

var centsPerUnit = decimal.NewFromInt(100)

// CardProcHandlers are the webhook endpoints the cardproc simulator drives:
// a synchronous real-time authorization decision, and an async consumer for
// capture/reversal events. See cmd/server/main.go for how each is mounted —
// the sync route verifies vendorsim signatures inline (below) rather than
// through vendorsim.VerifyWebhook, since that middleware's replay de-dup
// doesn't apply to a synchronous call-and-response; the async route uses it.
type CardProcHandlers struct {
	authorize  *usecase.AuthorizeTransactionUseCase
	capture    *usecase.CaptureAuthorizationUseCase
	reverse    *usecase.ReverseAuthorizationUseCase
	chargeback *usecase.ProcessChargebackWebhookUseCase
	secret     []byte
}

func NewCardProcHandlers(
	authorize *usecase.AuthorizeTransactionUseCase,
	capture *usecase.CaptureAuthorizationUseCase,
	reverse *usecase.ReverseAuthorizationUseCase,
	chargeback *usecase.ProcessChargebackWebhookUseCase,
	webhookSecret string,
) *CardProcHandlers {
	return &CardProcHandlers{authorize: authorize, capture: capture, reverse: reverse, chargeback: chargeback, secret: []byte(webhookSecret)}
}

type authorizeWebhookRequest struct {
	TransactionID        string `json:"transaction_id"`
	CardRef              string `json:"card_ref"`
	Amount               string `json:"amount"`
	Currency             string `json:"currency"`
	MerchantName         string `json:"merchant_name"`
	MerchantCategoryCode string `json:"merchant_category_code"`
}

type authorizeWebhookResponse struct {
	Decision        string `json:"decision"`
	AuthorizationID string `json:"authorization_id,omitempty"`
	ReasonCode      string `json:"reason_code,omitempty"`
}

// HandleAuthorize is the real-time authorization webhook: the cardproc
// simulator calls this synchronously and waits for approve/decline within
// its own request timeout, matching how a real card network's stand-in
// authorization flow behaves.
func (h *CardProcHandlers) HandleAuthorize(cards port.CardRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeAuthDecline(w, "", "invalid_request")
			return
		}

		if err := vendorsim.VerifySignature(h.secret, r.Header.Get(vendorsim.HeaderTimestamp), r.Header.Get(vendorsim.HeaderSignature), body, 5*time.Minute); err != nil {
			writeAuthDecline(w, "", "invalid_signature")
			return
		}

		var req authorizeWebhookRequest
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&req); err != nil {
			writeAuthDecline(w, "", "invalid_request")
			return
		}

		card, err := cards.GetByProcessorRef(r.Context(), req.CardRef)
		if err != nil || card == nil {
			writeAuthDecline(w, "", "card_not_found")
			return
		}

		// Magic-value convention (pkg/vendorsim): amount ending in .13 forces
		// a deterministic decline, for integration tests that don't want to
		// engineer a real decline condition (frozen card, limits, funds).
		if amt, err := money.Parse(req.Amount); err == nil {
			cents := amt.Mul(centsPerUnit).IntPart()
			if vendorsim.AmountEndsInCents(cents, 13) {
				writeAuthDecline(w, "", "forced_decline_magic_value")
				return
			}
		}

		out, err := h.authorize.Execute(r.Context(), usecase.AuthorizeTransactionInput{
			UserID:               card.UserID,
			CardID:               card.ID,
			Amount:               req.Amount,
			Currency:             req.Currency,
			MerchantName:         req.MerchantName,
			MerchantCategoryCode: req.MerchantCategoryCode,
			IdempotencyKey:       req.TransactionID,
		})
		if err != nil || out.Authorization == nil {
			writeAuthDecline(w, "", "internal_error")
			return
		}

		if out.Authorization.Status != domain.AuthStatusAuthorized {
			writeAuthDecline(w, out.Authorization.ID, out.Authorization.FailureReason)
			return
		}

		writeAPIJSON(w, http.StatusOK, authorizeWebhookResponse{
			Decision:        "approved",
			AuthorizationID: out.Authorization.ID,
		})
	}
}

func writeAuthDecline(w http.ResponseWriter, authorizationID, reasonCode string) {
	writeAPIJSON(w, http.StatusOK, authorizeWebhookResponse{
		Decision:        "declined",
		AuthorizationID: authorizationID,
		ReasonCode:      reasonCode,
	})
}

type cardEventWebhookPayload struct {
	AuthorizationID string `json:"authorization_id"`
	Reason          string `json:"reason,omitempty"`
}

// HandleEvents consumes the cardproc simulator's async lifecycle webhooks
// (capture, reversal, expiry, chargeback), mounted behind
// vendorsim.VerifyWebhook.
func (h *CardProcHandlers) HandleEvents(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get(vendorsim.HeaderEventType)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch eventType {
	case "card.captured":
		var payload cardEventWebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if _, err := h.capture.Execute(r.Context(), usecase.CaptureAuthorizationInput{
			AuthorizationID: payload.AuthorizationID,
		}); err != nil {
			writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	case "card.auth.reversed":
		var payload cardEventWebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if _, err := h.reverse.Execute(r.Context(), usecase.ReverseAuthorizationInput{
			AuthorizationID: payload.AuthorizationID,
			Reason:          payload.Reason,
		}); err != nil {
			writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	case "card.auth.expired":
		// A hold aging out unused is handled identically to an explicit
		// reversal (ReverseAuthorizationUseCase), just from a distinct
		// event type so audit/observability can tell the two apart.
		var payload cardEventWebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if _, err := h.reverse.Execute(r.Context(), usecase.ReverseAuthorizationInput{
			AuthorizationID: payload.AuthorizationID,
			Reason:          "expired",
		}); err != nil {
			writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	case "card.chargeback.opened":
		var payload usecase.ChargebackWebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if _, err := h.chargeback.Opened(r.Context(), payload); err != nil {
			writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	case "card.chargeback.won":
		if err := h.handleChargebackResolution(r.Context(), body, domain.DisputeStatusWon); err != nil {
			writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	case "card.chargeback.lost":
		if err := h.handleChargebackResolution(r.Context(), body, domain.DisputeStatusLost); err != nil {
			writeAPIError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	default:
		writeAPIError(w, http.StatusBadRequest, "unknown event type: "+eventType)
		return
	}

	writeAPIJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *CardProcHandlers) handleChargebackResolution(ctx context.Context, body []byte, outcome string) error {
	var payload usecase.ChargebackWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("invalid request body")
	}

	_, err := h.chargeback.Resolved(ctx, payload, outcome)

	return err
}

func writeAPIJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeAPIError(w http.ResponseWriter, status int, msg string) {
	writeAPIJSON(w, status, map[string]string{"error": msg})
}
