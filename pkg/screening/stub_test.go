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

import "testing"

func TestStubScreenerBlocksSanctionedFixture(t *testing.T) {
	s := NewStubScreener()

	result, err := s.ScreenOnboarding(Subject{
		UserID: "u1", FullName: "SANCTIONED Person", CountryCode: "US",
	}, Context{CheckType: CheckOnboarding, EntityType: "kyc_case", EntityID: "c1"})
	if err != nil {
		t.Fatal(err)
	}

	if result.Decision != DecisionBlock {
		t.Fatalf("decision = %q, want block", result.Decision)
	}
}

func TestStubScreenerClearsNormalSubject(t *testing.T) {
	s := NewStubScreener()

	result, err := s.ScreenOnboarding(Subject{
		UserID: "u1", FullName: "Jane Doe", CountryCode: "US",
	}, Context{CheckType: CheckOnboarding, EntityType: "kyc_case", EntityID: "c1"})
	if err != nil {
		t.Fatal(err)
	}

	if result.Decision != DecisionClear {
		t.Fatalf("decision = %q, want clear", result.Decision)
	}
}
