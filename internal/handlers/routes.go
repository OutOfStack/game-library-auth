package handlers

import (
	"log"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/data/user"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers routes for auth service
func RegisterRoutes(app *fiber.App, authConf *appconf.Auth, db *sqlx.DB, auth *auth.Auth, log *log.Logger) {

	authAPI := AuthAPI{
		UserRepo: user.NewRepo(db),
		Auth:     auth,
		AuthConf: authConf,
		Log:      log,
	}

	tokenAPI := TokenAPI{
		Auth: auth,
		Log:  log,
	}

	app.Post("/signin", authAPI.signInHandler)
	app.Post("/signup", authAPI.signUpHandler)

	app.Post("/token/verify", tokenAPI.verifyJWT)
}
