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

package money

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Parse validates and parses a decimal amount string.
func Parse(amount string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid amount %q: %w", amount, err)
	}

	if d.IsNegative() {
		return decimal.Zero, fmt.Errorf("amount must be positive: %s", amount)
	}

	return d, nil
}

// MustParse parses amount or panics (tests only).
func MustParse(amount string) decimal.Decimal {
	d, err := Parse(amount)
	if err != nil {
		panic(err)
	}

	return d
}
