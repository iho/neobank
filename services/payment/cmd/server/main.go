package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/pkg/idempotency"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/reqctx"
	"github.com/iho/neobank/pkg/sloghttp"
	"github.com/iho/neobank/pkg/screening"
	"github.com/iho/neobank/pkg/userclient"
	apiadapter "github.com/iho/neobank/services/payment/internal/adapter/api"
	sqlcrepo "github.com/iho/neobank/services/payment/internal/adapter/sqlc"
	"github.com/iho/neobank/services/payment/internal/config"
	genapi "github.com/iho/neobank/services/payment/internal/gen/api"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/iho/neobank/services/payment/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	var ledger *ledgerclient.Client
	ledgerConn, err := ledgerclient.New(ctx, ledgerclient.Config{Addr: cfg.LedgerAddr})
	if err != nil {
		logger.Warn("ledger not reachable (transfers will fail)", "error", err)
	} else {
		ledger = ledgerConn
		defer ledger.Close()
		logger.Info("connected to ledger", "addr", cfg.LedgerAddr)
	}

	queries := sqlc.New(pool)
	transferRepo := sqlcrepo.NewTransferRepository(queries)
	outboxRepo := sqlcrepo.NewOutboxRepository(queries)
	auditRepo := sqlcrepo.NewAuditRepository(queries)
	fraudRepo := sqlcrepo.NewFraudDecisionRepository(queries)
	sagaStore := sqlcrepo.NewSagaStore(queries)

	users := userclient.New(cfg.UserURL)
	fraudChecker := fraud.NewChecker()
	txRunner := pgutil.NewTxRunner(pool)
	screeningRepo := sqlcrepo.NewScreeningRepository(queries)
	p2pUC := usecase.NewP2PTransferUseCase(transferRepo, users, ledger, fraudChecker, fraudRepo, screeningRepo, screening.NewStubScreener(), outboxRepo, auditRepo, sagaStore, txRunner)

	strictServer := apiadapter.NewServer(p2pUC)
	strictHandler := genapi.NewStrictHandler(strictServer, nil)

	producer := outbox.NewProducer(outbox.ProducerConfig{
		KafkaBrokers:    cfg.KafkaBrokers,
		NotificationURL: cfg.NotificationURL,
		ProjectionURLs:  []string{outbox.WalletProjectionURL(cfg.UserURL)},
		Logger:          logger,
	})
	outboxWorker := outbox.NewWorker(outboxRepo, producer, "payment.events")
	go func() {
		if err := outboxWorker.Run(ctx); err != nil && err != context.Canceled {
			logger.Error("outbox worker stopped", "error", err)
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Use(reqctx.Middleware)
	r.Use(idempotency.Middleware(idempotency.NewStoreFromEnv(cfg.RedisURL, logger)))
	r.Use(sloghttp.AccessLog(logger, sloghttp.WithService("payment")))
	genapi.HandlerFromMux(strictHandler, r)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("payment service listening", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
