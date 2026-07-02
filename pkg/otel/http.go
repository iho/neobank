//
// Copyright (c) 2026 Sumicare
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

package otel

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/iho/neobank/pkg/reqctx"
)

// HTTPMiddleware traces inbound HTTP requests when OTEL_EXPORTER_OTLP_ENDPOINT is set.
func HTTPMiddleware(serviceName string) func(http.Handler) http.Handler {
	if !Enabled() {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, serviceName,
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
		)
	}
}

// OutboundTransport wraps the base round tripper with trace propagation and the
// existing correlation/actor headers from pkg/reqctx.
func OutboundTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if Enabled() {
		base = otelhttp.NewTransport(base)
	}

	return reqctx.Transport(base)
}