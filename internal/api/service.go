package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	_ "github.com/OutOfStack/game-library-auth/docs" // swagger docs
	"github.com/OutOfStack/game-library-auth/internal/api/auth"
	"github.com/OutOfStack/game-library-auth/internal/api/tools"
	"github.com/OutOfStack/game-library-auth/internal/api/unsubscribe"
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/middleware"
	fiberotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/contrib/v3/swaggo"
	fiberzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	rec "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/template/html/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.uber.org/zap"
)

// Service creates and configures auth app
func Service(
	log *zap.Logger,
	authAPI *auth.API,
	checkAPI *tools.HealthCheckAPI,
	unsubscribeAPI *unsubscribe.API,
	cfg *appconf.Cfg,
) (*fiber.App, *trace.TracerProvider, error) {
	tp, err := initTracer(log, cfg.Jaeger.OTLPEndpoint)
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
	app.Use(fiberotel.Middleware())
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(cfg.Web.AllowedCORSOrigin, ","),
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowCredentials: true,
	}))

	registerRoutes(app, authAPI, checkAPI, unsubscribeAPI)

	return app, tp, nil
}

// DebugService creates and configures debug app
func DebugService() *fiber.App {
	app := fiber.New(fiber.Config{AppName: "debug"})

	// apply middleware
	app.Use(pprof.New())

	return app
}

func registerRoutes(app *fiber.App, authAPI *auth.API, checkAPI *tools.HealthCheckAPI, unsubscribeAPI *unsubscribe.API) {
	// metrics
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// health
	app.Get("/readiness", checkAPI.Readiness)
	app.Get("/liveness", checkAPI.Liveness)

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

func initTracer(log *zap.Logger, otlpEndpoint string) (*trace.TracerProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create new exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(appconf.ServiceName),
			)),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Error("otel error", zap.Error(err))
	}))
	otel.SetTracerProvider(tp)

	return tp, nil
}
