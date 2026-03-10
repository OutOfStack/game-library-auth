package auth

import (
	"context"
	"strings"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/client/githubapi"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

var tracer = otel.Tracer("authapi")

// GoogleIDTokenClient provides methods for validating Google ID tokens
type GoogleIDTokenClient interface {
	ValidateIDToken(ctx context.Context, idToken string) (*idtoken.Payload, error)
}

// GitHubOAuthClient provides methods for GitHub OAuth authentication
type GitHubOAuthClient interface {
	ExchangeCodeForUser(ctx context.Context, code string) (githubapi.UserInfo, error)
}

// UserFacade provides methods for working with user facade
type UserFacade interface {
	GoogleOAuth(ctx context.Context, oauthID, email string) (model.User, error)
	GitHubOAuth(ctx context.Context, oauthID, email, username string) (model.User, error)
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
	ContactEmail               string
}

// API describes dependencies for auth endpoints
type API struct {
	log                 *zap.Logger
	googleIDTokenClient GoogleIDTokenClient
	githubOAuthClient   GitHubOAuthClient
	userFacade          UserFacade
	cfg                 APICfg
}

// NewAPI return new instance of auth api
func NewAPI(log *zap.Logger, googleIDTokenClient GoogleIDTokenClient, githubOAuthClient GitHubOAuthClient, userFacade UserFacade, cfg APICfg) (*API, error) {
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
		log:                 log,
		googleIDTokenClient: googleIDTokenClient,
		githubOAuthClient:   githubOAuthClient,
		userFacade:          userFacade,
		cfg:                 cfg,
	}, nil
}
