package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

var tracer = otel.Tracer("authapi")

// GoogleTokenValidator provides methods for validating Google ID tokens
type GoogleTokenValidator interface {
	Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error)
}

// UserFacade provides methods for working with user facade
type UserFacade interface {
	GoogleOAuth(ctx context.Context, oauthID, email string) (model.User, error)
	DeleteUser(ctx context.Context, userID string) error
	UpdateUserProfile(ctx context.Context, userID string, params model.UpdateProfileParams) (model.User, error)
	VerifyEmail(ctx context.Context, userID string, code string) (model.User, error)
	ResendVerificationEmail(ctx context.Context, userID string) error
	SignIn(ctx context.Context, username, password string) (model.User, error)
	SignUp(ctx context.Context, username, displayName, email, password string, isPublisher bool) (model.User, error)
	CreateTokens(ctx context.Context, user model.User) (facade.TokenPair, error)
	RefreshTokens(ctx context.Context, refreshTokenStr string) (facade.TokenPair, error)
	RevokeRefreshToken(ctx context.Context, refreshTokenStr string) error
	ValidateAccessToken(tokenStr string) bool
	GetClaimsFromAccessToken(tokenStr string) (auth.Claims, error)
}

// APICfg describes configuration for auth api
type APICfg struct {
	RefreshTokenCookieSameSite string
	RefreshTokenCookieSecure   bool
	GoogleOAuthClientID        string
	ContactEmail               string
}

// API describes dependencies for auth endpoints
type API struct {
	log                  *zap.Logger
	googleTokenValidator GoogleTokenValidator
	userFacade           UserFacade
	cfg                  APICfg
}

// NewAPI return new instance of auth api
func NewAPI(log *zap.Logger, googleTokenValidator GoogleTokenValidator, userFacade UserFacade, cfg APICfg) (*API, error) {
	if cfg.GoogleOAuthClientID == "" {
		return nil, errors.New("google client id is empty")
	}

	switch strings.ToLower(cfg.RefreshTokenCookieSameSite) {
	case "lax":
		cfg.RefreshTokenCookieSameSite = fiber.CookieSameSiteLaxMode
	case "strict":
		cfg.RefreshTokenCookieSameSite = fiber.CookieSameSiteStrictMode
	default:
		log.Warn("refresh token cookie same site wasn't strict or lax, defaulted to none", zap.String("sameSite", cfg.RefreshTokenCookieSameSite))
		cfg.RefreshTokenCookieSameSite = fiber.CookieSameSiteNoneMode
	}

	return &API{
		log:                  log,
		googleTokenValidator: googleTokenValidator,
		userFacade:           userFacade,
		cfg:                  cfg,
	}, nil
}
