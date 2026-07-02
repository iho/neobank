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

package sloghttp

import (
	"context"
	"log/slog"

	"github.com/iho/neobank/pkg/reqctx"
)

// Logger returns a slog.Logger enriched with correlation_id and user_id from ctx.
// Use in handlers and workers so ad-hoc logs match access-log field names.
func Logger(ctx context.Context, base *slog.Logger) *slog.Logger {
	if base == nil {
		base = slog.Default()
	}

	attrs := make([]any, 0, 4)
	if correlationID := reqctx.CorrelationID(ctx); correlationID != "" {
		attrs = append(attrs, "correlation_id", correlationID)
	}

	if actor := reqctx.Actor(ctx); actor != "" && actor != "system" {
		attrs = append(attrs, "user_id", actor)
	}

	if len(attrs) == 0 {
		return base
	}

	return base.With(attrs...)
}
