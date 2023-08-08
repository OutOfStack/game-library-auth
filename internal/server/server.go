package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Start starts service on specified address
func Start(app *fiber.App, address string) error {
	return app.Listen(address)
}

// StartWithGracefulShutdown starts service on specified address with graceful shutdown
func StartWithGracefulShutdown(app *fiber.App, log *zap.Logger, address string) error {
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- app.Listen(address)
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("listening and serving: %w", err)
	case <-shutdown:
		log.Info("Start shutdown")

		err := app.ShutdownWithTimeout(5 * time.Second)
		if err != nil {
			log.Error("Shutdown did not complete", zap.Error(err))
		}
	}

	return nil
}
