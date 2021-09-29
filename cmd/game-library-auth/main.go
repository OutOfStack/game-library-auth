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
	cfg "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/database"
)

type config struct {
	DB   appconf.DB   `mapstructure:",squash"`
	Web  appconf.Web  `mapstructure:",squash"`
	Auth appconf.Auth `mapstructure:",squash"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger := log.New(os.Stdout, "AUTH : ", log.LstdFlags)

	config := &config{}
	if err := cfg.Load(".", "app", "env", config); err != nil {
		logger.Fatalf("error parsing config: %v", err)
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
		err := server.Start(debugApp, config.Web.DebugAddress)
		logger.Printf("Debug service stopped %v\n", err)
	}()

	// start auth service
	app, err := handlers.AuthService(&config.Web, &config.Auth, db, logger)
	if err != nil {
		return fmt.Errorf("starting auth service: %w", err)
	}
	return server.StartWithGracefulShutdown(app, config.Web.Address)
}
