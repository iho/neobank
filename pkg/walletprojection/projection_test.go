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

package walletprojection

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/iho/neobank/pkg/events"
)

func TestApplyTransferCompleted(t *testing.T) {
	payload, _ := json.Marshal(events.TransferCompleted{
		TransferID: "t1", SenderUserID: "u1", RecipientUserID: "u2",
		Amount: "10.00", Currency: "USD",
	})
	occurred := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)

	rows, update, err := Apply(events.Envelope{
		EventID: "e1", EventType: events.TypeTransferCompleted, OccurredAt: occurred, Payload: payload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if update != nil {
		t.Fatal("expected no capture update")
	}

	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}

	if rows[0].Type != "p2p_out" || rows[0].UserID != "u1" {
		t.Fatalf("sender row = %+v", rows[0])
	}

	if rows[1].Type != "p2p_in" || rows[1].UserID != "u2" {
		t.Fatalf("recipient row = %+v", rows[1])
	}
}

func TestApplyCardAuthLifecycle(t *testing.T) {
	approvedPayload, _ := json.Marshal(events.CardAuthApproved{
		AuthorizationID: "a1", UserID: "u1", Amount: "5.00", Currency: "USD", MerchantName: "Coffee",
	})

	rows, _, err := Apply(events.Envelope{
		EventID: "e2", EventType: events.TypeCardAuthApproved, Payload: approvedPayload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 1 || rows[0].Type != "card_hold" || rows[0].Status != "authorized" {
		t.Fatalf("approved row = %+v", rows)
	}

	capturedPayload, _ := json.Marshal(events.CardAuthCaptured{
		AuthorizationID: "a1", UserID: "u1", Amount: "5.00", Currency: "USD",
	})

	_, update, err := Apply(events.Envelope{
		EventID: "e3", EventType: events.TypeCardAuthCaptured, Payload: capturedPayload,
	})
	if err != nil {
		t.Fatal(err)
	}

	if update == nil || update.Type != "card_purchase" || update.Status != "captured" {
		t.Fatalf("capture update = %+v", update)
	}
}
