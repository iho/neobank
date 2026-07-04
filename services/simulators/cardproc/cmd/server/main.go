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
	apiadapter "github.com/iho/neobank/services/simulators/cardproc/internal/adapter/api"
	"github.com/iho/neobank/services/simulators/cardproc/internal/adapter/cardclient"
	"github.com/iho/neobank/services/simulators/cardproc/internal/adapter/deliverystore"
	sqlcrepo "github.com/iho/neobank/services/simulators/cardproc/internal/adapter/sqlc"
	"github.com/iho/neobank/services/simulators/cardproc/internal/config"
	"github.com/iho/neobank/services/simulators/cardproc/internal/gen/sqlc"
	"github.com/iho/neobank/services/simulators/cardproc/internal/usecase"
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
	cardRepo := sqlcrepo.NewCardRepository(queries)
	txRepo := sqlcrepo.NewTransactionRepository(queries)
	chargebackRepo := sqlcrepo.NewChargebackRepository(queries)
	deliveryStore := deliverystore.NewPostgres(queries)

	dispatcher := vendorsim.NewDispatcher(deliveryStore, []byte(cfg.WebhookSecret), logger)
	dispatcher.Chaos = vendorsim.ChaosConfigFromEnv("CARDPROC_CHAOS")

	dispatchCtx, cancelDispatch := context.WithCancel(ctx)
	defer cancelDispatch()

	go dispatcher.Run(dispatchCtx, time.Second)

	cardClient := cardclient.New(cardclient.Config{
		AuthorizeURL: cfg.AuthorizeURL,
		Secret:       []byte(cfg.WebhookSecret),
	})

	issueCardUC := usecase.NewIssueCardUseCase(cardRepo)
	simulateTxUC := usecase.NewSimulateTransactionUseCase(cardRepo, txRepo, cardClient, dispatcher, cfg.EventsURL)
	captureTxUC := usecase.NewCaptureTransactionUseCase(txRepo, dispatcher, cfg.EventsURL)
	reverseTxUC := usecase.NewReverseTransactionUseCase(txRepo, dispatcher, cfg.EventsURL)
	expireAuthsUC := usecase.NewExpireAuthorizationsUseCase(txRepo, dispatcher, cfg.EventsURL)
	openChargebackUC := usecase.NewOpenChargebackUseCase(txRepo, chargebackRepo, dispatcher, cfg.EventsURL)
	resolveChargebackUC := usecase.NewResolveChargebackUseCase(chargebackRepo, dispatcher, cfg.EventsURL)

	server := apiadapter.NewServer(issueCardUC, simulateTxUC, captureTxUC, reverseTxUC, openChargebackUC, resolveChargebackUC, cardRepo, chargebackRepo, deliveryStore)

	sweepCtx, cancelSweep := context.WithCancel(ctx)
	defer cancelSweep()
	go func() {
		ticker := time.NewTicker(cfg.AuthSweepInterval)
		defer ticker.Stop()

		for {
			select {
			case <-sweepCtx.Done():
				return
			case <-ticker.C:
				if n, err := expireAuthsUC.Sweep(sweepCtx, cfg.AuthTTL); err != nil {
					logger.Error("auth expiry sweep failed", "error", err)
				} else if n > 0 {
					logger.Info("auth expiry sweep", "expired", n)
				}
			}
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Use(sloghttp.AccessLog(logger, sloghttp.WithService("cardproc-simulator")))
	server.Mount(r)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("cardproc simulator listening", "port", cfg.HTTPPort)
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
