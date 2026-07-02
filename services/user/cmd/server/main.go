package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	"github.com/iho/neobank/pkg/idempotency"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/otel"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/screening"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/reqctx"
	"github.com/iho/neobank/pkg/sloghttp"
	"github.com/iho/neobank/pkg/vault"
	apiadapter "github.com/iho/neobank/services/user/internal/adapter/api"
	grpcadapter "github.com/iho/neobank/services/user/internal/adapter/grpc"
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

	piiProtector, err := vault.NewProtectorFromEnv()
	if err != nil {
		logger.Error("vault pii protector init failed", "error", err)
		os.Exit(1)
	}
	if piiProtector.Enabled() {
		logger.Info("vault transit pii encryption enabled")
	}

	queries := sqlc.New(pool)
	userRepo := sqlcrepo.NewUserRepository(queries, piiProtector)
	walletRepo := sqlcrepo.NewWalletRepository(queries)
	kycRepo := sqlcrepo.NewKYCRepository(queries, piiProtector)
	outboxRepo := sqlcrepo.NewOutboxRepository(queries)
	auditRepo := sqlcrepo.NewAuditRepository(queries)
	piiAccessRepo := sqlcrepo.NewPIIAccessRepository(queries)
	gdprRepo := sqlcrepo.NewGDPRRepository(queries, piiProtector)
	sagaStore := sqlcrepo.NewSagaStore(queries)
	walletTxRepo := sqlcrepo.NewWalletTransactionRepository(queries)
	inboxRepo := sqlcrepo.NewConsumerInboxRepository(queries)

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
	referralInviteRepo := sqlcrepo.NewReferralInviteRepository(queries)
	acceptInviteUC := usecase.NewAcceptReferralInviteUseCase(referralInviteRepo)
	registerUC := usecase.NewRegisterUseCase(userRepo, tokenIssuer, acceptInviteUC)
	loginUC := usecase.NewLoginUseCase(userRepo, tokenIssuer)
	refreshUC := usecase.NewRefreshTokenUseCase(tokenIssuer, userRepo)
	txRunner := pgutil.NewTxRunner(pool)
	provisionWalletUC := usecase.NewProvisionWalletUseCase(walletRepo, ledgerAdapter, outboxRepo, auditRepo, sagaStore, txRunner)
	submitKYCUC := usecase.NewSubmitKYCUseCase(kycRepo, provisionWalletUC, screening.NewStubScreener(), outboxRepo, auditRepo, txRunner)
	getKYCStatusUC := usecase.NewGetKYCStatusUseCase(kycRepo)
	getProfileUC := usecase.NewGetProfileUseCase(userRepo)
	walletBalanceUC := usecase.NewGetWalletBalanceUseCase(walletRepo, ledgerAdapter)
	projectWalletEventUC := usecase.NewProjectWalletEventUseCase(walletTxRepo, inboxRepo)
	listWalletTxUC := usecase.NewListWalletTransactionsUseCase(walletTxRepo)
	exportWalletTxUC := usecase.NewExportWalletTransactionsUseCase(walletTxRepo)
	listWalletsUC := usecase.NewListWalletsUseCase(walletRepo, ledgerAdapter)
	createInviteUC := usecase.NewCreateReferralInviteUseCase(referralInviteRepo)
	listInvitesUC := usecase.NewListReferralInvitesUseCase(referralInviteRepo)
	exportGDPRUC := usecase.NewExportGDPRUseCase(userRepo, gdprRepo)
	maskGDPRUC := usecase.NewMaskGDPRUseCase(userRepo, gdprRepo, auditRepo, txRunner)
	depositRepo := sqlcrepo.NewDepositRepository(queries)
	depositWalletUC := usecase.NewDepositWalletUseCase(
		walletRepo, depositRepo, walletTxRepo, outboxRepo, auditRepo, ledgerAdapter,
		cfg.DepositSourceAccountID, cfg.DepositMaxAmount, txRunner,
	)
	changePasswordUC := usecase.NewChangePasswordUseCase(userRepo, userRepo)
	payeeRepo := sqlcrepo.NewSavedPayeeRepository(queries)
	listPayeesUC := usecase.NewListPayeesUseCase(payeeRepo)
	createPayeeUC := usecase.NewCreatePayeeUseCase(payeeRepo)
	deletePayeeUC := usecase.NewDeletePayeeUseCase(payeeRepo)
	upsertPayeeUC := usecase.NewUpsertPayeeUseCase(payeeRepo)
	deviceTokenRepo := sqlcrepo.NewDeviceTokenRepository(queries)
	registerDeviceUC := usecase.NewRegisterDeviceTokenUseCase(deviceTokenRepo)
	deleteDeviceUC := usecase.NewDeleteDeviceTokenUseCase(deviceTokenRepo)
	listDeviceTokensUC := usecase.NewListDeviceTokensUseCase(deviceTokenRepo)
	closeAccountUC := usecase.NewCloseAccountUseCase(maskGDPRUC, userRepo, auditRepo)

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
		listWalletTxUC, exportWalletTxUC, listWalletsUC, createInviteUC, listInvitesUC, projectWalletEventUC,
		provisionWalletUC, exportGDPRUC, maskGDPRUC, depositWalletUC, changePasswordUC,
		listPayeesUC, createPayeeUC, deletePayeeUC, upsertPayeeUC,
		registerDeviceUC, deleteDeviceUC, listDeviceTokensUC, closeAccountUC,
		userRepo, walletRepo, piiAccessRepo,
	)
	strictHandler := genapi.NewStrictHandler(strictServer, nil)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Use(otel.HTTPMiddleware("user"))
	r.Use(reqctx.Middleware)
	r.Use(idempotency.Middleware(idempotency.NewStoreFromEnv(cfg.RedisURL, logger)))
	r.Use(sloghttp.AccessLog(logger, sloghttp.WithService("user")))
	genapi.HandlerFromMux(strictHandler, r)

	grpcServer, err := grpcutil.NewServer()
	if err != nil {
		logger.Error("grpc server init failed", "error", err)
		os.Exit(1)
	}
	if grpcutil.MTLSEnabled() {
		logger.Info("grpc mTLS enabled")
	}
	neobankv1.RegisterUserInternalServiceServer(grpcServer, grpcadapter.NewServer(
		userRepo, walletRepo, listDeviceTokensUC, upsertPayeeUC, piiAccessRepo,
	))
	neobankv1.RegisterUserServiceServer(grpcServer, grpcadapter.NewGatewayServer(strictServer))
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Error("grpc listen failed", "error", err)
		os.Exit(1)
	}
	go func() {
		logger.Info("user service gRPC listening", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("grpc server failed", "error", err)
			os.Exit(1)
		}
	}()

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
	grpcServer.GracefulStop()
	_ = srv.Shutdown(shutdownCtx)
}
