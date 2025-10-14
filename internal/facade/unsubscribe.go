package facade

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"go.uber.org/zap"
)

// UnsubscribeEmail unsubscribes an email address from all notifications
func (p *Provider) UnsubscribeEmail(ctx context.Context, token string) (string, error) {
	// validate token and extract email
	email, err := p.unsubscribeTokenGenerator.ValidateToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid unsubscribe token: %w", err)
	}

	// get the email verification record to check token validity and expiry
	verification, err := p.userRepo.GetEmailVerificationByUnsubscribeToken(ctx, token)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return "", database.ErrNotFound
		}
		return "", fmt.Errorf("get email verification by unsubscribe token: %w", err)
	}

	// check if the unsubscribe token has expired
	if time.Now().After(verification.DateCreated.Add(model.UnsubscribeTokenTTL)) {
		return "", errors.New("unsubscribe link has expired")
	}

	// create unsubscribe record
	unsubscribe := database.NewEmailUnsubscribe(email)
	if err = p.userRepo.CreateEmailUnsubscribe(ctx, unsubscribe); err != nil {
		return "", fmt.Errorf("create email unsubscribe: %w", err)
	}

	p.log.Info("email unsubscribed", zap.String("email", email))
	return email, nil
}
