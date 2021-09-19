package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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

func main() {
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
