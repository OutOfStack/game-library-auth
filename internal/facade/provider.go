package facade

import (
	"context"

	"github.com/OutOfStack/game-library-auth/internal/client/mailersend"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"go.uber.org/zap"
)

// Provider represents dependencies for facade layer
type Provider struct {
	log                *zap.Logger
	userRepo           UserRepo
	emailSender        EmailSender
	disableEmailSender bool
}

// New creates a new facade provider
func New(log *zap.Logger, userRepo UserRepo, emailSender EmailSender, disableEmailSender bool) *Provider {
	return &Provider{
		log:                log,
		userRepo:           userRepo,
		emailSender:        emailSender,
		disableEmailSender: disableEmailSender,
	}
}

// UserRepo provides methods for working with user repo
type UserRepo interface {
	RunWithTx(ctx context.Context, f func(context.Context) error) error

	CreateUser(ctx context.Context, user database.User) error
	UpdateUser(ctx context.Context, user database.User) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (database.User, error)
	GetUserByUsername(ctx context.Context, username string) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	GetUserByOAuth(ctx context.Context, provider string, oauthID string) (database.User, error)
	CheckUserExists(ctx context.Context, name string, role database.Role) (bool, error)
	SetUserEmailVerified(ctx context.Context, userID string) error

	CreateEmailVerification(ctx context.Context, verification database.EmailVerification) error
	GetEmailVerificationByUserID(ctx context.Context, userID string) (database.EmailVerification, error)
	SetEmailVerificationMessageID(ctx context.Context, verificationID string, messageID string) error
	SetEmailVerificationUsed(ctx context.Context, id string, verified bool) error
}

// EmailSender provides methods for sending emails
type EmailSender interface {
	SendEmailVerification(ctx context.Context, req mailersend.SendEmailVerificationRequest) (string, error)
}
