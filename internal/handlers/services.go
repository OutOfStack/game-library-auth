package handlers

import (
	"fmt"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	rec "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.uber.org/zap"
)

var tracer = otel.Tracer(appconf.ServiceName)

// AuthService creates and configures auth app
func AuthService(log *zap.Logger, db *sqlx.DB, cfg appconf.Cfg) (*fiber.App, error) {
	err := initTracer(cfg.Zipkin.ReporterURL)
	if err != nil {
		return nil, fmt.Errorf("initializing exporter: %w", err)
	}
	privateKey, err := crypto.ReadPrivateKey(cfg.Auth.PrivateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("reading private key file: %w", err)
	}
	a, err := auth.New(cfg.Auth.SigningAlgorithm, privateKey)
	if err != nil {
		return nil, fmt.Errorf("initializing token service instance: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      appconf.ServiceName,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	})

	// apply middleware
	app.Use(rec.New())
	app.Use(otelfiber.Middleware(otelfiber.WithServerName(appconf.ServiceName + ".mw")))
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.Web.AllowedCORSOrigin,
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "POST, GET, OPTIONS",
	}))

	// register routes
	RegisterRoutes(log, app, cfg, db, a)

	return app, nil
}

// DebugService creates and configures debug app
func DebugService() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "debug",
	})

	// apply middleware
	app.Use(pprof.New())

	return app
}

func initTracer(reporterURL string) error {
	exporter, err := zipkin.New(reporterURL)
	if err != nil {
		return fmt.Errorf("can't create new exporter: %w", err)
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

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tp)

	return nil
}
