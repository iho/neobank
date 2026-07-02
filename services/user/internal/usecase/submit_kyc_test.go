package usecase

import "testing"

func TestWalletProvisionKey(t *testing.T) {
	if got := walletProvisionKey("kyc-123", "user-1"); got != "kyc-123:wallet" {
		t.Fatalf("expected kyc-123:wallet, got %s", got)
	}
	if got := walletProvisionKey("", "user-1"); got != "wallet:user-1:USD" {
		t.Fatalf("expected wallet:user-1:USD, got %s", got)
	}
}