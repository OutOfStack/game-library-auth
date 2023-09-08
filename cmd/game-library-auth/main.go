package main

import (
	_ "expvar"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	"github.com/OutOfStack/game-library-auth/internal/server"
	conf "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

func main() {
	var cfg appconf.Cfg
	if err := conf.Load(".", "app", "env", &cfg); err != nil {
		log.Fatalf("can't parse config: %v", err)
	}
	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("can't init logger: %v", err)
	}
	defer func(logger *zap.Logger) {
		if err = logger.Sync(); err != nil {
			log.Printf("can't sync logger: %v", err)
		}
	}(logger)

	if err = run(logger, cfg); err != nil {
		log.Fatal(err)
	}
}

func initLogger(cfg appconf.Cfg) (*zap.Logger, error) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	gelfWriter, err := gelf.NewTCPWriter(cfg.Graylog.Address)
	if err != nil {
		return nil, fmt.Errorf("can't create gelf writer: %v", err)
	}
	consoleWriter := zapcore.Lock(os.Stderr)

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(gelfWriter),
			zap.InfoLevel),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			consoleWriter,
			zap.InfoLevel))

	logger := zap.New(core, zap.WithCaller(false)).With(zap.String("source", appconf.ServiceName))

	return logger, nil
}

func run(logger *zap.Logger, cfg appconf.Cfg) error {
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
	defer db.Close()

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
