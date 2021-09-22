package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	cfg "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/gofiber/fiber/v2"
)

// SignIn describes login user.
type SignIn struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ErrResp describes error response.
type ErrResp struct {
	Error string `json:"error"`
}

func (s *SignIn) String() string {
	return fmt.Sprintf("Username: %s, Password: %s", s.Username, strings.Repeat("*", len(s.Password)))
}

type config struct {
	DB struct {
		Host       string `mapstructure:"DB_HOST"`
		Name       string `mapstructure:"DB_NAME"`
		User       string `mapstructure:"DB_USER"`
		Password   string `mapstructure:"DB_PASSWORD"`
		RequireSSL bool   `mapstructure:"DB_REQUIRESSL"`
	} `mapstructure:",squash"`
}

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{
		AppName: "game-library-auth",
	})

	app.Post("/signin", signInHandler)

	log.Fatal(app.Listen(":8081"))
}

func run() error {
	config := &config{}
	if err := cfg.LoadConfig(".", "app", "env", config); err != nil {
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

	return nil
}

func signInHandler(c *fiber.Ctx) error {
	var signIn SignIn
	if err := c.BodyParser(&signIn); err != nil {
		fmt.Printf("error parsing data: %v\n", err)
		if err := c.Status(http.StatusBadRequest).JSON(ErrResp{"Error parsing data"}); err != nil {
			return fmt.Errorf("error serializing data: %w", err)
		}
		return nil
	}

	resp := signIn.String()
	fmt.Println(resp)

	return c.JSON(resp)
}
