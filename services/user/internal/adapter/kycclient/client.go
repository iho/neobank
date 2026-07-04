// Package kycclient is a thin HTTP client for the KYC vendor simulator
// (services/simulators/kyc); a real identity-verification integration would
// implement the same usecase.KYCVendor interface behind this package.
package kycclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Config struct {
	BaseURL string
}

type Client struct {
	baseURL string
	http    *http.Client
}

func New(cfg Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type Applicant struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// SubmitApplicant asks the KYC vendor to run identity verification for a
// user; the verdict arrives later via webhook, not in this response.
func (c *Client) SubmitApplicant(ctx context.Context, externalRef, fullName, dateOfBirth, countryCode string) (Applicant, error) {
	body, err := json.Marshal(map[string]string{
		"external_ref":  externalRef,
		"full_name":     fullName,
		"date_of_birth": dateOfBirth,
		"country_code":  countryCode,
	})
	if err != nil {
		return Applicant{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/applicants", bytes.NewReader(body))
	if err != nil {
		return Applicant{}, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Applicant{}, fmt.Errorf("call kyc vendor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return Applicant{}, fmt.Errorf("kyc vendor returned status %d: %s", resp.StatusCode, respBody)
	}

	var applicant Applicant
	if err := json.NewDecoder(resp.Body).Decode(&applicant); err != nil {
		return Applicant{}, fmt.Errorf("decode response: %w", err)
	}

	return applicant, nil
}
