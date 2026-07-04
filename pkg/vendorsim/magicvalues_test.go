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

import "testing"

func TestContainsToken(t *testing.T) {
	cases := []struct {
		s     string
		token string
		want  bool
	}{
		{"Jane REJECT Doe", "REJECT", true},
		{"jane reject doe", "REJECT", true},
		{"Jane Doe", "REJECT", false},
		{"", "REJECT", false},
	}

	for _, tc := range cases {
		if got := ContainsToken(tc.s, tc.token); got != tc.want {
			t.Errorf("ContainsToken(%q, %q) = %v, want %v", tc.s, tc.token, got, tc.want)
		}
	}
}

func TestAmountEndsInCents(t *testing.T) {
	cases := []struct {
		amountMinor int64
		cents       int
		want        bool
	}{
		{10013, 13, true},
		{10000, 13, false},
		{13, 13, true},
		{-10013, 13, true},
		{9999, 99, true},
	}

	for _, tc := range cases {
		if got := AmountEndsInCents(tc.amountMinor, tc.cents); got != tc.want {
			t.Errorf("AmountEndsInCents(%d, %d) = %v, want %v", tc.amountMinor, tc.cents, got, tc.want)
		}
	}
}
