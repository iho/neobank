package integration

import (
	"net/http"
	"testing"

	"github.com/iho/neobank/pkg/gdpr"
)

func TestKYCRejectionAndResubmit(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "kyc-reject@example.com", "+15554000001")

	rejected := h.submitKYCWithName(t, user.UserID, "SANCTIONED Person", "US")
	if rejected.Status != "rejected" {
		t.Fatalf("first kyc status = %q", rejected.Status)
	}

	status := h.getKYCStatus(t, user.UserID)
	if status.Status != "rejected" || status.RejectionReason == "" {
		t.Fatalf("kyc status = %+v", status)
	}

	waitUntil(t, waitTimeout, func() bool {
		notifs := h.listNotifications(t, user.UserID)
		for _, n := range notifs.Notifications {
			if n.EventType == "user.kyc.rejected" {
				return true
			}
		}
		return false
	})

	approved := h.submitKYCWithName(t, user.UserID, "Clean User", "US")
	if approved.Status != "pending" {
		t.Fatalf("resubmit kyc status = %q, want pending (vendor call is async)", approved.Status)
	}

	waitUntil(t, waitTimeout, func() bool {
		return h.getKYCStatus(t, user.UserID).Status == "approved"
	})
}

func TestNotificationPreferencesMuteTransfers(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "mute-sender@example.com", "+15554000002")
	recipient := h.registerUser(t, "mute-recipient@example.com", "+15554000003")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)
	h.depositWallet(t, sender.UserID, "100.00", newID("mute-fund"))

	h.updateNotificationPreferences(t, sender.UserID, map[string]bool{"transfers": false})
	h.createTransferByEmail(t, sender.UserID, "mute-recipient@example.com", "10.00", newID("mute-p2p"))

	waitUntil(t, waitTimeout, func() bool {
		notifs := h.listNotifications(t, sender.UserID)
		for _, n := range notifs.Notifications {
			if n.EventType == "payment.transfer.completed" {
				t.Fatal("transfer notification should be muted")
			}
		}
		return true
	})
}

func TestDeviceTokenRegistration(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "device-user@example.com", "+15554000004")
	token := h.registerDeviceToken(t, user.UserID, "ios", "push-token-abc")
	if token.Platform != "ios" || token.Token != "push-token-abc" {
		t.Fatalf("device token = %+v", token)
	}

	status := (&httpClient{base: h.UserURL}).do(t, http.MethodDelete, "/api/v1/devices/"+token.ID, user.UserID, newID("del-device"), nil, nil)
	if status != http.StatusNoContent {
		t.Fatalf("delete device token status = %d", status)
	}
}

func TestAccountClosure(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "close-user@example.com", "+15554000005")
	h.submitKYC(t, user.UserID)
	h.depositWallet(t, user.UserID, "50.00", newID("close-fund"))
	wallet := h.getInternalWallet(t, user.UserID)
	card := h.issueCard(t, user.UserID, wallet.ID)

	if _, status := h.freezeCard(t, user.UserID, card.ID); status != http.StatusOK {
		t.Fatalf("freeze card status = %d", status)
	}

	h.closeAccount(t, user.UserID)

	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/auth/login", "", newID("login"), map[string]string{
		"email":    "close-user@example.com",
		"password": defaultPassword,
	}, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("login after close: status %d, want 401", status)
	}

	var cardStatus string
	err := h.pool.QueryRow(h.ctx, `SELECT status FROM card.cards WHERE id = $1`, card.ID).Scan(&cardStatus)
	if err != nil {
		t.Fatalf("load card: %v", err)
	}
	if cardStatus != "frozen" {
		t.Fatalf("card status = %q, want frozen", cardStatus)
	}

	var maskedEmail string
	err = h.pool.QueryRow(h.ctx, `SELECT email FROM "user".users WHERE id = $1`, user.UserID).Scan(&maskedEmail)
	if err != nil {
		t.Fatalf("load user: %v", err)
	}
	if maskedEmail != gdpr.MaskedEmail(user.UserID) {
		t.Fatalf("masked email = %q", maskedEmail)
	}
}

type kycStatusResponse struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type deviceTokenResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

type notificationPreferencesResponse struct {
	Transfers bool `json:"transfers"`
	Cards     bool `json:"cards"`
	KYC       bool `json:"kyc"`
	Push      bool `json:"push"`
	Email     bool `json:"email"`
}

func (h *Harness) getKYCStatus(t *testing.T, userID string) kycStatusResponse {
	t.Helper()
	var out kycStatusResponse
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodGet, "/api/v1/kyc/status", userID, "", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("kyc status: %d", status)
	}
	return out
}

func (h *Harness) updateNotificationPreferences(t *testing.T, userID string, prefs map[string]bool) notificationPreferencesResponse {
	t.Helper()
	var out notificationPreferencesResponse
	status := (&httpClient{base: h.NotificationURL}).do(t, http.MethodPatch, "/api/v1/notification-preferences", userID, "", prefs, &out)
	if status != http.StatusOK {
		t.Fatalf("update notification preferences: status %d", status)
	}
	return out
}

func (h *Harness) registerDeviceToken(t *testing.T, userID, platform, token string) deviceTokenResponse {
	t.Helper()
	var out deviceTokenResponse
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/devices", userID, newID("reg-device"), map[string]string{
		"platform": platform,
		"token":    token,
	}, &out)
	if status != http.StatusCreated {
		t.Fatalf("register device token: status %d", status)
	}
	return out
}

func (h *Harness) closeAccount(t *testing.T, userID string) {
	t.Helper()
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/account/close", userID, newID("close"), nil, nil)
	if status != http.StatusNoContent {
		t.Fatalf("close account: status %d", status)
	}
}

func (h *Harness) freezeCard(t *testing.T, userID, cardID string) (cardResponse, int) {
	t.Helper()
	var out cardResponse
	path := "/api/v1/cards/" + cardID + "/freeze"
	status := (&httpClient{base: h.CardURL}).do(t, http.MethodPost, path, userID, newID("freeze"), nil, &out)
	return out, status
}
