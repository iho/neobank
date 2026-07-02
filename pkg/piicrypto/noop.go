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

package piicrypto

import "context"

// Noop stores plaintext and uses the normalized phone as its own lookup key.
// Used when Vault is not configured (local dev, integration tests).
type Noop struct{}

func NewNoop() *Noop { return &Noop{} }

func (n *Noop) Enabled() bool { return false }

func (n *Noop) Encrypt(_ context.Context, plaintext string) (string, error) {
	return plaintext, nil
}

func (n *Noop) Decrypt(_ context.Context, stored string) (string, error) {
	return stored, nil
}

func (n *Noop) PhoneLookup(_ context.Context, phone string) (string, error) {
	normalized := NormalizePhone(phone)
	if normalized == "" {
		return "", nil
	}
	return normalized, nil
}