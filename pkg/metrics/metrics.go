package metrics

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var registry = prometheus.NewRegistry()

func init() {
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
}

// Registry exposes the application Prometheus registry.
func Registry() prometheus.Registerer { return registry }

// Handler serves /metrics from the application registry.
func Handler() http.Handler { return promhttp.HandlerFor(registry, promhttp.HandlerOpts{}) }

// Mount registers GET /metrics on the router.
func Mount(r chi.Router) {
	r.Method(http.MethodGet, "/metrics", Handler())
}