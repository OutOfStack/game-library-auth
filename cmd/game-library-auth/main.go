package main

import (
	"context"
	_ "expvar"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/api/grpc/authapi"
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	auth_ "github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/client/infoapi"
	"github.com/OutOfStack/game-library-auth/internal/client/resendapi"
	store "github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	zaplog "github.com/OutOfStack/game-library-auth/pkg/log"
	authpb "github.com/OutOfStack/game-library-auth/pkg/proto/authapi/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	// create infoapi client
	infoAPIClient, err := infoapi.NewClient(ctx, infoapi.Config{
		Address: cfg.InfoAPI.Address,
		Timeout: cfg.InfoAPI.Timeout,
		DialOptions: []grpc.DialOption{
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		},
	})
	if err != nil {
		return fmt.Errorf("create infoapi client: %w", err)
	}
	defer func() {
		if err = infoAPIClient.Close(); err != nil {
			logger.Error("close infoapi client", zap.Error(err))
		}
	}()

	// create user facade
	userFacade := facade.New(logger, userRepo, emailSender, auth, unsubscribeTokenGenerator, infoAPIClient)

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
	serverErrors := make(chan error, 3)

	go func() {
		debugApp := handlers.DebugService()
		logger.Info("Debug service started", zap.String("address", cfg.Web.DebugAddress))
		serverErrors <- debugApp.Listen(cfg.Web.DebugAddress)
	}()

	// start http service
	app, err := handlers.Service(authAPI, checkAPI, unsubscribeAPI, cfg)
	if err != nil {
		return fmt.Errorf("creating auth service: %w", err)
	}

	go func() {
		logger.Info("Auth service started", zap.String("address", cfg.Web.HTTPAddress))
		serverErrors <- app.Listen(cfg.Web.HTTPAddress)
	}()

	// start grpc service
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	authService := authapi.NewAuthService(logger, userFacade)
	authpb.RegisterAuthApiServiceServer(grpcServer, authService)
	// register reflection service for grpcurl and other tools
	reflection.Register(grpcServer)

	grpcListenConfig := net.ListenConfig{}
	listener, err := grpcListenConfig.Listen(ctx, "tcp", cfg.Web.GRPCAddress)
	if err != nil {
		return fmt.Errorf("failed to create gRPC listener: %w", err)
	}
	go func() {
		logger.Info("gRPC service started", zap.String("address", cfg.Web.GRPCAddress))
		serverErrors <- grpcServer.Serve(listener)
	}()

	// wait for shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))

		// stop grpc server
		grpcServer.GracefulStop()

		// stop http server
		if err = app.ShutdownWithTimeout(5 * time.Second); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		return nil
	}
}
