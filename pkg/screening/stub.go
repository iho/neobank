//
// Copyright (c) 2026 Sumicare
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package screening

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StubScreener is an in-process MVP provider. It clears by default and blocks
// obvious test fixtures (country KP, name containing "SANCTIONED").
type StubScreener struct {
	now func() time.Time
}

func NewStubScreener() *StubScreener {
	return &StubScreener{now: func() time.Time { return time.Now().UTC() }}
}

func (s *StubScreener) ScreenOnboarding(subject Subject, ctx Context) (Result, error) {
	return s.evaluate(subject.UserID, subject.FullName, subject.CountryCode, ctx)
}

func (s *StubScreener) ScreenCounterparty(counterparty Counterparty, ctx Context) (Result, error) {
	if strings.Contains(strings.ToUpper(counterparty.Phone), "SANCTIONED") {
		return s.block(counterparty.UserID, "stub_counterparty_match", ctx)
	}

	return s.evaluate(counterparty.UserID, "", "", ctx)
}

func (s *StubScreener) evaluate(userID, fullName, country string, ctx Context) (Result, error) {
	now := s.now()
	ref := uuid.NewString()
	decision := DecisionClear
	reason := "stub_clear"

	upperName := strings.ToUpper(fullName)
	if strings.Contains(upperName, "SANCTIONED") || strings.EqualFold(country, "KP") {
		return s.block(userID, "stub_sanctions_match", ctx)
	}

	raw, _ := json.Marshal(map[string]any{
		"provider":     ProviderStub,
		"check_type":   ctx.CheckType,
		"user_id":      userID,
		"decision":     decision,
		"reason_code":  reason,
		"provider_ref": ref,
		"checked_at":   now.Format(time.RFC3339),
	})

	return Result{
		Decision:    decision,
		ReasonCode:  reason,
		Provider:    ProviderStub,
		ProviderRef: ref,
		RawResponse: raw,
		CheckedAt:   now,
	}, nil
}

func (s *StubScreener) block(userID, reason string, ctx Context) (Result, error) {
	now := s.now()
	ref := uuid.NewString()
	raw, _ := json.Marshal(map[string]any{
		"provider":     ProviderStub,
		"check_type":   ctx.CheckType,
		"user_id":      userID,
		"decision":     DecisionBlock,
		"reason_code":  reason,
		"provider_ref": ref,
		"checked_at":   now.Format(time.RFC3339),
	})

	return Result{
		Decision:    DecisionBlock,
		ReasonCode:  reason,
		Provider:    ProviderStub,
		ProviderRef: ref,
		RawResponse: raw,
		CheckedAt:   now,
	}, nil
}
