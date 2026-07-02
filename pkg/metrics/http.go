package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type httpMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func newHTTPMetrics(service string) *httpMetrics {
	factory := promauto.With(registry)
	return &httpMetrics{
		requests: factory.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total HTTP requests",
			ConstLabels: prometheus.Labels{"service": service},
		}, []string{"method", "route", "status"}),
		duration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request latency",
			ConstLabels: prometheus.Labels{"service": service},
			Buckets:     prometheus.DefBuckets,
		}, []string{"method", "route", "status"}),
	}
}

// HTTPMiddleware records RED metrics for inbound HTTP handlers.
func HTTPMiddleware(service string) func(http.Handler) http.Handler {
	m := newHTTPMetrics(service)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			route := r.URL.Path
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				if pattern := rctx.RoutePattern(); pattern != "" {
					route = pattern
				}
			}
			status := strconv.Itoa(ww.Status())
			m.requests.WithLabelValues(r.Method, route, status).Inc()
			m.duration.WithLabelValues(r.Method, route, status).Observe(time.Since(start).Seconds())
		})
	}
}