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

import "testing"

func TestMaskedEmailIsUniquePerUser(t *testing.T) {
	a := MaskedEmail("550e8400-e29b-41d4-a716-446655440000")
	b := MaskedEmail("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if a == b {
		t.Fatalf("masked emails should differ: %q", a)
	}
	if a != "masked+550e8400-e29b-41d4-a716-446655440000@gdpr.invalid" {
		t.Fatalf("MaskedEmail() = %q", a)
	}
}