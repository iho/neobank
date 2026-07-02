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

package gdpr

import "fmt"

const (
	MaskedEmailDomain = "gdpr.invalid"
	RedactedValue     = "REDACTED"
	UserStatusMasked  = "masked"
)

// MaskedEmail returns a unique placeholder email that frees the original address
// for uniqueness constraints while keeping the user row for financial retention.
func MaskedEmail(userID string) string {
	return fmt.Sprintf("masked+%s@%s", userID, MaskedEmailDomain)
}