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
	"github.com/iho/neobank/pkg/sloghttp"
	"github.com/iho/neobank/pkg/vendorsim"
	apiadapter "github.com/iho/neobank/services/simulators/kyc/internal/adapter/api"
	"github.com/iho/neobank/services/simulators/kyc/internal/adapter/deliverystore"
	sqlcrepo "github.com/iho/neobank/services/simulators/kyc/internal/adapter/sqlc"
	"github.com/iho/neobank/services/simulators/kyc/internal/config"
	"github.com/iho/neobank/services/simulators/kyc/internal/gen/sqlc"
	"github.com/iho/neobank/services/simulators/kyc/internal/usecase"
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
	applicantRepo := sqlcrepo.NewApplicantRepository(queries)
	deliveryStore := deliverystore.NewPostgres(queries)

	dispatcher := vendorsim.NewDispatcher(deliveryStore, []byte(cfg.WebhookSecret), logger)
	dispatcher.Chaos = vendorsim.ChaosConfigFromEnv("KYC_CHAOS")

	dispatchCtx, cancelDispatch := context.WithCancel(ctx)
	defer cancelDispatch()

	go dispatcher.Run(dispatchCtx, time.Second)

	submitApplicantUC := usecase.NewSubmitApplicantUseCase(applicantRepo, dispatcher, cfg.EventsURL)
	resolveApplicantUC := usecase.NewResolveApplicantUseCase(applicantRepo, dispatcher, cfg.EventsURL)

	server := apiadapter.NewServer(submitApplicantUC, resolveApplicantUC, applicantRepo, deliveryStore)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Use(sloghttp.AccessLog(logger, sloghttp.WithService("kyc-simulator")))
	server.Mount(r)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("kyc simulator listening", "port", cfg.HTTPPort)
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
