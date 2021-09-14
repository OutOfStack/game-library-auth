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
	Login    string
	Password string
}

func (u *User) String() string {
	return fmt.Sprintf("Login: %s, Password: %s", u.Login, strings.Repeat("*", len(u.Password)))
}

func main() {
	app := fiber.New(fiber.Config{
		AppName: "game-library-auth",
		Views:   html.New("./views", ".html"),
	})

	app.Get("/", loginViewHandler)
	app.Get("/loggedin", loginSuccessViewHandler)
	app.Post("/login", loginHandler)

	log.Fatal(app.Listen(":3000"))
}

func loginViewHandler(c *fiber.Ctx) error {
	c.Render("login", nil)
	return nil
}

func loginSuccessViewHandler(c *fiber.Ctx) error {
	c.Render("loginSuccess", nil)
	return nil
}

func loginHandler(c *fiber.Ctx) error {
	fmt.Println("login handler")
	login := c.Context().FormValue("login")
	password := c.Context().FormValue("password")
	u := &User{
		Login:    string(login),
		Password: string(password),
	}
	fmt.Println(u.String())

	c.Redirect("loggedin", http.StatusFound)
	return nil
}
