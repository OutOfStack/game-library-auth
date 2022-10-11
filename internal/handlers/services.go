package handlers

import (
	"fmt"
	"log"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	trace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

var tracer = otel.Tracer(appconf.ServiceName)

// AuthService creates and configures auth app
func AuthService(conf *appconf.Web, authConf *appconf.Auth, zipkinConf *appconf.Zipkin, db *sqlx.DB, log *log.Logger) (*fiber.App, error) {
	err := initTracer(zipkinConf.ReporterURL)
	if err != nil {
		return nil, fmt.Errorf("initializing exporter: %w", err)
	}
	privateKey, err := crypto.ReadPrivateKey(authConf.PrivateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("reading private key file: %w", err)
	}
	a, err := auth.New(authConf.SigningAlgorithm, privateKey)
	if err != nil {
		return nil, fmt.Errorf("initializing token service instance: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      appconf.ServiceName,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	})

	// apply middleware
	app.Use(otelfiber.Middleware(fmt.Sprintf("%s.mw", appconf.ServiceName)))
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: conf.AllowedCORSOrigin,
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "POST, GET, OPTIONS",
	}))

	// register routes
	RegisterRoutes(app, authConf, db, a, log)

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
		return fmt.Errorf("creating new exporter: %w", err)
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
