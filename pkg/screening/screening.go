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
	"time"
)

const (
	ProviderStub = "stub"

	CheckOnboarding           = "onboarding"
	CheckTransferCounterparty = "transfer_counterparty"

	DecisionClear  = "clear"
	DecisionReview = "review"
	DecisionBlock  = "block"
)

// Subject is the person screened during KYC onboarding.
type Subject struct {
	UserID      string
	FullName    string
	DateOfBirth string
	CountryCode string
}

// Counterparty is screened before a transfer to another user.
type Counterparty struct {
	UserID string
	Phone  string
}

// Context tags why and where a screening ran.
type Context struct {
	CheckType     string
	EntityType    string
	EntityID      string
	CorrelationID string
}

// Result is the provider outcome persisted for audit.
type Result struct {
	CheckedAt   time.Time
	Decision    string
	ReasonCode  string
	Provider    string
	ProviderRef string
	RawResponse json.RawMessage
}

// Screener evaluates subjects against sanctions/PEP lists.
type Screener interface {
	ScreenOnboarding(subject Subject, ctx Context) (Result, error)
	ScreenCounterparty(counterparty Counterparty, ctx Context) (Result, error)
}
