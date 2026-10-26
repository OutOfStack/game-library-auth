package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/api/tools"
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/client/resendapi"
	store "github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/notification"
	"github.com/OutOfStack/game-library-auth/internal/worker"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	zaplog "github.com/OutOfStack/game-library-auth/pkg/logger"
	"github.com/OutOfStack/game-library-auth/pkg/observability"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := appconf.Load[appconf.Worker]()
	if err != nil {
		return fmt.Errorf("load worker config: %w", err)
	}
	logger := zaplog.New(cfg.Log.Level, cfg.Graylog.Address)
	defer func() { _ = logger.Sync() }()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	tp, err := observability.InitTracer(logger, cfg.Jaeger.OTLPEndpoint, "game-library-auth-worker")
	if err != nil {
		return err
	}
	defer func() {
		traceCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(traceCtx)
	}()
	db, err := database.New(cfg.DB.DSN)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("close worker database", zap.Error(closeErr))
		}
	}()
	cipher, err := notification.NewCipher(cfg.Outbox.Key)
	if err != nil {
		return err
	}
	sender := resendapi.NewClient(cfg.EmailDelivery.APIToken, cfg.EmailDelivery.APITimeout)
	processor, err := worker.New(logger, store.NewUserRepo(db, logger), sender, cipher, prometheus.DefaultRegisterer)
	if err != nil {
		return err
	}
	hc := tools.NewHealthCheck(db)
	server := observability.NewDebugServer(cfg.Debug.Address, hc.Liveness, hc.Readiness)
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("listen for debug server: %w", err)
	}
	serverDone := make(chan error, 1)
	workerDone := make(chan error, 1)
	go func() { serverDone <- server.Serve(listener) }()
	go func() { workerDone <- processor.Run(ctx); stop() }()
	hc.SetReady(true)
	logger.Info("Notification worker started", zap.String("debug_address", server.Addr))
	select {
	case <-ctx.Done():
	case err = <-serverDone:
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
	}
	hc.SetReady(false)
	stop()
	// Run bounds the current attempt to 30 seconds; no further claims are made
	if workerErr := <-workerDone; err == nil {
		err = workerErr
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
		_ = server.Close()
		if err == nil {
			err = shutdownErr
		}
	}
	return err
}
