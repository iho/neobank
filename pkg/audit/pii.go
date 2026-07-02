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

package audit

import (
	"context"

	"github.com/iho/neobank/pkg/reqctx"
)

// PII resource identifiers recorded when customer data is read (distinct from
// write-side audit_log lifecycle entries).
const (
	PIIResourceProfile            = "profile"
	PIIResourceKYCStatus          = "kyc_status"
	PIIResourceWalletBalance      = "wallet_balance"
	PIIResourceWalletTransactions = "wallet_transactions"
	PIIResourceUserByPhone        = "user_by_phone"
	PIIResourceInternalWallet     = "internal_wallet"
	PIIResourceGDPRExport         = "gdpr_export"
)

// PIIAccessEntry is one append-only record of a successful read of customer PII.
type PIIAccessEntry struct {
	Metadata      map[string]any
	SubjectUserID string
	Resource      string
	Actor         string
	CorrelationID string
}

// PIIAccessRecorder persists read-access audit entries outside domain transactions.
type PIIAccessRecorder interface {
	RecordPIIAccess(ctx context.Context, e PIIAccessEntry) error
}

// ResolvePIIAccess fills Actor/CorrelationID from ctx when the caller didn't set them.
func ResolvePIIAccess(ctx context.Context, e PIIAccessEntry) PIIAccessEntry {
	if e.Actor == "" {
		e.Actor = reqctx.Actor(ctx)
	}
	if e.Actor == "" {
		e.Actor = "system"
	}

	if e.CorrelationID == "" {
		e.CorrelationID = reqctx.CorrelationID(ctx)
	}

	if e.Metadata == nil {
		e.Metadata = map[string]any{}
	}

	return e
}