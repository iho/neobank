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

// Store encrypts plaintext for persistence when non-empty.
func Store(ctx context.Context, p Protector, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	return p.Encrypt(ctx, plaintext)
}

// Read decrypts stored values, passing through legacy plaintext rows.
func Read(ctx context.Context, p Protector, stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	return p.Decrypt(ctx, stored)
}