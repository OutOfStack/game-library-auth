package handlers

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

// AuthService creates and configures auth app
func AuthService(conf *appconf.Web, db *sqlx.DB, log *log.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "game-library-auth",
		ReadTimeout:  conf.ReadTimeout * time.Second,
		WriteTimeout: conf.WriteTimeout * time.Second,
	})

	// apply middleware
	app.Use(logger.New())

	// register routes
	RegisterRoutes(app, db, log)

	return app
}

// DebugService creates and configures debug app
func DebugService() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "pprof-debug",
	})

	// apply middleware
	app.Use(pprof.New())

	return app
}
