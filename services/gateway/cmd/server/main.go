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
	"github.com/iho/neobank/pkg/grpcutil"
	"github.com/iho/neobank/pkg/idempotency"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/otel"
	"github.com/iho/neobank/pkg/reqctx"
	"github.com/iho/neobank/pkg/sloghttp"
	apiadapter "github.com/iho/neobank/services/gateway/internal/adapter/api"
	"github.com/iho/neobank/services/gateway/internal/client"
	gwmiddleware "github.com/iho/neobank/services/gateway/internal/middleware"
	"github.com/iho/neobank/services/gateway/internal/config"
	genapi "github.com/iho/neobank/services/gateway/internal/gen/api"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	shutdownOtel, err := otel.InitIfEnabled(ctx, "gateway")
	if err != nil {
		logger.Error("otel init failed", "error", err)
	} else if otel.Enabled() {
		logger.Info("otel tracing enabled", "endpoint", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}

	ledger, err := ledgerclient.New(ctx, ledgerclient.Config{Addr: cfg.LedgerAddr})
	if err != nil {
		logger.Warn("ledger not reachable (ok for local dev)", "error", err)
	} else {
		defer ledger.Close()
		logger.Info("connected to ledger", "addr", cfg.LedgerAddr)
	}

	if !cfg.AllowDevAuth {
		logger.Info("dev auth bypass disabled (X-User-Id header and legacy access.* tokens rejected)", "app_env", cfg.AppEnv)
	} else {
		logger.Warn("dev auth bypass enabled - do not use in production", "app_env", cfg.AppEnv)
	}

	jwtAuth := auth.NewJWT(cfg.JWTSecret)
	if grpcutil.MTLSEnabled() {
		logger.Info("grpc mTLS enabled for backend clients")
	}

	users, err := client.NewUserClient(ctx, client.Config{Addr: cfg.UserGRPCAddr})
	if err != nil {
		logger.Error("user service dial failed", "error", err)
		os.Exit(1)
	}
	defer users.Close()

	payments, err := client.NewPaymentClient(ctx, client.Config{Addr: cfg.PaymentGRPCAddr})
	if err != nil {
		logger.Error("payment service dial failed", "error", err)
		os.Exit(1)
	}
	defer payments.Close()

	cards, err := client.NewCardClient(ctx, client.Config{Addr: cfg.CardGRPCAddr})
	if err != nil {
		logger.Error("card service dial failed", "error", err)
		os.Exit(1)
	}
	defer cards.Close()

	notifications, err := client.NewNotificationClient(ctx, client.Config{Addr: cfg.NotificationGRPCAddr})
	if err != nil {
		logger.Error("notification service dial failed", "error", err)
		os.Exit(1)
	}
	defer notifications.Close()
	strictServer := apiadapter.NewServer(jwtAuth, users, payments, cards, notifications, cfg.AllowDevAuth)
	strictHandler := genapi.NewStrictHandler(strictServer, nil)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Use(otel.HTTPMiddleware("gateway"))
	r.Use(reqctx.Middleware)
	r.Use(gwmiddleware.Actor(jwtAuth, cfg.AllowDevAuth))
	r.Use(idempotency.Middleware(idempotency.NewStoreFromEnv(cfg.RedisURL, logger)))
	r.Use(sloghttp.AccessLog(logger, sloghttp.WithService("gateway")))
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
	_ = shutdownOtel(shutdownCtx)
	_ = srv.Shutdown(shutdownCtx)
}
