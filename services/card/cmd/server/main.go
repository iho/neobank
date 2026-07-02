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
	"github.com/iho/neobank/pkg/otel"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/reqctx"
	"github.com/iho/neobank/pkg/sloghttp"
	"github.com/iho/neobank/pkg/userclient"
	apiadapter "github.com/iho/neobank/services/card/internal/adapter/api"
	"github.com/iho/neobank/services/card/internal/adapter/processor"
	sqlcrepo "github.com/iho/neobank/services/card/internal/adapter/sqlc"
	"github.com/iho/neobank/services/card/internal/config"
	genapi "github.com/iho/neobank/services/card/internal/gen/api"
	"github.com/iho/neobank/services/card/internal/gen/sqlc"
	"github.com/iho/neobank/services/card/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	shutdownOtel, err := otel.InitIfEnabled(ctx, "card")
	if err != nil {
		logger.Error("otel init failed", "error", err)
	} else if otel.Enabled() {
		logger.Info("otel tracing enabled", "endpoint", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	var ledger *ledgerclient.Client
	ledgerConn, err := ledgerclient.New(ctx, ledgerclient.Config{Addr: cfg.LedgerAddr})
	if err != nil {
		logger.Warn("ledger not reachable (card auth will fail)", "error", err)
	} else {
		ledger = ledgerConn
		defer ledger.Close()
		logger.Info("connected to ledger", "addr", cfg.LedgerAddr)
	}
	if cfg.SettlementLedgerAcctID == "" {
		logger.Warn("SETTLEMENT_LEDGER_ACCOUNT_ID not set (capture will fail)")
	}

	queries := sqlc.New(pool)
	cardRepo := sqlcrepo.NewCardRepository(queries)
	authRepo := sqlcrepo.NewAuthorizationRepository(queries)
	outboxRepo := sqlcrepo.NewOutboxRepository(queries)
	auditRepo := sqlcrepo.NewAuditRepository(queries)
	fraudRepo := sqlcrepo.NewFraudDecisionRepository(queries)
	sagaStore := sqlcrepo.NewSagaStore(queries)

	users, err := userclient.New(ctx, userclient.Config{Addr: cfg.UserGRPCAddr})
	if err != nil {
		logger.Error("user service connect failed", "error", err)
		os.Exit(1)
	}
	defer users.Close()
	proc := processor.NewMock()
	fraudChecker := fraud.NewChecker()

	txRunner := pgutil.NewTxRunner(pool)
	issueUC := usecase.NewIssueCardUseCase(cardRepo, users, proc, fraudChecker, fraudRepo, outboxRepo, auditRepo, sagaStore, txRunner)
	freezeUC := usecase.NewFreezeCardUseCase(cardRepo, outboxRepo, auditRepo, txRunner)
	unfreezeUC := usecase.NewUnfreezeCardUseCase(cardRepo, outboxRepo, auditRepo, txRunner)
	updateControlsUC := usecase.NewUpdateCardControlsUseCase(cardRepo)
	authorizeUC := usecase.NewAuthorizeTransactionUseCase(cardRepo, authRepo, users, ledger, fraudChecker, fraudRepo, outboxRepo, auditRepo, sagaStore, txRunner)
	captureUC := usecase.NewCaptureAuthorizationUseCase(authRepo, ledger, outboxRepo, auditRepo, cfg.SettlementLedgerAcctID, txRunner)
	listAuthsUC := usecase.NewListAuthorizationsUseCase(authRepo)

	strictServer := apiadapter.NewServer(issueUC, freezeUC, unfreezeUC, updateControlsUC, authorizeUC, captureUC, listAuthsUC)
	strictHandler := genapi.NewStrictHandler(strictServer, nil)

	producer := outbox.NewProducer(outbox.ProducerConfig{
		KafkaBrokers:    cfg.KafkaBrokers,
		NotificationURL: cfg.NotificationURL,
		ProjectionURLs:  []string{outbox.WalletProjectionURL(cfg.UserURL)},
		Logger:          logger,
	})
	outboxWorker := outbox.NewWorker(outboxRepo, producer, "card.events")
	go func() {
		if err := outboxWorker.Run(ctx); err != nil && err != context.Canceled {
			logger.Error("outbox worker stopped", "error", err)
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Use(otel.HTTPMiddleware("card"))
	r.Use(reqctx.Middleware)
	r.Use(idempotency.Middleware(idempotency.NewStoreFromEnv(cfg.RedisURL, logger)))
	r.Use(sloghttp.AccessLog(logger, sloghttp.WithService("card")))
	genapi.HandlerFromMux(strictHandler, r)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("card service listening", "port", cfg.HTTPPort)
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
	_ = shutdownOtel(shutdownCtx)
	_ = srv.Shutdown(shutdownCtx)
}
