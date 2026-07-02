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
	"github.com/iho/neobank/pkg/idempotency"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/otel"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/screening"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/reqctx"
	"github.com/iho/neobank/pkg/sloghttp"
	apiadapter "github.com/iho/neobank/services/user/internal/adapter/api"
	"github.com/iho/neobank/services/user/internal/adapter/auth"
	kafkaadapter "github.com/iho/neobank/services/user/internal/adapter/kafka"
	ledgeradapter "github.com/iho/neobank/services/user/internal/adapter/ledger"
	sqlcrepo "github.com/iho/neobank/services/user/internal/adapter/sqlc"
	"github.com/iho/neobank/services/user/internal/config"
	genapi "github.com/iho/neobank/services/user/internal/gen/api"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/iho/neobank/services/user/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	shutdownOtel, err := otel.InitIfEnabled(ctx, "user")
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

	var ledgerConn *ledgerclient.Client
	ledger, err := ledgerclient.New(ctx, ledgerclient.Config{Addr: cfg.LedgerAddr})
	if err != nil {
		logger.Warn("ledger not reachable (wallet provisioning disabled)", "error", err)
	} else {
		ledgerConn = ledger
		defer ledger.Close()
		logger.Info("connected to ledger", "addr", cfg.LedgerAddr)
	}
	ledgerAdapter := ledgeradapter.New(ledgerConn)

	queries := sqlc.New(pool)
	userRepo := sqlcrepo.NewUserRepository(queries)
	walletRepo := sqlcrepo.NewWalletRepository(queries)
	kycRepo := sqlcrepo.NewKYCRepository(queries)
	outboxRepo := sqlcrepo.NewOutboxRepository(queries)
	auditRepo := sqlcrepo.NewAuditRepository(queries)
	sagaStore := sqlcrepo.NewSagaStore(queries)
	walletTxRepo := sqlcrepo.NewWalletTransactionRepository(queries)

	producer := outbox.NewProducer(outbox.ProducerConfig{
		KafkaBrokers:    cfg.KafkaBrokers,
		NotificationURL: cfg.NotificationURL,
		Logger:          logger,
	})
	outboxWorker := outbox.NewWorker(outboxRepo, producer, "user.events")
	go func() {
		if err := outboxWorker.Run(ctx); err != nil && err != context.Canceled {
			logger.Error("outbox worker stopped", "error", err)
		}
	}()

	tokenIssuer := auth.NewJWT(cfg.JWTSecret)
	registerUC := usecase.NewRegisterUseCase(userRepo, tokenIssuer)
	loginUC := usecase.NewLoginUseCase(userRepo, tokenIssuer)
	refreshUC := usecase.NewRefreshTokenUseCase(tokenIssuer, userRepo)
	txRunner := pgutil.NewTxRunner(pool)
	provisionWalletUC := usecase.NewProvisionWalletUseCase(walletRepo, ledgerAdapter, outboxRepo, auditRepo, sagaStore, txRunner)
	submitKYCUC := usecase.NewSubmitKYCUseCase(kycRepo, provisionWalletUC, screening.NewStubScreener(), outboxRepo, auditRepo, txRunner)
	getKYCStatusUC := usecase.NewGetKYCStatusUseCase(kycRepo)
	getProfileUC := usecase.NewGetProfileUseCase(userRepo)
	walletBalanceUC := usecase.NewGetWalletBalanceUseCase(walletRepo, ledgerAdapter)
	projectWalletEventUC := usecase.NewProjectWalletEventUseCase(walletTxRepo)
	listWalletTxUC := usecase.NewListWalletTransactionsUseCase(walletTxRepo)

	if cfg.KafkaBrokers != "" {
		consumer := kafkaadapter.NewConsumer(cfg.KafkaBrokers, "user-wallet-projection", projectWalletEventUC, logger)
		go func() {
			if err := consumer.Run(ctx, "payment.events", "card.events"); err != nil && err != context.Canceled {
				logger.Error("wallet projection consumer stopped", "error", err)
			}
		}()
		logger.Info("wallet projection kafka consumer enabled", "brokers", cfg.KafkaBrokers)
	}

	strictServer := apiadapter.NewServer(
		registerUC, loginUC, refreshUC, submitKYCUC, getKYCStatusUC, getProfileUC, walletBalanceUC,
		listWalletTxUC, projectWalletEventUC,
		provisionWalletUC, userRepo, walletRepo,
	)
	strictHandler := genapi.NewStrictHandler(strictServer, nil)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Use(otel.HTTPMiddleware("user"))
	r.Use(reqctx.Middleware)
	r.Use(idempotency.Middleware(idempotency.NewStoreFromEnv(cfg.RedisURL, logger)))
	r.Use(sloghttp.AccessLog(logger, sloghttp.WithService("user")))
	genapi.HandlerFromMux(strictHandler, r)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("user service listening", "port", cfg.HTTPPort)
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
