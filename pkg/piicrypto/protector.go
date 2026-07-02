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

import (
	"context"
	"strings"
)

// CiphertextPrefix marks values encrypted by Vault Transit in the database.
const CiphertextPrefix = "vault:v1:"

// Protector encrypts and decrypts field-level PII and derives blind indexes for lookup.
type Protector interface {
	Enabled() bool
	Encrypt(ctx context.Context, plaintext string) (string, error)
	Decrypt(ctx context.Context, stored string) (string, error)
	PhoneLookup(ctx context.Context, phone string) (string, error)
}

// NormalizePhone trims whitespace for consistent lookup indexes.
func NormalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}

// IsEncrypted reports whether stored looks like a Vault ciphertext blob.
func IsEncrypted(stored string) bool {
	return strings.HasPrefix(stored, CiphertextPrefix)
}