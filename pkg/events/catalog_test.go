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

package events_test

import (
	"encoding/json"
	"testing"

	"github.com/iho/neobank/pkg/events"
)

func TestCatalogMatchesRegisteredEvents(t *testing.T) {
	seen := make(map[string]struct{}, len(events.Catalog()))
	for _, entry := range events.Catalog() {
		if entry.EventType == "" {
			t.Fatal("catalog entry missing event_type")
		}
		if _, dup := seen[entry.EventType]; dup {
			t.Fatalf("duplicate catalog entry %q", entry.EventType)
		}
		seen[entry.EventType] = struct{}{}
	}

	for _, evt := range events.RegisteredEvents() {
		entry, ok := events.LookupCatalog(evt.EventType())
		if !ok {
			t.Fatalf("missing catalog entry for %q", evt.EventType())
		}
		if entry.EventVersion != evt.Version() {
			t.Fatalf("%s version = %d, want %d", evt.EventType(), entry.EventVersion, evt.Version())
		}
		if entry.AggregateType != evt.AggregateType() {
			t.Fatalf("%s aggregate_type = %q, want %q", evt.EventType(), entry.AggregateType, evt.AggregateType())
		}

		payload, err := events.MarshalPayload(evt)
		if err != nil {
			t.Fatalf("marshal %s: %v", evt.EventType(), err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(payload, &fields); err != nil {
			t.Fatalf("unmarshal payload %s: %v", evt.EventType(), err)
		}
		for _, name := range entry.PayloadFields {
			if _, ok := fields[name]; !ok {
				t.Fatalf("%s payload missing field %q", evt.EventType(), name)
			}
		}
	}

	if len(events.RegisteredEvents()) != len(events.Catalog()) {
		t.Fatalf("registered events = %d, catalog entries = %d",
			len(events.RegisteredEvents()), len(events.Catalog()))
	}
}

func TestCatalogDocumentHasEnvelopeSpec(t *testing.T) {
	doc := events.CatalogDocumentJSON()
	if doc.CatalogVersion == "" {
		t.Fatal("catalog_version required")
	}
	if len(doc.Envelope.Required) == 0 {
		t.Fatal("envelope required fields required")
	}
	if len(doc.Events) == 0 {
		t.Fatal("events required")
	}
}