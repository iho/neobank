package api

import (
	"encoding/json"
	"net/http"

	"github.com/iho/neobank/services/user/internal/usecase"
)

// KYCVendorHandlers consumes the KYC vendor simulator's async verdict
// webhook. Unlike the rails/cardproc webhooks, there's no synchronous path
// here — identity verification is inherently async, matching real vendors
// like Onfido/Sumsub.
type KYCVendorHandlers struct {
	processVerdict *usecase.ProcessKYCVerdictUseCase
}

func NewKYCVendorHandlers(processVerdict *usecase.ProcessKYCVerdictUseCase) *KYCVendorHandlers {
	return &KYCVendorHandlers{processVerdict: processVerdict}
}

type kycVerdictWebhookPayload struct {
	ApplicantID string `json:"applicant_id"`
	ExternalRef string `json:"external_ref"`
	Verdict     string `json:"verdict"`
	Reason      string `json:"reason,omitempty"`
}

// HandleKYCEvents handles POST /webhooks/kyc/events, mounted behind
// vendorsim.VerifyWebhook (see cmd/server/main.go).
func (h *KYCVendorHandlers) HandleKYCEvents(w http.ResponseWriter, r *http.Request) {
	var payload kycVerdictWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeKYCError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	kycCase, err := h.processVerdict.Execute(r.Context(), usecase.ProcessKYCVerdictInput{
		ApplicantID: payload.ApplicantID,
		Verdict:     payload.Verdict,
		Reason:      payload.Reason,
	})
	if err != nil {
		writeKYCError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeKYCJSON(w, http.StatusOK, map[string]string{
		"kyc_case_id": kycCase.ID,
		"status":      string(kycCase.Status),
	})
}

func writeKYCJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeKYCError(w http.ResponseWriter, status int, msg string) {
	writeKYCJSON(w, status, map[string]string{"error": msg})
}
