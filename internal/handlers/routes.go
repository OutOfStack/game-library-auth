package handlers

import (
	"log"

	"github.com/OutOfStack/game-library-auth/internal/data/user"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers routes for auth service
func RegisterRoutes(app *fiber.App, db *sqlx.DB, log *log.Logger) {

	authAPI := AuthAPI{
		UserRepo: user.NewRepo(db),
		Log:      log,
	}

	app.Post("/signin", authAPI.signInHandler)
	app.Post("/signup", authAPI.signUpHandler)
}
