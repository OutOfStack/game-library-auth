package handlers

import (
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/data/user"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// RegisterRoutes registers routes for auth service
func RegisterRoutes(log *zap.Logger, app *fiber.App, authConf *appconf.Auth, db *sqlx.DB, auth *auth.Auth) {

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

	checkAPI := CheckAPI{
		DB: db,
	}

	app.Get("/readiness", checkAPI.Readiness)
	app.Get("/liveness", checkAPI.Liveness)

	app.Post("/signin", authAPI.signInHandler)
	app.Post("/signup", authAPI.signUpHandler)

	app.Post("/token/verify", tokenAPI.verifyToken)
}
