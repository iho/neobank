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
	"github.com/iho/neobank/pkg/notify"
	"github.com/iho/neobank/pkg/otel"
	"github.com/iho/neobank/pkg/reqctx"
	"github.com/iho/neobank/pkg/sloghttp"
	"github.com/iho/neobank/pkg/userclient"
	apiadapter "github.com/iho/neobank/services/notification/internal/adapter/api"
	kafkaadapter "github.com/iho/neobank/services/notification/internal/adapter/kafka"
	sqlcrepo "github.com/iho/neobank/services/notification/internal/adapter/sqlc"
	"github.com/iho/neobank/services/notification/internal/config"
	genapi "github.com/iho/neobank/services/notification/internal/gen/api"
	"github.com/iho/neobank/services/notification/internal/gen/sqlc"
	"github.com/iho/neobank/services/notification/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	shutdownOtel, err := otel.InitIfEnabled(ctx, "notification")
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

	queries := sqlc.New(pool)
	notificationRepo := sqlcrepo.NewNotificationRepository(queries)
	inboxRepo := sqlcrepo.NewConsumerInboxRepository(queries)
	prefsRepo := sqlcrepo.NewPreferencesRepository(queries)

	userClient, err := userclient.New(ctx, userclient.Config{Addr: cfg.UserGRPCAddr})
	if err != nil {
		logger.Error("user service connect failed", "error", err)
		os.Exit(1)
	}
	defer userClient.Close()
	delivery := usecase.NewDeliveryService(
		notify.NewLogDispatcher(logger),
		usecase.NewUserClientDelivery(userClient),
		usecase.NewUserClientDelivery(userClient),
		prefsRepo,
	)

	ingestUC := usecase.NewIngestEventUseCase(notificationRepo, inboxRepo, prefsRepo, delivery)
	listUC := usecase.NewListNotificationsUseCase(notificationRepo)
	markReadUC := usecase.NewMarkNotificationReadUseCase(notificationRepo)
	markAllUC := usecase.NewMarkAllNotificationsReadUseCase(notificationRepo)
	getPrefsUC := usecase.NewGetNotificationPreferencesUseCase(prefsRepo)
	updatePrefsUC := usecase.NewUpdateNotificationPreferencesUseCase(prefsRepo)

	strictServer := apiadapter.NewServer(ingestUC, listUC, markReadUC, markAllUC, getPrefsUC, updatePrefsUC)
	strictHandler := genapi.NewStrictHandler(strictServer, nil)

	if cfg.KafkaBrokers != "" {
		consumer := kafkaadapter.NewConsumer(cfg.KafkaBrokers, "notification-service", ingestUC, logger)
		go func() {
			if err := consumer.Run(ctx, "payment.events", "card.events", "user.events"); err != nil && err != context.Canceled {
				logger.Error("kafka consumer stopped", "error", err)
			}
		}()
		logger.Info("kafka consumer enabled", "brokers", cfg.KafkaBrokers)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Use(otel.HTTPMiddleware("notification"))
	r.Use(reqctx.Middleware)
	r.Use(sloghttp.AccessLog(logger, sloghttp.WithService("notification")))
	genapi.HandlerFromMux(strictHandler, r)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("notification service listening", "port", cfg.HTTPPort)
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
