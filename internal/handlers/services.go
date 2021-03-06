package handlers

import (
	"fmt"
	"log"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/jmoiron/sqlx"
)

// AuthService creates and configures auth app
func AuthService(conf *appconf.Web, authConf *appconf.Auth, db *sqlx.DB, log *log.Logger) (*fiber.App, error) {
	app := fiber.New(fiber.Config{
		AppName:      "game-library-auth",
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	})

	privateKey, err := crypto.ReadPrivateKey(authConf.PrivateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("reading private key file: %w", err)
	}
	a, err := auth.New(authConf.SigningAlgorithm, privateKey)
	if err != nil {
		return nil, fmt.Errorf("initializing token service instance: %w", err)
	}

	// apply middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: conf.AllowedCORSOrigin,
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "POST, GET, OPTIONS",
	}))

	// register routes
	RegisterRoutes(app, authConf, db, a, log)

	return app, nil
}

// DebugService creates and configures debug app
func DebugService() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "debug",
	})

	// apply middleware
	app.Use(pprof.New())

	return app
}
