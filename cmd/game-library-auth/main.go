package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	cfg "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
)

type User struct {
	Username string
	Password string
}

type ViewData struct {
	ReturnUrl string
	Error     string
}

func (u *User) String() string {
	return fmt.Sprintf("Login: %s, Password: %s", u.Username, strings.Repeat("*", len(u.Password)))
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
		Views:   html.New("./views", ".html"),
	})

	app.Get("/", func(c *fiber.Ctx) error {
		c.Redirect("/login")
		return nil
	})
	app.Get("/login", loginViewHandler)
	app.Post("/login", loginHandler)

	log.Fatal(app.Listen(":3000"))
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

func loginViewHandler(c *fiber.Ctx) error {
	returnUrl := c.Query("returnUrl")
	viewData := &ViewData{
		ReturnUrl: returnUrl,
	}
	if returnUrl == "" {
		viewData.Error = "no returnUrl param provided"
	}
	c.Render("login", viewData)
	return nil
}

func loginHandler(c *fiber.Ctx) error {
	username := c.Context().FormValue("username")
	password := c.Context().FormValue("password")

	returnUrl := c.Query("returnUrl")

	if returnUrl == "" {
		viewData := &ViewData{
			Error: "no returnUrl param provided",
		}

		c.Render("login", viewData)
		return nil
	}

	u := &User{
		Username: string(username),
		Password: string(password),
	}
	fmt.Println(u.String())
	fmt.Printf("returnUrl: %s\n", returnUrl)

	c.Redirect(returnUrl, http.StatusFound)
	return nil
}
