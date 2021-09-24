package main

import (
	_ "expvar"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/data/user"
	cfg "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

// ErrResp describes error response.
type ErrResp struct {
	Error string `json:"error"`
}

type config struct {
	DB  appconf.DB  `mapstructure:",squash"`
	Web appconf.Web `mapstructure:",squash"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log := log.New(os.Stdout, "AUTH : ", log.LstdFlags)

	config := &config{}
	if err := cfg.Load(".", "app", "env", config); err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

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
	go func() {
		app := fiber.New(fiber.Config{
			AppName: "pprof-debug",
		})
		app.Use(pprof.New())
		err := app.Listen(config.Web.DebugAddress)
		log.Printf("Debug service stopped %v\n", err)
	}()

	// start auth service
	app := fiber.New(fiber.Config{
		AppName:      "game-library-auth",
		ReadTimeout:  config.Web.ReadTimeout * time.Second,
		WriteTimeout: config.Web.WriteTimeout * time.Second,
	})

	app.Post("/signin", signInHandler)

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- app.Listen(config.Web.Address)
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

func signInHandler(c *fiber.Ctx) error {
	var signIn user.SignIn
	if err := c.BodyParser(&signIn); err != nil {
		fmt.Printf("error parsing data: %v\n", err)
		if err := c.Status(http.StatusBadRequest).JSON(ErrResp{"Error parsing data"}); err != nil {
			return fmt.Errorf("error serializing data: %w", err)
		}
		return nil
	}

	resp := "OK"
	fmt.Println(resp)

	return c.JSON(resp)
}
