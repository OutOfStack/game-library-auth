package main

import (
	_ "expvar"
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/OutOfStack/game-library-auth/internal/handlers"
	"github.com/OutOfStack/game-library-auth/internal/server"
	conf "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	zaplog "github.com/OutOfStack/game-library-auth/pkg/log"
	"go.uber.org/zap"
)

// @title Game library auth API
// @version 0.3
// @description API for game library auth service
// @termsOfService http://swagger.io/terms/

// @host localhost:8001
// @BasePath /
// @query.collection.format multi
// @schemes http
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := conf.Load()
	if err != nil {
		log.Fatalf("can't parse config: %v", err)
	}

	// init logger
	logger := zaplog.New(cfg)
	defer func() {
		if err = logger.Sync(); err != nil {
			logger.Error("can't sync logger", zap.Error(err))
		}
	}()

	// connect to database
	db, err := database.New(cfg.DB.DSN)
	if err != nil {
		return fmt.Errorf("can't open db: %w", err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			logger.Error("close database", zap.Error(err))
		}
	}()

	// start debug service
	go func() {
		debugApp := handlers.DebugService()
		logger.Info("Debug service started", zap.String("address", cfg.Web.DebugAddress))
		err = server.Start(debugApp, cfg.Web.DebugAddress)
		if err != nil {
			logger.Info("Debug service stopped", zap.Error(err))
		}
	}()

	// start auth service
	app, err := handlers.Service(logger, db, cfg)
	if err != nil {
		return fmt.Errorf("creating auth service: %w", err)
	}
	logger.Info("Auth service started", zap.String("address", cfg.Web.Address))
	return server.StartWithGracefulShutdown(app, logger, cfg.Web.Address)
}
