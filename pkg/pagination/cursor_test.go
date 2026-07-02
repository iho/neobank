package pagination

import (
	"testing"
	"time"
)

func TestCursorRoundTrip(t *testing.T) {
	at := time.Date(2026, 3, 15, 12, 30, 0, 123456789, time.UTC)
	encoded := Encode(Cursor{CreatedAt: at, ID: "abc-123"})
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !decoded.CreatedAt.Equal(at) || decoded.ID != "abc-123" {
		t.Fatalf("decoded = %+v", decoded)
	}
}

func TestDecodeInvalid(t *testing.T) {
	if _, err := Decode("not-valid"); err == nil {
		t.Fatal("expected error")
	}
}