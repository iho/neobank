package integration

import (
	"net/http"
	"testing"

	"github.com/iho/neobank/pkg/gdpr"
)

func TestGDPRExportAndMask(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "gdpr-user@example.com", "+15551000888")
	h.submitKYC(t, user.UserID)

	export := h.exportGDPR(t, user.UserID)
	if export.Profile.Email != "gdpr-user@example.com" {
		t.Fatalf("export email = %q", export.Profile.Email)
	}
	if export.WalletTransactionCount < 0 {
		t.Fatal("expected wallet_transaction_count >= 0")
	}

	var exportCount int
	err := h.pool.QueryRow(h.ctx, `
		SELECT COUNT(*) FROM "user".gdpr_requests
		WHERE user_id = $1 AND request_type = 'export'`, user.UserID).Scan(&exportCount)
	if err != nil {
		t.Fatalf("count gdpr export requests: %v", err)
	}
	if exportCount != 1 {
		t.Fatalf("export request count = %d, want 1", exportCount)
	}

	mask := h.maskGDPR(t, user.UserID)
	if mask.Status != gdpr.UserStatusMasked {
		t.Fatalf("mask status = %q", mask.Status)
	}

	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/auth/login", "", newID("login"), map[string]string{
		"email":    "gdpr-user@example.com",
		"password": defaultPassword,
	}, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("login after mask: status %d, want 401", status)
	}

	var maskedEmail string
	err = h.pool.QueryRow(h.ctx, `SELECT email FROM "user".users WHERE id = $1`, user.UserID).Scan(&maskedEmail)
	if err != nil {
		t.Fatalf("load masked user: %v", err)
	}
	if maskedEmail != gdpr.MaskedEmail(user.UserID) {
		t.Fatalf("masked email = %q, want %q", maskedEmail, gdpr.MaskedEmail(user.UserID))
	}
}

type gdprExportResponse struct {
	UserID                 string `json:"user_id"`
	ExportedAt             string `json:"exported_at"`
	WalletTransactionCount int64  `json:"wallet_transaction_count"`
	Profile                struct {
		Email string `json:"email"`
	} `json:"profile"`
}

type gdprMaskResponse struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

func (h *Harness) exportGDPR(t *testing.T, userID string) gdprExportResponse {
	t.Helper()
	var out gdprExportResponse
	path := "/api/v1/internal/users/" + userID + "/gdpr/export"
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, path, "", newID("gdpr-export"), nil, &out)
	if status != http.StatusOK {
		t.Fatalf("gdpr export: status %d", status)
	}
	return out
}

func (h *Harness) maskGDPR(t *testing.T, userID string) gdprMaskResponse {
	t.Helper()
	var out gdprMaskResponse
	path := "/api/v1/internal/users/" + userID + "/gdpr/mask"
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, path, "", newID("gdpr-mask"), nil, &out)
	if status != http.StatusOK {
		t.Fatalf("gdpr mask: status %d", status)
	}
	return out
}