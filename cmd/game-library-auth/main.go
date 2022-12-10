package main

import (
	_ "expvar"
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	"github.com/OutOfStack/game-library-auth/internal/server"
	cfg "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type config struct {
	DB     appconf.DB     `mapstructure:",squash"`
	Web    appconf.Web    `mapstructure:",squash"`
	Auth   appconf.Auth   `mapstructure:",squash"`
	Zipkin appconf.Zipkin `mapstructure:",squash"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	loggerCfg := zap.NewProductionConfig()
	loggerCfg.DisableCaller = true
	loggerCfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, err := loggerCfg.Build()
	if err != nil {
		return fmt.Errorf("can't initialize zap logger: %w", err)
	}
	defer logger.Sync()

	config := &config{}
	if err := cfg.Load(".", "app", "env", config); err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	// connect to database
	db, err := database.Open(database.Config{
		Host:       config.DB.Host,
		Name:       config.DB.Name,
		User:       config.DB.User,
		Password:   config.DB.Password,
		RequireSSL: config.DB.RequireSSL,
	})

	if err != nil {
		return fmt.Errorf("opening db: %w", err)
	}
	defer db.Close()

	// start debug service
	debugApp := handlers.DebugService()
	go func() {
		logger.Info("Debug service started", zap.String("address", config.Web.DebugAddress))
		err := server.Start(debugApp, config.Web.DebugAddress)
		logger.Info("Debug service stopped", zap.Error(err))
	}()

	// start auth service
	app, err := handlers.AuthService(logger, &config.Web, &config.Auth, &config.Zipkin, db)
	if err != nil {
		return fmt.Errorf("creating auth service: %w", err)
	}
	logger.Info("Auth service started", zap.String("address", config.Web.Address))
	return server.StartWithGracefulShutdown(app, logger, config.Web.Address)
}
