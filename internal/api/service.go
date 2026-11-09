package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/OutOfStack/game-library-auth/api"
	"github.com/OutOfStack/game-library-auth/internal/api/auth"
	"github.com/OutOfStack/game-library-auth/internal/api/openapi"
	"github.com/OutOfStack/game-library-auth/internal/api/unsubscribe"
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/middleware"
	"github.com/OutOfStack/game-library-auth/internal/web"
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
	viewEngine := html.NewFileSystem(http.FS(web.Templates()), ".html")

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

// aliases give embedded APIs distinct field names
type (
	authHandlers        = auth.API
	unsubscribeHandlers = unsubscribe.API
)

// server implements operations generated from OpenAPI spec
type server struct {
	*authHandlers
	*unsubscribeHandlers
}

var _ openapi.ServerInterface = server{}

func registerRoutes(app *fiber.App, authAPI *auth.API, unsubscribeAPI *unsubscribe.API) {
	openapi.RegisterHandlers(app, server{authHandlers: authAPI, unsubscribeHandlers: unsubscribeAPI})

	// swagger
	app.Get("/swagger/openapi.yaml", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/yaml")
		return c.Send(api.OpenAPISpec)
	})
	app.Get("/swagger/*", swaggo.New(swaggo.Config{URL: "/swagger/openapi.yaml"}))
}
