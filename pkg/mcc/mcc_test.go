package mcc_test

import (
	"testing"

	"github.com/iho/neobank/pkg/mcc"
)

func TestCategoryLabel(t *testing.T) {
	if mcc.CategoryLabel("5411") != "Groceries" {
		t.Fatal("expected groceries")
	}
	if mcc.CategoryLabel("") != "Purchase" {
		t.Fatal("expected purchase default")
	}
	if mcc.CategoryLabel("9999") != "Retail" {
		t.Fatal("expected retail fallback")
	}
}