package main

import (
	_ "expvar"
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	"github.com/OutOfStack/game-library-auth/internal/server"
	conf "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	lg "github.com/OutOfStack/game-library-auth/pkg/log"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// load config
	var cfg appconf.Cfg
	if err := conf.Load(".", "app", "env", &cfg); err != nil {
		log.Fatalf("can't parse config: %v", err)
	}

	// init logger
	logger, err := lg.InitLogger(cfg.Graylog.Address)
	if err != nil {
		log.Fatalf("can't init logger: %v", err)
	}
	defer func(logger *zap.Logger) {
		if err = logger.Sync(); err != nil {
			log.Printf("can't sync logger: %v", err)
		}
	}(logger)

	// connect to database
	db, err := database.Open(database.Config{
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		RequireSSL: cfg.DB.RequireSSL,
	})
	if err != nil {
		return fmt.Errorf("can't open db: %w", err)
	}
	defer func(db *sqlx.DB) {
		if err = db.Close(); err != nil {
			logger.Error("calling database close", zap.Error(err))
		}
	}(db)

	// start debug service
	debugApp := handlers.DebugService()
	go func() {
		logger.Info("Debug service started", zap.String("address", cfg.Web.DebugAddress))
		err = server.Start(debugApp, cfg.Web.DebugAddress)
		if err != nil {
			logger.Info("Debug service stopped", zap.Error(err))
		}
	}()

	// start auth service
	app, err := handlers.AuthService(logger, db, cfg.Web, cfg.Auth, cfg.Zipkin)
	if err != nil {
		return fmt.Errorf("creating auth service: %w", err)
	}
	logger.Info("Auth service started", zap.String("address", cfg.Web.Address))
	return server.StartWithGracefulShutdown(app, logger, cfg.Web.Address)
}
