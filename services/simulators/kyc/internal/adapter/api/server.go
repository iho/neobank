package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/iho/neobank/pkg/vendorsim"
	"github.com/iho/neobank/services/simulators/kyc/internal/domain"
	"github.com/iho/neobank/services/simulators/kyc/internal/port"
	"github.com/iho/neobank/services/simulators/kyc/internal/usecase"
)

type Server struct {
	submitApplicant  *usecase.SubmitApplicantUseCase
	resolveApplicant *usecase.ResolveApplicantUseCase
	applicants       port.ApplicantRepository
	deliveries       vendorsim.DeliveryStore
}

func NewServer(
	submitApplicant *usecase.SubmitApplicantUseCase,
	resolveApplicant *usecase.ResolveApplicantUseCase,
	applicants port.ApplicantRepository,
	deliveries vendorsim.DeliveryStore,
) *Server {
	return &Server{
		submitApplicant:  submitApplicant,
		resolveApplicant: resolveApplicant,
		applicants:       applicants,
		deliveries:       deliveries,
	}
}

func (s *Server) Mount(r chi.Router) {
	r.Get("/health", s.health)
	r.Post("/v1/applicants", s.createApplicant)
	r.Get("/v1/applicants/{id}", s.getApplicant)
	r.Post("/sim/reviews/{id}/resolve", s.resolveReview)
	r.Get("/sim/deliveries", s.listDeliveries)
	r.Get("/sim/deliveries/{id}", s.getDelivery)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "kyc-simulator"})
}

type createApplicantRequest struct {
	ExternalRef string `json:"external_ref"`
	FullName    string `json:"full_name"`
	DateOfBirth string `json:"date_of_birth"`
	CountryCode string `json:"country_code"`
}

type applicantResponse struct {
	ID          string `json:"id"`
	ExternalRef string `json:"external_ref"`
	FullName    string `json:"full_name"`
	DateOfBirth string `json:"date_of_birth"`
	CountryCode string `json:"country_code"`
	Status      string `json:"status"`
	Reason      string `json:"reason,omitempty"`
}

func (s *Server) createApplicant(w http.ResponseWriter, r *http.Request) {
	var req createApplicantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	applicant, err := s.submitApplicant.Execute(r.Context(), usecase.SubmitApplicantInput{
		ExternalRef: req.ExternalRef,
		FullName:    req.FullName,
		DateOfBirth: req.DateOfBirth,
		CountryCode: req.CountryCode,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toApplicantResponse(applicant))
}

func (s *Server) getApplicant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	applicant, err := s.applicants.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if applicant == nil {
		writeError(w, http.StatusNotFound, "applicant not found")
		return
	}

	writeJSON(w, http.StatusOK, toApplicantResponse(*applicant))
}

type resolveReviewRequest struct {
	Verdict string `json:"verdict"`
	Reason  string `json:"reason"`
}

func (s *Server) resolveReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req resolveReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	applicant, err := s.resolveApplicant.Execute(r.Context(), usecase.ResolveApplicantInput{
		ApplicantID: id,
		Verdict:     req.Verdict,
		Reason:      req.Reason,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toApplicantResponse(applicant))
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

func toApplicantResponse(a domain.Applicant) applicantResponse {
	return applicantResponse{
		ID:          a.ID,
		ExternalRef: a.ExternalRef,
		FullName:    a.FullName,
		DateOfBirth: a.DateOfBirth,
		CountryCode: a.CountryCode,
		Status:      a.Status,
		Reason:      a.Reason,
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
