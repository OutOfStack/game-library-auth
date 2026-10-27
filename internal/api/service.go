package api

import (
	"fmt"
	"strings"

	_ "github.com/OutOfStack/game-library-auth/docs" // swagger docs
	"github.com/OutOfStack/game-library-auth/internal/api/auth"
	"github.com/OutOfStack/game-library-auth/internal/api/unsubscribe"
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/middleware"
	"github.com/OutOfStack/game-library-auth/pkg/observability"
	fiberotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/contrib/v3/swaggo"
	fiberzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	rec "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/template/html/v3"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// Service creates and configures auth app
func Service(
	log *zap.Logger,
	authAPI *auth.API,
	unsubscribeAPI *unsubscribe.API,
	cfg *appconf.API,
) (*fiber.App, *trace.TracerProvider, error) {
	tp, err := observability.InitTracer(log, cfg.Jaeger.OTLPEndpoint, appconf.ServiceName)
	if err != nil {
		return nil, nil, fmt.Errorf("init exporter: %w", err)
	}

	// initialize HTML template engine
	viewEngine := html.New("./internal/web/templates", ".html")

	app := fiber.New(fiber.Config{
		AppName:      appconf.ServiceName,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		Views:        viewEngine,
	})

	// apply middleware
	app.Use(middleware.Metrics())
	app.Use(rec.New())
	app.Use(fiberotel.New())
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(cfg.Web.AllowedCORSOrigin, ","),
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowCredentials: true,
	}))

	registerRoutes(app, authAPI, unsubscribeAPI)

	return app, tp, nil
}

func registerRoutes(app *fiber.App, authAPI *auth.API, unsubscribeAPI *unsubscribe.API) {
	// user
	app.Post("/signin", authAPI.SignInHandler)
	app.Post("/signup", authAPI.SignUpHandler)
	app.Patch("/account", authAPI.UpdateProfileHandler)
	app.Delete("/account", authAPI.DeleteAccountHandler)
	app.Post("/oauth/google", authAPI.GoogleOAuthHandler)
	app.Post("/oauth/github", authAPI.GitHubOAuthHandler)
	app.Post("/logout", authAPI.LogoutHandler)

	// email verification
	app.Post("/verify-email", authAPI.VerifyEmailHandler)
	app.Post("/resend-verification", authAPI.ResendVerificationEmailHandler)

	// unsubscribe
	app.Get("/unsubscribe", unsubscribeAPI.UnsubscribeHandler)
	app.Post("/unsubscribe", unsubscribeAPI.UnsubscribeConfirmHandler)

	// token
	app.Post("/refresh", authAPI.RefreshTokenHandler)

	// swagger
	app.Get("/swagger/*", swaggo.HandlerDefault)
}
