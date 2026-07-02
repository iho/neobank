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

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	notificationRepo := sqlcrepo.NewNotificationRepository(queries)

	ingestUC := usecase.NewIngestEventUseCase(notificationRepo)
	listUC := usecase.NewListNotificationsUseCase(notificationRepo)

	strictServer := apiadapter.NewServer(ingestUC, listUC)
	strictHandler := genapi.NewStrictHandler(strictServer, nil)

	if cfg.KafkaBrokers != "" {
		consumer := kafkaadapter.NewConsumer(cfg.KafkaBrokers, "notification-service", ingestUC, logger)
		go func() {
			if err := consumer.Run(ctx, "payment.events", "card.events"); err != nil && err != context.Canceled {
				logger.Error("kafka consumer stopped", "error", err)
			}
		}()
		logger.Info("kafka consumer enabled", "brokers", cfg.KafkaBrokers)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
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
	_ = srv.Shutdown(shutdownCtx)
}