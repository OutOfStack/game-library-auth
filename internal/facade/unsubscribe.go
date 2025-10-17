package facade

import (
	"context"
	"fmt"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"go.uber.org/zap"
)

// UnsubscribeEmail unsubscribes an email address from all notifications
func (p *Provider) UnsubscribeEmail(ctx context.Context, token string) (string, error) {
	// validate token and extract email
	email, err := p.unsubscribeTokenGenerator.ValidateToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid unsubscribe token: %w", err)
	}

	// create unsubscribe record
	unsubscribe := database.NewEmailUnsubscribe(email)
	if err = p.userRepo.CreateEmailUnsubscribe(ctx, unsubscribe); err != nil {
		return "", fmt.Errorf("create email unsubscribe: %w", err)
	}

	p.log.Info("email unsubscribed", zap.String("email", email))
	return email, nil
}

// IsEmailUnsubscribed checks if an email address is unsubscribed from all notifications
func (p *Provider) IsEmailUnsubscribed(ctx context.Context, email string) (bool, error) {
	return p.userRepo.IsEmailUnsubscribed(ctx, email)
}
