// event-catalog prints the versioned domain event contract as JSON for compliance
// archives and consumer onboarding.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/iho/neobank/pkg/events"
)

func main() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(events.CatalogDocumentJSON()); err != nil {
		fmt.Fprintf(os.Stderr, "encode catalog: %v\n", err)
		os.Exit(1)
	}
}