package handlers

import (
	"context"
	"fmt"

	_ "github.com/OutOfStack/game-library-auth/docs" // swagger docs
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	auth_ "github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/client/mailersend"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	rec "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	swag "github.com/swaggo/http-swagger/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

var tracer = otel.Tracer("api")

// Service creates and configures auth app
func Service(log *zap.Logger, db *sqlx.DB, cfg appconf.Cfg) (*fiber.App, error) {
	ctx := context.Background()

	err := initTracer(cfg.Zipkin.ReporterURL)
	if err != nil {
		return nil, fmt.Errorf("init exporter: %w", err)
	}
	privateKey, err := crypto.ReadPrivateKey(cfg.Auth.PrivateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read private key file: %w", err)
	}
	auth, err := auth_.New(cfg.Auth.SigningAlgorithm, privateKey, cfg.Auth.Issuer)
	if err != nil {
		return nil, fmt.Errorf("create token service instance: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      appconf.ServiceName,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	})

	// apply middleware
	app.Use(rec.New())
	app.Use(otelfiber.Middleware(otelfiber.WithServerName(appconf.ServiceName)))
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.Web.AllowedCORSOrigin,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,DELETE,PATCH,OPTIONS",
	}))

	// create google token validator
	googleTokenValidator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return nil, fmt.Errorf("create google token validator: %w", err)
	}

	// create email sender
	emailSender, err := mailersend.NewClient(mailersend.Config{
		APIToken:  cfg.EmailSender.APIToken,
		FromEmail: cfg.EmailSender.EmailFrom,
		Timeout:   cfg.EmailSender.APITimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("create mailersend client: %w", err)
	}

	// register routes
	authAPI, err := NewAuthAPI(log, &cfg, auth, database.NewUserRepo(db), googleTokenValidator, emailSender)
	if err != nil {
		return nil, fmt.Errorf("create auth api: %w", err)
	}
	checkAPI := NewCheckAPI(db)
	registerRoutes(app, authAPI, checkAPI)

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

func registerRoutes(app *fiber.App, authAPI *AuthAPI, checkAPI *CheckAPI) {
	app.Get("/readiness", checkAPI.Readiness)
	app.Get("/liveness", checkAPI.Liveness)

	app.Post("/signin", authAPI.SignInHandler)
	app.Post("/signup", authAPI.SignUpHandler)
	app.Patch("/account", authAPI.UpdateProfileHandler)
	app.Delete("/account", authAPI.DeleteAccountHandler)
	app.Post("/oauth/google", authAPI.GoogleOAuthHandler)

	app.Post("/verify-email", authAPI.VerifyEmailHandler)
	app.Post("/resend-verification", authAPI.ResendVerificationEmailHandler)
	app.Post("/token/verify", authAPI.VerifyTokenHandler)

	// swagger
	app.Get("/swagger/*", adaptor.HTTPHandler(swag.Handler()))
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
