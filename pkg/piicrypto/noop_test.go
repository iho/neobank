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

import "testing"

func TestNoopRoundTrip(t *testing.T) {
	p := NewNoop()
	ctx := t.Context()

	stored, err := Store(ctx, p, "+15551234567")
	if err != nil {
		t.Fatal(err)
	}
	if stored != "+15551234567" {
		t.Fatalf("stored = %q", stored)
	}
	plain, err := Read(ctx, p, stored)
	if err != nil || plain != "+15551234567" {
		t.Fatalf("read = %q, err = %v", plain, err)
	}
	lookup, err := p.PhoneLookup(ctx, "+15551234567")
	if err != nil || lookup != "+15551234567" {
		t.Fatalf("lookup = %q, err = %v", lookup, err)
	}
}