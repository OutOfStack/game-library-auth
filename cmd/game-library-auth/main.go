package main

import (
	"context"
	_ "expvar"
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	auth_ "github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/client/resendapi"
	store "github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	"github.com/OutOfStack/game-library-auth/internal/server"
	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	zaplog "github.com/OutOfStack/game-library-auth/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

// @title Game library auth API
// @version 0.4
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
	cfg, err := appconf.Get()
	if err != nil {
		log.Fatalf("can't parse config: %v", err)
	}

	// init logger
	logger := zaplog.New(cfg.Log.Level, cfg.Graylog.Address)
	defer func() {
		if err = logger.Sync(); err != nil {
			logger.Error("can't sync logger", zap.Error(err))
		}
	}()

	ctx := context.Background()

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

	// create user repo
	userRepo := store.NewUserRepo(db, logger)

	// create google token validator
	googleTokenValidator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return fmt.Errorf("create google token validator: %w", err)
	}

	// create email sender
	emailSender, err := resendapi.NewClient(resendapi.Config{
		APIToken:       cfg.EmailSender.APIToken,
		FromEmail:      cfg.EmailSender.EmailFrom,
		ContactEmail:   cfg.EmailSender.ContactEmail,
		BaseURL:        cfg.EmailSender.BaseURL,
		UnsubscribeURL: cfg.EmailSender.UnsubscribeURL,
		Timeout:        cfg.EmailSender.APITimeout,
	})
	if err != nil {
		return fmt.Errorf("create email sender client: %w", err)
	}

	// create auth token service
	privateKey, err := crypto.ReadPrivateKey(cfg.Auth.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("read private key file: %w", err)
	}
	auth, err := auth_.New(cfg.Auth.SigningAlgorithm, privateKey, cfg.Auth.Issuer, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)
	if err != nil {
		return fmt.Errorf("create token service instance: %w", err)
	}

	// create unsubscribe token generator
	unsubscribeTokenGenerator := auth_.NewUnsubscribeTokenGenerator([]byte(cfg.EmailSender.UnsubscribeSecret))

	// create user facade
	userFacade := facade.New(logger, userRepo, emailSender, auth, unsubscribeTokenGenerator)

	// auth api
	authAPI, err := handlers.NewAuthAPI(logger, googleTokenValidator, userFacade, handlers.AuthAPICfg{
		RefreshTokenCookieSameSite: cfg.Web.RefreshCookieSameSite,
		RefreshTokenCookieSecure:   cfg.Web.RefreshCookieSecure,
		GoogleOAuthClientID:        cfg.Auth.GoogleClientID,
		ContactEmail:               cfg.EmailSender.ContactEmail,
	})
	if err != nil {
		return fmt.Errorf("create auth api: %w", err)
	}

	// unsubscribe api
	unsubscribeAPI := handlers.NewUnsubscribeAPI(logger, unsubscribeTokenGenerator, userFacade, cfg.EmailSender.ContactEmail)

	// health api
	checkAPI := handlers.NewCheckAPI(db)

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
	app, err := handlers.Service(authAPI, checkAPI, unsubscribeAPI, cfg)
	if err != nil {
		return fmt.Errorf("creating auth service: %w", err)
	}
	logger.Info("Auth service started", zap.String("address", cfg.Web.Address))
	return server.StartWithGracefulShutdown(app, logger, cfg.Web.Address)
}
