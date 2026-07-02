package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type schema struct {
	name  string
	qual  string
	outbox bool
	breaks bool
}

var schemas = []schema{
	{name: "user", qual: `"user"`, outbox: true, breaks: false},
	{name: "payment", qual: "payment", outbox: true, breaks: true},
	{name: "card", qual: "card", outbox: true, breaks: true},
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"
	}
	interval := envDuration("SCRAPE_INTERVAL", 30*time.Second)
	addr := envOr("HTTP_ADDR", ":9091")

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("database connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	reg := prometheus.NewRegistry()
	gauges := struct {
		outboxLag *prometheus.GaugeVec
		sagaOpen  *prometheus.GaugeVec
		breaks    *prometheus.GaugeVec
	}{
		outboxLag: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "neobank_outbox_oldest_unpublished_seconds",
			Help: "Age in seconds of the oldest unpublished outbox event",
		}, []string{"schema"}),
		sagaOpen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "neobank_saga_alerts_open",
			Help: "Count of open or investigating saga alerts",
		}, []string{"schema"}),
		breaks: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "neobank_reconciliation_breaks_open",
			Help: "Count of unresolved reconciliation breaks",
		}, []string{"schema"}),
	}
	reg.MustRegister(gauges.outboxLag, gauges.sagaOpen, gauges.breaks)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		scrape := func() {
			if err := refresh(ctx, pool, gauges); err != nil {
				logger.Error("metrics scrape failed", "error", err)
			}
		}
		scrape()
		for {
			select {
			case <-ticker.C:
				scrape()
			case <-ctx.Done():
				return
			}
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		logger.Info("ops-metrics listening", "addr", addr, "interval", interval)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func refresh(ctx context.Context, pool *pgxpool.Pool, g struct {
	outboxLag *prometheus.GaugeVec
	sagaOpen  *prometheus.GaugeVec
	breaks    *prometheus.GaugeVec
}) error {
	for _, s := range schemas {
		open, err := countSagaAlerts(ctx, pool, s.qual)
		if err != nil {
			return err
		}
		g.sagaOpen.WithLabelValues(s.name).Set(float64(open))

		if s.outbox {
			lag, err := outboxLag(ctx, pool, s.qual)
			if err != nil {
				return err
			}
			g.outboxLag.WithLabelValues(s.name).Set(lag)
		}
		if s.breaks {
			breaks, err := countBreaks(ctx, pool, s.qual)
			if err != nil {
				return err
			}
			g.breaks.WithLabelValues(s.name).Set(float64(breaks))
		}
	}
	return nil
}

func outboxLag(ctx context.Context, pool *pgxpool.Pool, qual string) (float64, error) {
	q := `
SELECT COALESCE(EXTRACT(EPOCH FROM (now() - MIN(e.created_at))), 0)
FROM ` + qual + `.outbox_events e
LEFT JOIN ` + qual + `.outbox_publications p ON p.event_id = e.id
WHERE p.event_id IS NULL`
	var lag float64
	err := pool.QueryRow(ctx, q).Scan(&lag)
	return lag, err
}

func countSagaAlerts(ctx context.Context, pool *pgxpool.Pool, qual string) (int, error) {
	q := `SELECT COUNT(*) FROM ` + qual + `.saga_alerts WHERE alert_status IN ('open', 'investigating')`
	var n int
	return n, pool.QueryRow(ctx, q).Scan(&n)
}

func countBreaks(ctx context.Context, pool *pgxpool.Pool, qual string) (int, error) {
	q := `SELECT COUNT(*) FROM ` + qual + `.reconciliation_breaks WHERE status = 'open'`
	var n int
	return n, pool.QueryRow(ctx, q).Scan(&n)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}