package authapi

import (
	pb "github.com/OutOfStack/game-library-auth/pkg/proto/authapi/v1"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("grpc-auth-service")

// AuthFacade defines the interface for authentication operations
type AuthFacade interface {
	ValidateAccessToken(tokenStr string) bool
}

// AuthService implements the AuthApiService gRPC service
type AuthService struct {
	pb.UnimplementedAuthApiServiceServer
	log        *zap.Logger
	authFacade AuthFacade
}

// NewAuthService creates a new AuthService instance
func NewAuthService(logger *zap.Logger, authFacade AuthFacade) *AuthService {
	return &AuthService{
		log:        logger,
		authFacade: authFacade,
	}
}
