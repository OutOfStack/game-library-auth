package handlers

import (
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/data"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// RegisterRoutes registers routes for auth service
func RegisterRoutes(log *zap.Logger, app *fiber.App, cfg appconf.Cfg, db *sqlx.DB, auth *auth.Auth) {
	authAPI := AuthAPI{
		DB:   data.NewRepo(db),
		Auth: auth,
		Cfg:  cfg,
		Log:  log,
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

	app.Post("/signin", authAPI.SignInHandler)
	app.Post("/signup", authAPI.SignUpHandler)
	app.Post("/update_profile", authAPI.UpdateProfileHandler)

	app.Post("/token/verify", tokenAPI.VerifyToken)
}
