package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

// Start starts service on specified address
func Start(app *fiber.App, address string) error {
	return app.Listen(address)
}

// StartWithGracefulShutdown starts service on specified address with graceful shutdown
func StartWithGracefulShutdown(app *fiber.App, address string) error {
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
		log.Println("Start shutdown")

		err := app.Shutdown()
		if err != nil {
			log.Printf("Shutdown did not complete: %v", err)
		}
	}

	return nil
}
