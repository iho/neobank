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

package vendorsim

import "strings"

// Magic-value conventions shared across all simulators (rails, cardproc,
// kyc, fx), so integration tests can deterministically trigger an outcome
// through ordinary-looking input instead of a separate test-only API.
//
// Recommended per-simulator use of ContainsToken (case-insensitive substring
// match) against a free-text field:
//
//   - "REJECT" in a name/reference -> force a rejection/decline outcome
//   - "REVIEW" in a name/reference -> force a manual-review/pending outcome
//   - "RETURN" in a payment reference -> force a return/bounce after apparent success
//
// Recommended use of AmountEndsInCents against an amount in minor units:
//
//   - cents == 13 -> force a synchronous decline (e.g. card authorization)
//   - cents == 99 -> force an asynchronous failure after initial acceptance

// ContainsToken reports whether s contains token, case-insensitively.
func ContainsToken(s, token string) bool {
	return strings.Contains(strings.ToUpper(s), strings.ToUpper(token))
}

// AmountEndsInCents reports whether amountMinor's last two digits equal cents,
// letting tests pick an outcome via amount alone (e.g. 10013 ends in 13).
func AmountEndsInCents(amountMinor int64, cents int) bool {
	if amountMinor < 0 {
		amountMinor = -amountMinor
	}

	return int(amountMinor%100) == cents
}
