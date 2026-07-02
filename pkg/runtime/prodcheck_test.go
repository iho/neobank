package runtime

import "testing"

func TestRequireProductionSecrets(t *testing.T) {
	t.Parallel()
	if err := RequireProductionSecrets("development", ""); err != nil {
		t.Fatalf("dev should allow empty: %v", err)
	}
	if err := RequireProductionSecrets("production", "dev-secret-change-me"); err == nil {
		t.Fatal("expected error for default secret in production")
	}
	if err := RequireProductionSecrets("production", "real-secret"); err != nil {
		t.Fatalf("production with real secret: %v", err)
	}
}