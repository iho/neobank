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
	"github.com/iho/neobank/pkg/auth"
	"github.com/iho/neobank/pkg/idempotency"
	"github.com/iho/neobank/pkg/ledgerclient"
	apiadapter "github.com/iho/neobank/services/gateway/internal/adapter/api"
	"github.com/iho/neobank/services/gateway/internal/client"
	"github.com/iho/neobank/services/gateway/internal/config"
	genapi "github.com/iho/neobank/services/gateway/internal/gen/api"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	ledger, err := ledgerclient.New(ctx, ledgerclient.Config{Addr: cfg.LedgerAddr})
	if err != nil {
		logger.Warn("ledger not reachable (ok for local dev)", "error", err)
	} else {
		defer ledger.Close()
		logger.Info("connected to ledger", "addr", cfg.LedgerAddr)
	}

	jwtAuth := auth.NewJWT(cfg.JWTSecret)
	users := client.NewUserClient(cfg.UserURL)
	payments := client.NewPaymentClient(cfg.PaymentURL)
	cards := client.NewCardClient(cfg.CardURL)
	notifications := client.NewNotificationClient(cfg.NotificationURL)
	strictServer := apiadapter.NewServer(jwtAuth, users, payments, cards, notifications)
	strictHandler := genapi.NewStrictHandler(strictServer, nil)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Use(idempotency.Middleware(idempotency.NewStoreFromEnv(cfg.RedisURL, logger)))
	genapi.HandlerFromMux(strictHandler, r)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("gateway listening", "port", cfg.HTTPPort)
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